package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sql2api/internal/config"
	"sql2api/internal/repository"

	"gorm.io/gorm"
)

// QueryResult 查询结果
type QueryResult struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int64                    `json:"total"`
}

// ExecuteResult 执行结果
type ExecuteResult struct {
	AffectedRows int64 `json:"affected_rows"`
	LastInsertID int64 `json:"last_insert_id,omitempty"`
}

// BatchQuery 批量查询项
type BatchQuery struct {
	SQL    string                 `json:"sql"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// BatchResult 批量执行结果
type BatchResult struct {
	Results            []ExecuteResult `json:"results"`
	TotalAffectedRows  int64           `json:"total_affected_rows"`
	Success            bool            `json:"success"`
}

// SQLEngine SQL 查询引擎
type SQLEngine struct {
	db           *gorm.DB
	dbType       string // "postgres" 或 "oracle"
	config       *config.SQLConfig
	security     *SecurityValidator
	validator    *QueryValidator
	dialect      DatabaseDialect
	errorMapper  *DatabaseErrorMapper
	monitor      *PerformanceMonitor
	memOptimizer *MemoryOptimizer
}

// NewSQLEngine 创建 SQL 查询引擎
func NewSQLEngine(repos *repository.Repositories, cfg *config.SQLConfig) (*SQLEngine, error) {
	if repos == nil || repos.GetDB() == nil {
		return nil, errors.New("database connection is required")
	}

	if cfg == nil {
		return nil, errors.New("SQL configuration is required")
	}

	if !cfg.Enabled {
		return nil, errors.New("SQL functionality is disabled")
	}

	// 获取数据库类型
	dbType := "postgres" // 默认为 PostgreSQL
	if db := repos.GetDB(); db != nil {
		if dialector := db.Dialector.Name(); dialector == "oracle" {
			dbType = "oracle"
		}
	}

	// 创建安全验证器
	security := NewSecurityValidator(cfg)

	// 创建查询验证器
	validator := NewQueryValidator()

	// 创建数据库方言
	dialectFactory := NewDialectFactory()
	dialect := dialectFactory.CreateDialect(dbType)

	// 创建错误映射器
	errorMapper := NewDatabaseErrorMapper(dbType)

	// 创建性能监控器
	monitor := NewPerformanceMonitor(
		true,                    // enabled
		true,                    // logQueries
		true,                    // logErrors
		int64(cfg.MaxQueryTime*1000/2), // slowQueryMs (一半的超时时间作为慢查询阈值)
	)

	// 创建内存优化器
	memOptimizer := NewMemoryOptimizer(
		cfg.MaxResultSize, // maxResultSize
		100,               // batchSize
	)

	return &SQLEngine{
		db:           repos.GetDB(),
		dbType:       dbType,
		config:       cfg,
		security:     security,
		validator:    validator,
		dialect:      dialect,
		errorMapper:  errorMapper,
		monitor:      monitor,
		memOptimizer: memOptimizer,
	}, nil
}

// ExecuteQuery 执行查询操作（SELECT）
func (e *SQLEngine) ExecuteQuery(ctx context.Context, query string, params map[string]interface{}) (*QueryResult, error) {
	// 开始监控
	queryCtx := e.monitor.StartQuery(ctx, "select", e.dbType, query)
	defer func() {
		// 这里会在函数返回时调用，需要在后面设置结果
	}()

	// 查询结构验证
	if err := e.validator.ValidateQueryStructure(query); err != nil {
		queryCtx.Finish(false, 0, 0, err)
		return nil, fmt.Errorf("query structure validation failed: %w", err)
	}

	// 安全验证
	if err := e.security.ValidateQuery(query, params); err != nil {
		queryCtx.Finish(false, 0, 0, err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// 检查是否为查询操作
	if !e.security.IsSelectQuery(query) {
		err := errors.New("only SELECT queries are allowed in ExecuteQuery")
		queryCtx.Finish(false, 0, 0, err)
		return nil, err
	}

	// 创建带超时的上下文
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(e.config.MaxQueryTime)*time.Second)
	defer cancel()

	// 执行查询
	rows, err := e.executeRawQuery(execCtx, query, params)
	if err != nil {
		mappedErr := e.errorMapper.MapError(err)
		queryCtx.Finish(false, 0, 0, err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// 解析结果
	result, err := e.parseQueryResult(rows)
	if err != nil {
		queryCtx.Finish(false, 0, 0, err)
		return nil, fmt.Errorf("failed to parse query result: %w", err)
	}

	// 内存优化
	result.Rows = e.memOptimizer.OptimizeResultSet(result.Rows)
	result.Total = int64(len(result.Rows))

	// 检查结果集大小限制
	if len(result.Rows) > e.config.MaxResultSize {
		err := fmt.Errorf("result set too large: %d rows (max: %d)", len(result.Rows), e.config.MaxResultSize)
		queryCtx.Finish(false, 0, int64(len(result.Rows)), err)
		return nil, err
	}

	// 记录成功执行
	queryCtx.Finish(true, 0, result.Total, nil)
	return result, nil
}

// ExecuteSQL 执行任意 SQL 操作（INSERT、UPDATE、DELETE）
func (e *SQLEngine) ExecuteSQL(ctx context.Context, query string, params map[string]interface{}) (*ExecuteResult, error) {
	// 查询结构验证
	if err := e.validator.ValidateQueryStructure(query); err != nil {
		return nil, fmt.Errorf("query structure validation failed: %w", err)
	}

	// 安全验证
	if err := e.security.ValidateQuery(query, params); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// 检查是否允许原生 SQL
	if !e.config.EnableRawSQL {
		return nil, errors.New("raw SQL execution is disabled")
	}

	// 创建带超时的上下文
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(e.config.MaxQueryTime)*time.Second)
	defer cancel()

	// 执行 SQL
	result := e.db.WithContext(queryCtx).Exec(query, e.convertParams(params)...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to execute SQL: %w", result.Error)
	}

	return &ExecuteResult{
		AffectedRows: result.RowsAffected,
		LastInsertID: 0, // GORM 不直接支持 LastInsertID
	}, nil
}

// ExecuteBatch 执行批量 SQL 操作
func (e *SQLEngine) ExecuteBatch(ctx context.Context, queries []BatchQuery, transactional bool) (*BatchResult, error) {
	if !e.config.EnableBatch {
		return nil, errors.New("batch operations are disabled")
	}

	if len(queries) == 0 {
		return nil, errors.New("no queries provided")
	}

	// 验证所有查询
	for i, query := range queries {
		if err := e.security.ValidateQuery(query.SQL, query.Params); err != nil {
			return nil, fmt.Errorf("security validation failed for query %d: %w", i, err)
		}
	}

	// 创建带超时的上下文
	batchCtx, cancel := context.WithTimeout(ctx, time.Duration(e.config.MaxQueryTime)*time.Second)
	defer cancel()

	if transactional && e.config.EnableTransactions {
		return e.executeBatchWithTransaction(batchCtx, queries)
	}

	return e.executeBatchWithoutTransaction(batchCtx, queries)
}

// executeRawQuery 执行原生查询
func (e *SQLEngine) executeRawQuery(ctx context.Context, query string, params map[string]interface{}) (*sql.Rows, error) {
	// 获取底层的 sql.DB
	sqlDB, err := e.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 转换参数
	args := e.convertParams(params)

	// 执行查询
	return sqlDB.QueryContext(ctx, query, args...)
}

// convertParams 转换参数映射为参数数组
func (e *SQLEngine) convertParams(params map[string]interface{}) []interface{} {
	if len(params) == 0 {
		return nil
	}

	// 简单实现：按参数名排序后转换为数组
	// 实际应用中需要更复杂的参数替换逻辑
	var args []interface{}
	for _, value := range params {
		args = append(args, value)
	}
	return args
}

// parseQueryResult 解析查询结果
func (e *SQLEngine) parseQueryResult(rows *sql.Rows) (*QueryResult, error) {
	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &QueryResult{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	// 创建扫描目标
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 扫描所有行
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 构建行数据
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	result.Total = int64(len(result.Rows))
	return result, nil
}

// executeBatchWithTransaction 在事务中执行批量操作
func (e *SQLEngine) executeBatchWithTransaction(ctx context.Context, queries []BatchQuery) (*BatchResult, error) {
	result := &BatchResult{
		Results: make([]ExecuteResult, 0, len(queries)),
	}

	// 开始事务
	tx := e.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 确保事务会被处理
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 执行所有查询
	for i, query := range queries {
		execResult := tx.Exec(query.SQL, e.convertParams(query.Params)...)
		if execResult.Error != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to execute query %d: %w", i, execResult.Error)
		}

		result.Results = append(result.Results, ExecuteResult{
			AffectedRows: execResult.RowsAffected,
			LastInsertID: 0,
		})
		result.TotalAffectedRows += execResult.RowsAffected
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result.Success = true
	return result, nil
}

// executeBatchWithoutTransaction 不使用事务执行批量操作
func (e *SQLEngine) executeBatchWithoutTransaction(ctx context.Context, queries []BatchQuery) (*BatchResult, error) {
	result := &BatchResult{
		Results: make([]ExecuteResult, 0, len(queries)),
	}

	// 逐个执行查询
	for i, query := range queries {
		execResult := e.db.WithContext(ctx).Exec(query.SQL, e.convertParams(query.Params)...)
		if execResult.Error != nil {
			return nil, fmt.Errorf("failed to execute query %d: %w", i, execResult.Error)
		}

		result.Results = append(result.Results, ExecuteResult{
			AffectedRows: execResult.RowsAffected,
			LastInsertID: 0,
		})
		result.TotalAffectedRows += execResult.RowsAffected
	}

	result.Success = true
	return result, nil
}

// GetDatabaseType 获取数据库类型
func (e *SQLEngine) GetDatabaseType() string {
	return e.dbType
}

// IsEnabled 检查 SQL 功能是否启用
func (e *SQLEngine) IsEnabled() bool {
	return e.config.Enabled
}

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sql2api/internal/config"
	"sql2api/internal/model"
	"sql2api/internal/repository"
	"sql2api/internal/sql"
)

// QueryBuilder 查询构建器类型别名
type QueryBuilder = sql.QueryBuilder

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(dbType string) *QueryBuilder {
	return sql.NewQueryBuilder(dbType)
}

// SQLService SQL 业务服务接口
type SQLService interface {
	// 执行查询操作
	ExecuteQuery(ctx context.Context, req *model.SQLRequest) (*model.SQLResponse, error)
	
	// 执行 SQL 操作（INSERT、UPDATE、DELETE）
	ExecuteSQL(ctx context.Context, req *model.SQLRequest) (*model.SQLResponse, error)
	
	// 执行批量 SQL 操作
	ExecuteBatch(ctx context.Context, req *model.BatchSQLRequest) (*model.BatchSQLResponse, error)
	
	// 执行便捷插入操作
	ExecuteInsert(ctx context.Context, req *model.InsertRequest) (*model.SQLResponse, error)
	
	// 执行批量插入操作
	ExecuteBatchInsert(ctx context.Context, req *model.BatchInsertRequest) (*model.SQLResponse, error)
	
	// 健康检查
	HealthCheck() error
}

// sqlService SQL 业务服务实现
type sqlService struct {
	sqlEngine *sql.SQLEngine
	config    *config.SQLConfig
	builder   *QueryBuilder
}

// NewSQLService 创建 SQL 业务服务
func NewSQLService(repos *repository.Repositories, cfg *config.SQLConfig) (SQLService, error) {
	if repos == nil {
		return nil, errors.New("repositories cannot be nil")
	}
	
	if cfg == nil {
		return nil, errors.New("SQL configuration cannot be nil")
	}
	
	if !cfg.Enabled {
		return nil, errors.New("SQL service is disabled")
	}
	
	// 创建 SQL 查询引擎
	engine, err := sql.NewSQLEngine(repos, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL engine: %w", err)
	}
	
	// 创建查询构建器
	builder := NewQueryBuilder(engine.GetDatabaseType())
	
	return &sqlService{
		sqlEngine: engine,
		config:    cfg,
		builder:   builder,
	}, nil
}

// ExecuteQuery 执行查询操作
func (s *sqlService) ExecuteQuery(ctx context.Context, req *model.SQLRequest) (*model.SQLResponse, error) {
	startTime := time.Now()
	
	// 验证请求
	if err := s.validateSQLRequest(req); err != nil {
		return s.createErrorResponse(model.SQLErrorParams, "Request validation failed", err.Error()), nil
	}
	
	// 构建查询
	query, params, err := s.buildQuery(req)
	if err != nil {
		return s.createErrorResponse(model.SQLErrorSyntax, "Query building failed", err.Error()), nil
	}
	
	// 应用分页和排序
	query = s.applyPaginationAndSort(query, req)
	
	// 执行查询
	result, err := s.sqlEngine.ExecuteQuery(ctx, query, params)
	if err != nil {
		return s.handleExecutionError(err), nil
	}
	
	// 构建响应
	response := s.buildQueryResponse(result, req)
	response.ExecutionTime = float64(time.Since(startTime).Nanoseconds()) / 1e6 // 转换为毫秒
	
	return response, nil
}

// ExecuteSQL 执行 SQL 操作
func (s *sqlService) ExecuteSQL(ctx context.Context, req *model.SQLRequest) (*model.SQLResponse, error) {
	startTime := time.Now()
	
	// 验证请求
	if err := s.validateSQLRequest(req); err != nil {
		return s.createErrorResponse(model.SQLErrorParams, "Request validation failed", err.Error()), nil
	}
	
	// 构建查询
	query, params, err := s.buildQuery(req)
	if err != nil {
		return s.createErrorResponse(model.SQLErrorSyntax, "Query building failed", err.Error()), nil
	}
	
	// 执行 SQL
	result, err := s.sqlEngine.ExecuteSQL(ctx, query, params)
	if err != nil {
		return s.handleExecutionError(err), nil
	}
	
	// 构建响应
	response := model.NewSQLSuccessResponse(nil, result.AffectedRows, "SQL executed successfully")
	response.ExecutionTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	
	return &response, nil
}

// ExecuteBatch 执行批量 SQL 操作
func (s *sqlService) ExecuteBatch(ctx context.Context, req *model.BatchSQLRequest) (*model.BatchSQLResponse, error) {
	startTime := time.Now()
	
	// 验证批量请求
	if err := s.validateBatchRequest(req); err != nil {
		return s.createBatchErrorResponse(model.SQLErrorParams, "Batch request validation failed", err.Error()), nil
	}
	
	// 构建批量查询
	var batchQueries []sql.BatchQuery
	for _, sqlReq := range req.Operations {
		query, params, err := s.buildQuery(&sqlReq)
		if err != nil {
			return s.createBatchErrorResponse(model.SQLErrorSyntax, "Query building failed", err.Error()), nil
		}
		
		batchQueries = append(batchQueries, sql.BatchQuery{
			SQL:    query,
			Params: params,
		})
	}
	
	// 执行批量操作
	result, err := s.sqlEngine.ExecuteBatch(ctx, batchQueries, req.Transactional)
	if err != nil {
		return s.handleBatchExecutionError(err), nil
	}
	
	// 构建响应
	response := s.buildBatchResponse(result, req)
	response.ExecutionTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	
	return response, nil
}

// ExecuteInsert 执行便捷插入操作
func (s *sqlService) ExecuteInsert(ctx context.Context, req *model.InsertRequest) (*model.SQLResponse, error) {
	startTime := time.Now()
	
	// 验证插入请求
	if err := s.validateInsertRequest(req); err != nil {
		return s.createErrorResponse(model.SQLErrorParams, "Insert request validation failed", err.Error()), nil
	}
	
	// 构建插入查询
	query, params, err := s.builder.BuildInsertQuery(req)
	if err != nil {
		return s.createErrorResponse(model.SQLErrorSyntax, "Insert query building failed", err.Error()), nil
	}
	
	// 执行插入
	result, err := s.sqlEngine.ExecuteSQL(ctx, query, params)
	if err != nil {
		return s.handleExecutionError(err), nil
	}
	
	// 构建响应
	response := model.NewSQLSuccessResponse(nil, result.AffectedRows, "Insert executed successfully")
	response.ExecutionTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	
	return &response, nil
}

// ExecuteBatchInsert 执行批量插入操作
func (s *sqlService) ExecuteBatchInsert(ctx context.Context, req *model.BatchInsertRequest) (*model.SQLResponse, error) {
	startTime := time.Now()
	
	// 验证批量插入请求
	if err := s.validateBatchInsertRequest(req); err != nil {
		return s.createErrorResponse(model.SQLErrorParams, "Batch insert request validation failed", err.Error()), nil
	}
	
	// 构建批量插入查询
	query, params, err := s.builder.BuildBatchInsertQuery(req)
	if err != nil {
		return s.createErrorResponse(model.SQLErrorSyntax, "Batch insert query building failed", err.Error()), nil
	}
	
	// 执行批量插入
	result, err := s.sqlEngine.ExecuteSQL(ctx, query, params)
	if err != nil {
		return s.handleExecutionError(err), nil
	}
	
	// 构建响应
	response := model.NewSQLSuccessResponse(nil, result.AffectedRows, "Batch insert executed successfully")
	response.ExecutionTime = float64(time.Since(startTime).Nanoseconds()) / 1e6
	
	return &response, nil
}

// HealthCheck 健康检查
func (s *sqlService) HealthCheck() error {
	if s.sqlEngine == nil {
		return errors.New("SQL engine is not initialized")
	}
	
	if !s.sqlEngine.IsEnabled() {
		return errors.New("SQL engine is disabled")
	}
	
	return nil
}

// ===== 辅助方法 =====

// validateSQLRequest 验证 SQL 请求
func (s *sqlService) validateSQLRequest(req *model.SQLRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	// 验证数据库类型
	if !model.ValidateDatabaseType(req.DatabaseType) {
		return fmt.Errorf("unsupported database type: %s", req.DatabaseType)
	}

	// 验证查询内容
	if req.SQL == "" && req.Query == nil {
		return errors.New("either SQL or Query must be provided")
	}

	if req.SQL != "" && req.Query != nil {
		return errors.New("cannot provide both SQL and Query")
	}

	// 验证结构化查询
	if req.Query != nil {
		if err := s.validateStructuredQuery(req.Query); err != nil {
			return fmt.Errorf("structured query validation failed: %w", err)
		}
	}

	return nil
}

// validateStructuredQuery 验证结构化查询
func (s *sqlService) validateStructuredQuery(query *model.StructuredQuery) error {
	if query.Table == "" {
		return errors.New("table name is required")
	}

	if !model.ValidateSQLAction(query.Action) {
		return fmt.Errorf("invalid action: %s", query.Action)
	}

	// 根据操作类型验证必要字段
	switch query.Action {
	case "select":
		// SELECT 查询不需要额外验证
	case "insert":
		if len(query.Data) == 0 {
			return errors.New("data is required for insert operation")
		}
	case "update":
		if len(query.Data) == 0 {
			return errors.New("data is required for update operation")
		}
	case "delete":
		if len(query.Where) == 0 {
			return errors.New("where condition is required for delete operation")
		}
	}

	return nil
}

// validateBatchRequest 验证批量请求
func (s *sqlService) validateBatchRequest(req *model.BatchSQLRequest) error {
	if req == nil {
		return errors.New("batch request cannot be nil")
	}

	if len(req.Operations) == 0 {
		return errors.New("no operations provided")
	}

	if len(req.Operations) > 100 {
		return errors.New("too many operations (max: 100)")
	}

	// 验证数据库类型
	if !model.ValidateDatabaseType(req.DatabaseType) {
		return fmt.Errorf("unsupported database type: %s", req.DatabaseType)
	}

	// 验证每个操作
	for i, op := range req.Operations {
		if err := s.validateSQLRequest(&op); err != nil {
			return fmt.Errorf("operation %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateInsertRequest 验证插入请求
func (s *sqlService) validateInsertRequest(req *model.InsertRequest) error {
	if req == nil {
		return errors.New("insert request cannot be nil")
	}

	if req.Table == "" {
		return errors.New("table name is required")
	}

	if len(req.Data) == 0 {
		return errors.New("data is required")
	}

	// 验证数据库类型
	if !model.ValidateDatabaseType(req.DatabaseType) {
		return fmt.Errorf("unsupported database type: %s", req.DatabaseType)
	}

	// 验证冲突处理
	if !model.ValidateOnConflictAction(req.OnConflict) {
		return fmt.Errorf("invalid on_conflict action: %s", req.OnConflict)
	}

	return nil
}

// validateBatchInsertRequest 验证批量插入请求
func (s *sqlService) validateBatchInsertRequest(req *model.BatchInsertRequest) error {
	if req == nil {
		return errors.New("batch insert request cannot be nil")
	}

	if req.Table == "" {
		return errors.New("table name is required")
	}

	if len(req.Data) == 0 {
		return errors.New("data is required")
	}

	if len(req.Data) > 1000 {
		return errors.New("too many records (max: 1000)")
	}

	// 验证数据库类型
	if !model.ValidateDatabaseType(req.DatabaseType) {
		return fmt.Errorf("unsupported database type: %s", req.DatabaseType)
	}

	// 验证冲突处理
	if !model.ValidateOnConflictAction(req.OnConflict) {
		return fmt.Errorf("invalid on_conflict action: %s", req.OnConflict)
	}

	return nil
}

// buildQuery 构建查询
func (s *sqlService) buildQuery(req *model.SQLRequest) (string, map[string]interface{}, error) {
	if req.SQL != "" {
		// 使用原生 SQL
		return req.SQL, req.Params, nil
	}

	if req.Query != nil {
		// 使用结构化查询
		return s.builder.BuildStructuredQuery(req.Query)
	}

	return "", nil, errors.New("no query provided")
}

// applyPaginationAndSort 应用分页和排序
func (s *sqlService) applyPaginationAndSort(query string, req *model.SQLRequest) string {
	// 应用排序
	if req.Sort != nil && req.Sort.SortBy != "" {
		query = s.builder.ApplySort(query, req.Sort.SortBy, req.Sort.SortOrder)
	}

	// 应用分页
	if req.Pagination != nil && req.Pagination.PageSize > 0 {
		offset := 0
		if req.Pagination.Page > 1 {
			offset = (req.Pagination.Page - 1) * req.Pagination.PageSize
		}
		query = s.builder.ApplyPagination(query, offset, req.Pagination.PageSize)
	}

	return query
}

// buildQueryResponse 构建查询响应
func (s *sqlService) buildQueryResponse(result *sql.QueryResult, req *model.SQLRequest) *model.SQLResponse {
	response := model.NewSQLSuccessResponse(result.Rows, 0, "Query executed successfully")
	response.Columns = result.Columns
	response.Total = result.Total

	// 设置分页信息
	if req.Pagination != nil {
		response.Page = req.Pagination.Page
		response.PageSize = req.Pagination.PageSize
	}

	return &response
}

// buildBatchResponse 构建批量响应
func (s *sqlService) buildBatchResponse(result *sql.BatchResult, req *model.BatchSQLRequest) *model.BatchSQLResponse {
	var results []model.SQLOperationResult

	for i, sqlResult := range result.Results {
		opResult := model.SQLOperationResult{
			Index:        i,
			Success:      true,
			AffectedRows: sqlResult.AffectedRows,
		}
		results = append(results, opResult)
	}

	return &model.BatchSQLResponse{
		Success:           result.Success,
		Message:           "Batch executed successfully",
		Results:           results,
		Timestamp:         time.Now(),
		TotalAffectedRows: result.TotalAffectedRows,
		ExecutedCount:     len(results),
		FailedCount:       0,
	}
}

// createErrorResponse 创建错误响应
func (s *sqlService) createErrorResponse(code int, message, details string) *model.SQLResponse {
	response := model.NewSQLErrorResponse(code, message, details)
	return &response
}

// createBatchErrorResponse 创建批量错误响应
func (s *sqlService) createBatchErrorResponse(code int, message, details string) *model.BatchSQLResponse {
	response := model.NewBatchSQLErrorResponse(code, message, details)
	return &response
}

// handleExecutionError 处理执行错误
func (s *sqlService) handleExecutionError(err error) *model.SQLResponse {
	// 根据错误类型返回相应的错误码
	errMsg := err.Error()

	if contains(errMsg, "timeout") || contains(errMsg, "context deadline exceeded") {
		return s.createErrorResponse(model.SQLErrorTimeout, "Query timeout", errMsg)
	}

	if contains(errMsg, "connection") || contains(errMsg, "connect") {
		return s.createErrorResponse(model.SQLErrorConnection, "Database connection error", errMsg)
	}

	if contains(errMsg, "permission") || contains(errMsg, "access denied") {
		return s.createErrorResponse(model.SQLErrorPermission, "Permission denied", errMsg)
	}

	if contains(errMsg, "syntax") || contains(errMsg, "invalid") {
		return s.createErrorResponse(model.SQLErrorSyntax, "SQL syntax error", errMsg)
	}

	if contains(errMsg, "result set too large") {
		return s.createErrorResponse(model.SQLErrorResultSize, "Result set too large", errMsg)
	}

	// 默认为语法错误
	return s.createErrorResponse(model.SQLErrorSyntax, "SQL execution error", errMsg)
}

// handleBatchExecutionError 处理批量执行错误
func (s *sqlService) handleBatchExecutionError(err error) *model.BatchSQLResponse {
	errMsg := err.Error()

	if contains(errMsg, "transaction") {
		return s.createBatchErrorResponse(model.SQLErrorTransaction, "Transaction failed", errMsg)
	}

	return s.createBatchErrorResponse(model.SQLErrorSyntax, "Batch execution error", errMsg)
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
			 s[len(s)-len(substr):] == substr ||
			 findSubstring(s, substr))))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

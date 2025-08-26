package sql

import (
	"fmt"
	"regexp"
	"strings"
)

// DatabaseDialect 数据库方言接口
type DatabaseDialect interface {
	// 转换分页查询
	ApplyPagination(query string, offset, limit int) string
	// 转换排序查询
	ApplySort(query string, sortBy, sortOrder string) string
	// 获取限制查询
	GetLimitQuery(limit int) string
	// 转换数据类型
	ConvertDataType(value interface{}) interface{}
	// 获取当前时间函数
	GetCurrentTimeFunction() string
	// 检查是否支持某个功能
	SupportsFeature(feature string) bool
}

// PostgreSQLDialect PostgreSQL 方言
type PostgreSQLDialect struct{}

// NewPostgreSQLDialect 创建 PostgreSQL 方言
func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{}
}

// ApplyPagination 应用分页（PostgreSQL 使用 LIMIT OFFSET）
func (d *PostgreSQLDialect) ApplyPagination(query string, offset, limit int) string {
	if limit <= 0 {
		return query
	}

	query = strings.TrimSpace(query)
	if strings.HasSuffix(strings.ToLower(query), ";") {
		query = query[:len(query)-1]
	}

	if offset > 0 {
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)
	}
	return fmt.Sprintf("%s LIMIT %d", query, limit)
}

// ApplySort 应用排序
func (d *PostgreSQLDialect) ApplySort(query string, sortBy, sortOrder string) string {
	if sortBy == "" {
		return query
	}

	query = strings.TrimSpace(query)
	if strings.HasSuffix(strings.ToLower(query), ";") {
		query = query[:len(query)-1]
	}

	// 验证排序方向
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	// 检查查询是否已经有 ORDER BY
	lowerQuery := strings.ToLower(query)
	if strings.Contains(lowerQuery, "order by") {
		return fmt.Sprintf("%s, %s %s", query, sortBy, sortOrder)
	}

	return fmt.Sprintf("%s ORDER BY %s %s", query, sortBy, sortOrder)
}

// GetLimitQuery 获取限制查询
func (d *PostgreSQLDialect) GetLimitQuery(limit int) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

// ConvertDataType 转换数据类型
func (d *PostgreSQLDialect) ConvertDataType(value interface{}) interface{} {
	// PostgreSQL 的数据类型转换
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return value
	}
}

// GetCurrentTimeFunction 获取当前时间函数
func (d *PostgreSQLDialect) GetCurrentTimeFunction() string {
	return "NOW()"
}

// SupportsFeature 检查是否支持某个功能
func (d *PostgreSQLDialect) SupportsFeature(feature string) bool {
	supportedFeatures := map[string]bool{
		"cte":              true,  // Common Table Expressions
		"window_functions": true,  // 窗口函数
		"json":             true,  // JSON 支持
		"arrays":           true,  // 数组支持
		"regex":            true,  // 正则表达式
		"full_text_search": true,  // 全文搜索
		"upsert":           true,  // ON CONFLICT
	}
	return supportedFeatures[feature]
}

// OracleDialect Oracle 方言
type OracleDialect struct{}

// NewOracleDialect 创建 Oracle 方言
func NewOracleDialect() *OracleDialect {
	return &OracleDialect{}
}

// ApplyPagination 应用分页（Oracle 使用 ROWNUM 或 ROW_NUMBER()）
func (d *OracleDialect) ApplyPagination(query string, offset, limit int) string {
	if limit <= 0 {
		return query
	}

	query = strings.TrimSpace(query)
	if strings.HasSuffix(strings.ToLower(query), ";") {
		query = query[:len(query)-1]
	}

	// Oracle 12c+ 支持 OFFSET FETCH
	if offset > 0 {
		return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", query, offset, limit)
	}
	return fmt.Sprintf("%s FETCH FIRST %d ROWS ONLY", query, limit)
}

// ApplySort 应用排序
func (d *OracleDialect) ApplySort(query string, sortBy, sortOrder string) string {
	if sortBy == "" {
		return query
	}

	query = strings.TrimSpace(query)
	if strings.HasSuffix(strings.ToLower(query), ";") {
		query = query[:len(query)-1]
	}

	// 验证排序方向
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "ASC"
	}

	// 检查查询是否已经有 ORDER BY
	lowerQuery := strings.ToLower(query)
	if strings.Contains(lowerQuery, "order by") {
		return fmt.Sprintf("%s, %s %s", query, sortBy, sortOrder)
	}

	return fmt.Sprintf("%s ORDER BY %s %s", query, sortBy, sortOrder)
}

// GetLimitQuery 获取限制查询
func (d *OracleDialect) GetLimitQuery(limit int) string {
	return fmt.Sprintf("FETCH FIRST %d ROWS ONLY", limit)
}

// ConvertDataType 转换数据类型
func (d *OracleDialect) ConvertDataType(value interface{}) interface{} {
	// Oracle 的数据类型转换
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return value
	}
}

// GetCurrentTimeFunction 获取当前时间函数
func (d *OracleDialect) GetCurrentTimeFunction() string {
	return "SYSDATE"
}

// SupportsFeature 检查是否支持某个功能
func (d *OracleDialect) SupportsFeature(feature string) bool {
	supportedFeatures := map[string]bool{
		"cte":              true,  // Common Table Expressions (11g+)
		"window_functions": true,  // 窗口函数
		"json":             true,  // JSON 支持 (12c+)
		"arrays":           false, // 不支持数组
		"regex":            true,  // 正则表达式
		"full_text_search": true,  // Oracle Text
		"upsert":           true,  // MERGE 语句
	}
	return supportedFeatures[feature]
}

// DialectFactory 方言工厂
type DialectFactory struct{}

// NewDialectFactory 创建方言工厂
func NewDialectFactory() *DialectFactory {
	return &DialectFactory{}
}

// CreateDialect 根据数据库类型创建方言
func (f *DialectFactory) CreateDialect(dbType string) DatabaseDialect {
	switch strings.ToLower(dbType) {
	case "postgres", "postgresql":
		return NewPostgreSQLDialect()
	case "oracle":
		return NewOracleDialect()
	default:
		// 默认使用 PostgreSQL 方言
		return NewPostgreSQLDialect()
	}
}

// GetSupportedDialects 获取支持的方言列表
func (f *DialectFactory) GetSupportedDialects() []string {
	return []string{"postgres", "postgresql", "oracle"}
}

// IsDialectSupported 检查是否支持指定的方言
func (f *DialectFactory) IsDialectSupported(dbType string) bool {
	supported := f.GetSupportedDialects()
	lowerDbType := strings.ToLower(dbType)
	
	for _, dialect := range supported {
		if dialect == lowerDbType {
			return true
		}
	}
	return false
}

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	dialect DatabaseDialect
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(dialect DatabaseDialect) *QueryOptimizer {
	return &QueryOptimizer{
		dialect: dialect,
	}
}

// OptimizeQuery 优化查询
func (o *QueryOptimizer) OptimizeQuery(query string) string {
	// 基本的查询优化
	query = strings.TrimSpace(query)
	
	// 移除多余的空格
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	
	// 移除注释
	query = regexp.MustCompile(`--.*$`).ReplaceAllString(query, "")
	query = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(query, "")
	
	return query
}

// AddQueryHints 添加查询提示
func (o *QueryOptimizer) AddQueryHints(query string, hints []string) string {
	if len(hints) == 0 {
		return query
	}

	// 根据数据库类型添加不同的查询提示
	// 这里只是示例，实际实现需要根据具体需求
	return query
}

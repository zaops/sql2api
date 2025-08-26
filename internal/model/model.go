package model

import (
	"strings"
	"time"
)

// ===== 基础响应结构 =====

// APIResponse 统一 API 响应结构
type APIResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Error        *APIError   `json:"error,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
	AffectedRows int64       `json:"affected_rows,omitempty"` // SQL 操作影响的行数
	Total        int64       `json:"total,omitempty"`         // 查询结果总数
	Page         int         `json:"page,omitempty"`          // 当前页码
	PageSize     int         `json:"page_size,omitempty"`     // 每页大小
}

// APIError API 错误结构
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SQLError SQL 专用错误结构
type SQLError struct {
	Code     int    `json:"code"`     // 错误码：4001-4007
	Message  string `json:"message"`  // 错误消息
	Details  string `json:"details,omitempty"` // 详细信息
	SQLState string `json:"sql_state,omitempty"` // 数据库特定的错误状态
	Query    string `json:"query,omitempty"` // 出错的查询（敏感信息已脱敏）
}

// SQL 错误码常量
const (
	SQLErrorSyntax      = 4001 // SQL 语法错误
	SQLErrorParams      = 4002 // 参数错误或缺失
	SQLErrorPermission  = 4003 // 权限不足
	SQLErrorConnection  = 4004 // 数据库连接错误
	SQLErrorTransaction = 4005 // 事务执行失败
	SQLErrorTimeout     = 4006 // 查询超时
	SQLErrorResultSize  = 4007 // 结果集过大
)

// SuccessResponse 成功响应类型别名（用于 Swagger 文档）
type SuccessResponse = APIResponse

// ErrorResponse 错误响应类型别名（用于 Swagger 文档）
type ErrorResponse = APIResponse

// ===== 响应创建函数 =====

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, message ...string) APIResponse {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return APIResponse{
		Success:   true,
		Message:   msg,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string, details ...string) APIResponse {
	apiError := &APIError{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 && details[0] != "" {
		apiError.Details = details[0]
	}

	return APIResponse{
		Success:   false,
		Error:     apiError,
		Timestamp: time.Now(),
	}
}

// NewSuccessResponseWithPagination 创建带分页信息的成功响应
func NewSuccessResponseWithPagination(data interface{}, total int64, page, pageSize int, message ...string) APIResponse {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return APIResponse{
		Success:   true,
		Message:   msg,
		Data:      data,
		Timestamp: time.Now(),
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}
}

// NewSuccessResponseWithAffectedRows 创建带影响行数的成功响应
func NewSuccessResponseWithAffectedRows(data interface{}, affectedRows int64, message ...string) APIResponse {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return APIResponse{
		Success:      true,
		Message:      msg,
		Data:         data,
		Timestamp:    time.Now(),
		AffectedRows: affectedRows,
	}
}

// ===== 健康检查相关 =====

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Services  map[string]string `json:"services,omitempty"`
}

// ===== 数据库迁移相关 =====

// ModelMigrator 模型迁移器接口
type ModelMigrator interface {
	GetModels() []interface{}
}

// DefaultModelMigrator 默认模型迁移器
type DefaultModelMigrator struct{}

// GetModels 获取需要迁移的模型列表
func (m *DefaultModelMigrator) GetModels() []interface{} {
	// 现在不需要迁移任何模型，因为我们只使用 SQL API
	return []interface{}{}
}

// ===== SQL 相关数据结构 =====

// DatabaseType 数据库类型常量
type DatabaseType string

const (
	DatabasePostgreSQL DatabaseType = "postgres"
	DatabaseOracle     DatabaseType = "oracle"
)

// SQLRequest 通用 SQL 请求结构
type SQLRequest struct {
	DatabaseType string                 `json:"database_type" binding:"required,oneof=postgres oracle" example:"postgres"`
	SQL          string                 `json:"sql,omitempty" example:"SELECT * FROM items WHERE active = $1"`
	Query        *StructuredQuery       `json:"query,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty" example:"{\"active\": true}"`
	Pagination   *PaginationConfig      `json:"pagination,omitempty"`
	Sort         *SortConfig            `json:"sort,omitempty"`
}

// StructuredQuery 结构化查询（JSON 转 SQL）
type StructuredQuery struct {
	Table   string                 `json:"table" binding:"required" example:"items"`
	Action  string                 `json:"action" binding:"required,oneof=select insert update delete" example:"select"`
	Fields  []string               `json:"fields,omitempty" example:"[\"id\", \"name\", \"created_at\"]"`
	Where   map[string]interface{} `json:"where,omitempty" example:"{\"active\": true, \"category\": \"electronics\"}"`
	Data    map[string]interface{} `json:"data,omitempty" example:"{\"name\": \"New Item\", \"category\": \"electronics\"}"`
	GroupBy []string               `json:"group_by,omitempty" example:"[\"category\"]"`
	Having  map[string]interface{} `json:"having,omitempty"`
	OrderBy []OrderByClause        `json:"order_by,omitempty"`
	Limit   int                    `json:"limit,omitempty" example:"100"`
}

// OrderByClause 排序子句
type OrderByClause struct {
	Field string `json:"field" binding:"required" example:"created_at"`
	Order string `json:"order" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// PaginationConfig 分页配置
type PaginationConfig struct {
	Page     int `json:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int `json:"page_size" binding:"omitempty,min=1,max=1000" example:"20"`
}

// SortConfig 排序配置
type SortConfig struct {
	SortBy    string `json:"sort_by" example:"created_at"`
	SortOrder string `json:"sort_order" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// BatchSQLRequest 批量 SQL 请求结构
type BatchSQLRequest struct {
	DatabaseType   string       `json:"database_type" binding:"required,oneof=postgres oracle" example:"postgres"`
	Operations     []SQLRequest `json:"operations" binding:"required,min=1,max=100"`
	Transactional  bool         `json:"transactional" example:"true"`
	ContinueOnError bool        `json:"continue_on_error" example:"false"`
}

// InsertRequest 便捷插入请求结构
type InsertRequest struct {
	DatabaseType string                 `json:"database_type" binding:"required,oneof=postgres oracle" example:"postgres"`
	Table        string                 `json:"table" binding:"required" example:"items"`
	Data         map[string]interface{} `json:"data" binding:"required" example:"{\"name\": \"New Item\", \"category\": \"electronics\"}"`
	OnConflict   string                 `json:"on_conflict,omitempty" binding:"omitempty,oneof=ignore update" example:"ignore"`
	ReturnFields []string               `json:"return_fields,omitempty" example:"[\"id\", \"created_at\"]"`
}

// BatchInsertRequest 批量插入请求结构
type BatchInsertRequest struct {
	DatabaseType string                   `json:"database_type" binding:"required,oneof=postgres oracle" example:"postgres"`
	Table        string                   `json:"table" binding:"required" example:"items"`
	Data         []map[string]interface{} `json:"data" binding:"required,min=1,max=1000"`
	OnConflict   string                   `json:"on_conflict,omitempty" binding:"omitempty,oneof=ignore update" example:"ignore"`
	ReturnFields []string                 `json:"return_fields,omitempty" example:"[\"id\", \"created_at\"]"`
}

// ===== SQL 响应结构 =====

// SQLResponse SQL 响应结构（基于 APIResponse 扩展）
type SQLResponse struct {
	Success      bool                     `json:"success"`
	Message      string                   `json:"message,omitempty"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Error        *SQLError                `json:"error,omitempty"`
	Timestamp    time.Time                `json:"timestamp"`
	AffectedRows int64                    `json:"affected_rows"`
	Total        int64                    `json:"total,omitempty"`
	Page         int                      `json:"page,omitempty"`
	PageSize     int                      `json:"page_size,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	ExecutionTime float64                 `json:"execution_time,omitempty"` // 执行时间（毫秒）
}

// BatchSQLResponse 批量 SQL 响应结构
type BatchSQLResponse struct {
	Success           bool                     `json:"success"`
	Message           string                   `json:"message,omitempty"`
	Results           []SQLOperationResult     `json:"results"`
	Error             *SQLError                `json:"error,omitempty"`
	Timestamp         time.Time                `json:"timestamp"`
	TotalAffectedRows int64                    `json:"total_affected_rows"`
	ExecutedCount     int                      `json:"executed_count"`
	FailedCount       int                      `json:"failed_count"`
	ExecutionTime     float64                  `json:"execution_time,omitempty"`
}

// SQLOperationResult 单个 SQL 操作结果
type SQLOperationResult struct {
	Index         int                      `json:"index"`
	Success       bool                     `json:"success"`
	AffectedRows  int64                    `json:"affected_rows"`
	Data          []map[string]interface{} `json:"data,omitempty"`
	Error         *SQLError                `json:"error,omitempty"`
	ExecutionTime float64                  `json:"execution_time,omitempty"`
}

// ===== SQL 响应创建函数 =====

// NewSQLSuccessResponse 创建 SQL 成功响应
func NewSQLSuccessResponse(data []map[string]interface{}, affectedRows int64, message ...string) SQLResponse {
	msg := "SQL executed successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	return SQLResponse{
		Success:      true,
		Message:      msg,
		Data:         data,
		Timestamp:    time.Now(),
		AffectedRows: affectedRows,
		Total:        int64(len(data)),
	}
}

// NewSQLErrorResponse 创建 SQL 错误响应
func NewSQLErrorResponse(code int, message string, details ...string) SQLResponse {
	sqlError := &SQLError{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 && details[0] != "" {
		sqlError.Details = details[0]
	}

	return SQLResponse{
		Success:   false,
		Error:     sqlError,
		Timestamp: time.Now(),
	}
}

// NewBatchSQLSuccessResponse 创建批量 SQL 成功响应
func NewBatchSQLSuccessResponse(results []SQLOperationResult, message ...string) BatchSQLResponse {
	msg := "Batch SQL executed successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	var totalAffectedRows int64
	executedCount := 0
	failedCount := 0

	for _, result := range results {
		if result.Success {
			totalAffectedRows += result.AffectedRows
			executedCount++
		} else {
			failedCount++
		}
	}

	return BatchSQLResponse{
		Success:           failedCount == 0,
		Message:           msg,
		Results:           results,
		Timestamp:         time.Now(),
		TotalAffectedRows: totalAffectedRows,
		ExecutedCount:     executedCount,
		FailedCount:       failedCount,
	}
}

// NewBatchSQLErrorResponse 创建批量 SQL 错误响应
func NewBatchSQLErrorResponse(code int, message string, details ...string) BatchSQLResponse {
	sqlError := &SQLError{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 && details[0] != "" {
		sqlError.Details = details[0]
	}

	return BatchSQLResponse{
		Success:   false,
		Error:     sqlError,
		Timestamp: time.Now(),
	}
}

// ===== 验证函数 =====

// ValidateDatabaseType 验证数据库类型
func ValidateDatabaseType(dbType string) bool {
	validTypes := []DatabaseType{DatabasePostgreSQL, DatabaseOracle}
	for _, validType := range validTypes {
		if DatabaseType(dbType) == validType {
			return true
		}
	}
	return false
}

// ValidateSQLAction 验证 SQL 操作类型
func ValidateSQLAction(action string) bool {
	validActions := []string{"select", "insert", "update", "delete"}
	for _, validAction := range validActions {
		if strings.ToLower(action) == validAction {
			return true
		}
	}
	return false
}

// ValidateOnConflictAction 验证冲突处理动作
func ValidateOnConflictAction(action string) bool {
	if action == "" {
		return true // 空值是有效的
	}
	validActions := []string{"ignore", "update"}
	for _, validAction := range validActions {
		if strings.ToLower(action) == validAction {
			return true
		}
	}
	return false
}

// ValidateSortOrder 验证排序方向
func ValidateSortOrder(order string) bool {
	if order == "" {
		return true // 空值是有效的，默认为 asc
	}
	validOrders := []string{"asc", "desc"}
	for _, validOrder := range validOrders {
		if strings.ToLower(order) == validOrder {
			return true
		}
	}
	return false
}

// GetSQLErrorMessage 根据错误码获取错误消息
func GetSQLErrorMessage(code int) string {
	messages := map[int]string{
		SQLErrorSyntax:      "SQL syntax error",
		SQLErrorParams:      "Parameter error or missing",
		SQLErrorPermission:  "Insufficient permissions",
		SQLErrorConnection:  "Database connection error",
		SQLErrorTransaction: "Transaction execution failed",
		SQLErrorTimeout:     "Query timeout",
		SQLErrorResultSize:  "Result set too large",
	}

	if msg, exists := messages[code]; exists {
		return msg
	}
	return "Unknown SQL error"
}

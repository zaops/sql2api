package handler

import (
	"fmt"
	"net/http"
	"strings"

	"sql2api/internal/model"
	"sql2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SQLHandler SQL API 处理器
type SQLHandler struct {
	sqlService service.SQLService
}

// NewSQLHandler 创建 SQL API 处理器
func NewSQLHandler(sqlService service.SQLService) *SQLHandler {
	return &SQLHandler{
		sqlService: sqlService,
	}
}

// HandleSQL 通用 SQL 查询端点
// @Summary 执行 SQL 查询
// @Description 支持原生 SQL 和结构化查询，包含分页和排序功能
// @Tags SQL
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.SQLRequest true "SQL 查询请求"
// @Success 200 {object} model.SQLResponse "查询成功"
// @Failure 400 {object} model.SQLResponse "请求格式错误"
// @Failure 401 {object} model.SQLResponse "未认证"
// @Failure 403 {object} model.SQLResponse "权限不足"
// @Failure 500 {object} model.SQLResponse "服务器内部错误"
// @Router /api/v1/sql [post]
func (h *SQLHandler) HandleSQL(c *gin.Context) {
	var req model.SQLRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.NewSQLErrorResponse(model.SQLErrorParams, "Invalid request format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 检查权限
	if !h.checkSQLPermission(c, &req) {
		response := model.NewSQLErrorResponse(model.SQLErrorPermission, "Insufficient permissions")
		c.JSON(http.StatusForbidden, response)
		return
	}

	// 根据查询类型选择执行方法
	var response *model.SQLResponse
	var err error

	if h.isQueryOperation(&req) {
		// 执行查询操作
		response, err = h.sqlService.ExecuteQuery(c.Request.Context(), &req)
	} else {
		// 执行修改操作
		response, err = h.sqlService.ExecuteSQL(c.Request.Context(), &req)
	}

	if err != nil {
		response = model.NewSQLErrorResponse(model.SQLErrorSyntax, "SQL execution failed", err.Error())
		c.JSON(http.StatusInternalServerError, *response)
		return
	}

	// 根据响应状态设置 HTTP 状态码
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = h.getHTTPStatusFromSQLError(response.Error)
	}

	c.JSON(statusCode, response)
}

// HandleBatchSQL 批量 SQL 操作端点
// @Summary 执行批量 SQL 操作
// @Description 支持批量 SQL 操作，可选择事务模式
// @Tags SQL
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.BatchSQLRequest true "批量 SQL 请求"
// @Success 200 {object} model.BatchSQLResponse "批量操作成功"
// @Failure 400 {object} model.BatchSQLResponse "请求格式错误"
// @Failure 401 {object} model.BatchSQLResponse "未认证"
// @Failure 403 {object} model.BatchSQLResponse "权限不足"
// @Failure 500 {object} model.BatchSQLResponse "服务器内部错误"
// @Router /api/v1/sql/batch [post]
func (h *SQLHandler) HandleBatchSQL(c *gin.Context) {
	var req model.BatchSQLRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.NewBatchSQLErrorResponse(model.SQLErrorParams, "Invalid request format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 检查批量操作权限
	if !h.checkBatchPermission(c) {
		response := model.NewBatchSQLErrorResponse(model.SQLErrorPermission, "Insufficient permissions for batch operations")
		c.JSON(http.StatusForbidden, response)
		return
	}

	// 检查每个操作的权限
	for i, operation := range req.Operations {
		if !h.checkSQLPermission(c, &operation) {
			response := model.NewBatchSQLErrorResponse(model.SQLErrorPermission, 
				"Insufficient permissions", 
				fmt.Sprintf("Operation %d permission denied", i))
			c.JSON(http.StatusForbidden, response)
			return
		}
	}

	// 执行批量操作
	response, err := h.sqlService.ExecuteBatch(c.Request.Context(), &req)
	if err != nil {
		response = model.NewBatchSQLErrorResponse(model.SQLErrorTransaction, "Batch execution failed", err.Error())
		c.JSON(http.StatusInternalServerError, *response)
		return
	}

	// 根据响应状态设置 HTTP 状态码
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = h.getHTTPStatusFromSQLError(response.Error)
	}

	c.JSON(statusCode, response)
}

// HandleInsertSQL 便捷插入端点
// @Summary 执行便捷插入操作
// @Description 提供简化的插入操作，支持冲突处理
// @Tags SQL
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.InsertRequest true "插入请求"
// @Success 201 {object} model.SQLResponse "插入成功"
// @Failure 400 {object} model.SQLResponse "请求格式错误"
// @Failure 401 {object} model.SQLResponse "未认证"
// @Failure 403 {object} model.SQLResponse "权限不足"
// @Failure 500 {object} model.SQLResponse "服务器内部错误"
// @Router /api/v1/sql/insert [post]
func (h *SQLHandler) HandleInsertSQL(c *gin.Context) {
	var req model.InsertRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.NewSQLErrorResponse(model.SQLErrorParams, "Invalid request format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 检查插入权限
	if !h.checkInsertPermission(c, req.Table) {
		response := model.NewSQLErrorResponse(model.SQLErrorPermission, "Insufficient permissions for insert operation")
		c.JSON(http.StatusForbidden, response)
		return
	}

	// 执行插入操作
	response, err := h.sqlService.ExecuteInsert(c.Request.Context(), &req)
	if err != nil {
		response = model.NewSQLErrorResponse(model.SQLErrorSyntax, "Insert execution failed", err.Error())
		c.JSON(http.StatusInternalServerError, *response)
		return
	}

	// 根据响应状态设置 HTTP 状态码
	statusCode := http.StatusCreated
	if !response.Success {
		statusCode = h.getHTTPStatusFromSQLError(response.Error)
	}

	c.JSON(statusCode, response)
}

// HandleBatchInsert 批量插入端点
// @Summary 执行批量插入操作
// @Description 提供批量插入功能，支持冲突处理
// @Tags SQL
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body model.BatchInsertRequest true "批量插入请求"
// @Success 201 {object} model.SQLResponse "批量插入成功"
// @Failure 400 {object} model.SQLResponse "请求格式错误"
// @Failure 401 {object} model.SQLResponse "未认证"
// @Failure 403 {object} model.SQLResponse "权限不足"
// @Failure 500 {object} model.SQLResponse "服务器内部错误"
// @Router /api/v1/sql/batch-insert [post]
func (h *SQLHandler) HandleBatchInsert(c *gin.Context) {
	var req model.BatchInsertRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.NewSQLErrorResponse(model.SQLErrorParams, "Invalid request format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 检查批量插入权限
	if !h.checkInsertPermission(c, req.Table) {
		response := model.NewSQLErrorResponse(model.SQLErrorPermission, "Insufficient permissions for batch insert operation")
		c.JSON(http.StatusForbidden, response)
		return
	}

	// 执行批量插入操作
	response, err := h.sqlService.ExecuteBatchInsert(c.Request.Context(), &req)
	if err != nil {
		response = model.NewSQLErrorResponse(model.SQLErrorSyntax, "Batch insert execution failed", err.Error())
		c.JSON(http.StatusInternalServerError, *response)
		return
	}

	// 根据响应状态设置 HTTP 状态码
	statusCode := http.StatusCreated
	if !response.Success {
		statusCode = h.getHTTPStatusFromSQLError(response.Error)
	}

	c.JSON(statusCode, response)
}

// ===== 辅助方法 =====

// isQueryOperation 判断是否为查询操作
func (h *SQLHandler) isQueryOperation(req *model.SQLRequest) bool {
	if req.SQL != "" {
		// 检查原生 SQL 是否为查询操作
		sql := strings.TrimSpace(strings.ToLower(req.SQL))
		return strings.HasPrefix(sql, "select") || strings.HasPrefix(sql, "with")
	}

	if req.Query != nil {
		// 检查结构化查询的操作类型
		return strings.ToLower(req.Query.Action) == "select"
	}

	return false
}

// checkSQLPermission 检查 SQL 操作权限
func (h *SQLHandler) checkSQLPermission(c *gin.Context, req *model.SQLRequest) bool {
	// 获取操作类型
	action := h.getSQLAction(req)
	if action == "" {
		return false
	}

	// 检查对应的权限
	permission := fmt.Sprintf("sql.%s", action)
	return h.hasPermission(c, permission)
}

// checkBatchPermission 检查批量操作权限
func (h *SQLHandler) checkBatchPermission(c *gin.Context) bool {
	return h.hasPermission(c, "sql.batch")
}

// checkInsertPermission 检查插入权限
func (h *SQLHandler) checkInsertPermission(c *gin.Context, table string) bool {
	// 检查基本插入权限
	if !h.hasPermission(c, "sql.insert") {
		return false
	}

	// 可以在这里添加表级别的权限检查
	// 例如：检查是否有访问特定表的权限

	return true
}

// getSQLAction 获取 SQL 操作类型
func (h *SQLHandler) getSQLAction(req *model.SQLRequest) string {
	if req.SQL != "" {
		// 从原生 SQL 中提取操作类型
		sql := strings.TrimSpace(strings.ToLower(req.SQL))
		words := strings.Fields(sql)
		if len(words) > 0 {
			switch words[0] {
			case "select", "with":
				return "query"
			case "insert":
				return "insert"
			case "update":
				return "update"
			case "delete":
				return "delete"
			}
		}
	}

	if req.Query != nil {
		// 从结构化查询中获取操作类型
		action := strings.ToLower(req.Query.Action)
		if action == "select" {
			return "query"
		}
		return action
	}

	return ""
}

// hasPermission 检查是否有指定权限
func (h *SQLHandler) hasPermission(c *gin.Context, permission string) bool {
	// 从上下文中获取 API Key 管理器
	apiKeyManager, exists := c.Get("api_key_manager")
	if !exists {
		return false
	}

	// 获取 API Key
	apiKey := h.getAPIKey(c)
	if apiKey == "" {
		return false
	}

	// 检查权限
	if manager, ok := apiKeyManager.(interface{ HasPermission(string, string) bool }); ok {
		return manager.HasPermission(apiKey, permission)
	}

	return false
}

// getAPIKey 从请求中获取 API Key
func (h *SQLHandler) getAPIKey(c *gin.Context) string {
	// 从 Header 中获取
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// 从查询参数中获取
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// getHTTPStatusFromSQLError 根据 SQL 错误码获取 HTTP 状态码
func (h *SQLHandler) getHTTPStatusFromSQLError(sqlError *model.SQLError) int {
	if sqlError == nil {
		return http.StatusInternalServerError
	}

	switch sqlError.Code {
	case model.SQLErrorParams:
		return http.StatusBadRequest
	case model.SQLErrorPermission:
		return http.StatusForbidden
	case model.SQLErrorConnection:
		return http.StatusServiceUnavailable
	case model.SQLErrorTimeout:
		return http.StatusRequestTimeout
	case model.SQLErrorSyntax:
		return http.StatusBadRequest
	case model.SQLErrorTransaction:
		return http.StatusInternalServerError
	case model.SQLErrorResultSize:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

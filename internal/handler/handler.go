package handler

import (
	"net/http"

	"sql2api/internal/model"
	"sql2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器（仅保留健康检查）
type AuthHandler struct{}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Health 健康检查
// @Summary 认证服务健康检查
// @Description 检查认证服务的健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} model.SuccessResponse "服务健康"
// @Router /api/v1/auth/health [post]
func (h *AuthHandler) Health(c *gin.Context) {
	healthInfo := map[string]interface{}{
		"status":      "healthy",
		"auth_method": "api_key",
		"timestamp":   c.GetHeader("Date"),
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(healthInfo, "Authentication service is healthy"))
}

// Handlers 处理器集合
type Handlers struct {
	Auth     *AuthHandler
	Resource *ResourceHandler
	SQL      *SQLHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(services *service.Services) *Handlers {
	handlers := &Handlers{
		Auth:     NewAuthHandler(),
		Resource: NewResourceHandler(services.Item),
	}

	// 如果 SQL 服务可用，则创建 SQL 处理器
	if services.SQL != nil {
		handlers.SQL = NewSQLHandler(services.SQL)
	}

	return handlers
}

// HandlerManager 处理器管理器接口
type HandlerManager interface {
	GetAuthHandler() *AuthHandler
	GetResourceHandler() *ResourceHandler
	GetSQLHandler() *SQLHandler
}

// handlerManager 处理器管理器实现
type handlerManager struct {
	handlers *Handlers
}

// NewHandlerManager 创建处理器管理器
func NewHandlerManager(services *service.Services) HandlerManager {
	return &handlerManager{
		handlers: NewHandlers(services),
	}
}

// GetAuthHandler 获取认证处理器
func (hm *handlerManager) GetAuthHandler() *AuthHandler {
	return hm.handlers.Auth
}

// GetResourceHandler 获取资源处理器
func (hm *handlerManager) GetResourceHandler() *ResourceHandler {
	return hm.handlers.Resource
}

// GetSQLHandler 获取 SQL 处理器
func (hm *handlerManager) GetSQLHandler() *SQLHandler {
	return hm.handlers.SQL
}

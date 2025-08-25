package handler

import (
	"sql2api/internal/middleware"
	"sql2api/internal/service"
)

// Handlers 处理器集合
type Handlers struct {
	Auth     *AuthHandler
	Resource *ResourceHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(services *service.Services, jwtManager *middleware.JWTManager) *Handlers {
	return &Handlers{
		Auth:     NewAuthHandler(services.User, jwtManager),
		Resource: NewResourceHandler(services.User, services.Item),
	}
}

// HandlerManager 处理器管理器接口
type HandlerManager interface {
	GetAuthHandler() *AuthHandler
	GetResourceHandler() *ResourceHandler
}

// handlerManager 处理器管理器实现
type handlerManager struct {
	handlers *Handlers
}

// NewHandlerManager 创建处理器管理器
func NewHandlerManager(services *service.Services, jwtManager *middleware.JWTManager) HandlerManager {
	return &handlerManager{
		handlers: NewHandlers(services, jwtManager),
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

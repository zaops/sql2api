package model

import (
	"time"
)

// APIResponse 统一 API 响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError API 错误结构
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse 成功响应类型别名（用于 Swagger 文档）
type SuccessResponse = APIResponse

// ErrorResponse 错误响应类型别名（用于 Swagger 文档）
type ErrorResponse = APIResponse

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

// UnifiedRequest 统一请求结构
type UnifiedRequest struct {
	Action  string      `json:"action" validate:"required,oneof=create read update delete"`
	Payload interface{} `json:"payload"`
}

// ResourceRequest 统一资源请求
type ResourceRequest struct {
	Action   string                 `json:"action" binding:"required" example:"create"`
	Resource string                 `json:"resource" binding:"required" example:"item"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// PaginationRequest 分页请求结构
type PaginationRequest struct {
	Page     int    `json:"page,omitempty" form:"page" validate:"omitempty,min=1"`
	PageSize int    `json:"page_size,omitempty" form:"page_size" validate:"omitempty,min=1,max=100"`
	OrderBy  string `json:"order_by,omitempty" form:"order_by"`
	Order    string `json:"order,omitempty" form:"order" validate:"omitempty,oneof=asc desc"`
}

// GetOffset 获取分页偏移量
func (p *PaginationRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取分页限制
func (p *PaginationRequest) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// GetOrder 获取排序方向
func (p *PaginationRequest) GetOrder() string {
	if p.Order == "desc" {
		return "desc"
	}
	return "asc"
}

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// CalculateTotalPages 计算总页数
func (p *PaginationResponse) CalculateTotalPages() {
	if p.PageSize > 0 {
		p.TotalPages = int((p.Total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Services  map[string]string `json:"services,omitempty"`
}

// ActionType 操作类型常量
type ActionType string

const (
	ActionCreate    ActionType = "create"
	ActionRead      ActionType = "read"
	ActionList      ActionType = "list"
	ActionGet       ActionType = "get"
	ActionUpdate    ActionType = "update"
	ActionDelete    ActionType = "delete"
	ActionSearch    ActionType = "search"
	ActionSetActive ActionType = "set_active"
)

// ResourceType 资源类型常量
type ResourceType string

const (
	ResourceUser ResourceType = "user"
	ResourceItem ResourceType = "item"
)

// ValidateAction 验证操作类型
func ValidateAction(action string) bool {
	validActions := []ActionType{
		ActionCreate, ActionRead, ActionList, ActionGet,
		ActionUpdate, ActionDelete, ActionSearch, ActionSetActive,
	}
	for _, validAction := range validActions {
		if ActionType(action) == validAction {
			return true
		}
	}
	return false
}

// ValidateResource 验证资源类型
func ValidateResource(resource string) bool {
	validResources := []ResourceType{ResourceUser, ResourceItem}
	for _, validResource := range validResources {
		if ResourceType(resource) == validResource {
			return true
		}
	}
	return false
}

// DatabaseConfig 数据库配置接口（用于模型层）
type DatabaseConfig interface {
	GetType() string
	GetDSN() string
}

// ModelMigrator 模型迁移接口
type ModelMigrator interface {
	Migrate() error
	GetModels() []interface{}
}

// DefaultModelMigrator 默认模型迁移器
type DefaultModelMigrator struct{}

// GetModels 获取所有需要迁移的模型
func (m *DefaultModelMigrator) GetModels() []interface{} {
	return []interface{}{
		&User{},
		&Item{},
	}
}

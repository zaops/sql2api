package model

import (
	"time"

	"gorm.io/gorm"
)

// Item 项目模型
type Item struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"type:varchar(255);not null;index" json:"name" validate:"required,min=1,max=255"`
	Value       int64          `gorm:"not null;default:0" json:"value"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Category    string         `gorm:"type:varchar(100);index" json:"category,omitempty" validate:"omitempty,max=100"`
	Tags        string         `gorm:"type:text" json:"tags,omitempty"` // JSON 字符串存储标签数组
	IsActive    bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedBy   uint           `gorm:"not null;index" json:"created_by"` // 创建者用户ID
	
	// 关联关系
	Creator User `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"creator,omitempty"`
}

// TableName 指定表名
func (Item) TableName() string {
	return "items"
}

// BeforeCreate GORM 钩子：创建前处理
func (i *Item) BeforeCreate(tx *gorm.DB) error {
	// 确保名称不为空
	if i.Name == "" {
		return gorm.ErrInvalidData
	}
	
	// 如果没有设置 IsActive，默认为 true
	if !i.IsActive {
		i.IsActive = true
	}
	
	// 确保有创建者ID
	if i.CreatedBy == 0 {
		return gorm.ErrInvalidData
	}
	
	return nil
}

// BeforeUpdate GORM 钩子：更新前处理
func (i *Item) BeforeUpdate(tx *gorm.DB) error {
	// 确保名称不为空
	if i.Name == "" {
		return gorm.ErrInvalidData
	}
	
	return nil
}

// ItemCreateRequest 创建项目请求结构
type ItemCreateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Value       int64  `json:"value,omitempty"`
	Description string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Category    string `json:"category,omitempty" validate:"omitempty,max=100"`
	Tags        string `json:"tags,omitempty"`
}

// ItemUpdateRequest 更新项目请求结构
type ItemUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Value       *int64  `json:"value,omitempty"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Category    *string `json:"category,omitempty" validate:"omitempty,max=100"`
	Tags        *string `json:"tags,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// ItemResponse 项目响应结构
type ItemResponse struct {
	ID          uint         `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Name        string       `json:"name"`
	Value       int64        `json:"value"`
	Description string       `json:"description,omitempty"`
	Category    string       `json:"category,omitempty"`
	Tags        string       `json:"tags,omitempty"`
	IsActive    bool         `json:"is_active"`
	CreatedBy   uint         `json:"created_by"`
	Creator     UserResponse `json:"creator,omitempty"`
}

// ToResponse 转换为响应结构
func (i *Item) ToResponse() ItemResponse {
	response := ItemResponse{
		ID:          i.ID,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		Name:        i.Name,
		Value:       i.Value,
		Description: i.Description,
		Category:    i.Category,
		Tags:        i.Tags,
		IsActive:    i.IsActive,
		CreatedBy:   i.CreatedBy,
	}
	
	// 如果加载了创建者信息，则包含在响应中
	if i.Creator.ID != 0 {
		response.Creator = i.Creator.ToResponse()
	}
	
	return response
}

// ItemQueryRequest 查询项目请求结构
type ItemQueryRequest struct {
	Name     string `json:"name,omitempty" form:"name"`
	Category string `json:"category,omitempty" form:"category"`
	IsActive *bool  `json:"is_active,omitempty" form:"is_active"`
	Page     int    `json:"page,omitempty" form:"page" validate:"omitempty,min=1"`
	PageSize int    `json:"page_size,omitempty" form:"page_size" validate:"omitempty,min=1,max=100"`
	OrderBy  string `json:"order_by,omitempty" form:"order_by"` // id, name, value, created_at, updated_at
	Order    string `json:"order,omitempty" form:"order"`       // asc, desc
}

// GetOffset 获取分页偏移量
func (q *ItemQueryRequest) GetOffset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	return (q.Page - 1) * q.PageSize
}

// GetLimit 获取分页限制
func (q *ItemQueryRequest) GetLimit() int {
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q.PageSize
}

// GetOrderBy 获取排序字段
func (q *ItemQueryRequest) GetOrderBy() string {
	validOrderBy := map[string]bool{
		"id":         true,
		"name":       true,
		"value":      true,
		"created_at": true,
		"updated_at": true,
	}
	
	if validOrderBy[q.OrderBy] {
		return q.OrderBy
	}
	return "id" // 默认按 ID 排序
}

// GetOrder 获取排序方向
func (q *ItemQueryRequest) GetOrder() string {
	if q.Order == "desc" {
		return "desc"
	}
	return "asc" // 默认升序
}

// ItemListResponse 项目列表响应结构
type ItemListResponse struct {
	Items      []ItemResponse `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

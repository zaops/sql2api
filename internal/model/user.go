package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Username     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"username" validate:"required,min=3,max=50"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // 不在 JSON 中返回密码哈希
	Email        string         `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty" validate:"omitempty,email"`
	FullName     string         `gorm:"type:varchar(255)" json:"full_name,omitempty"`
	IsActive     bool           `gorm:"default:true;not null" json:"is_active"`
	LastLoginAt  *time.Time     `gorm:"" json:"last_login_at,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建前处理
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 确保用户名不为空
	if u.Username == "" {
		return gorm.ErrInvalidData
	}

	// 如果没有设置 IsActive，默认为 true
	if !u.IsActive {
		u.IsActive = true
	}

	return nil
}

// SetPassword 设置密码（自动哈希）
func (u *User) SetPassword(password string) error {
	if len(password) < 6 {
		return gorm.ErrInvalidData
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdateLastLogin 更新最后登录时间
func (u *User) UpdateLastLogin(tx *gorm.DB) error {
	now := time.Now()
	u.LastLoginAt = &now
	return tx.Model(u).Update("last_login_at", now).Error
}

// UserCreateRequest 创建用户请求结构
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	FullName string `json:"full_name,omitempty" validate:"omitempty,max=255"`
}

// UserUpdateRequest 更新用户请求结构
type UserUpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	FullName *string `json:"full_name,omitempty" validate:"omitempty,max=255"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// UserResponse 用户响应结构（不包含敏感信息）
type UserResponse struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Username    string     `json:"username"`
	Email       string     `json:"email,omitempty"`
	FullName    string     `json:"full_name,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// ToResponse 转换为响应结构
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required" example:"newpassword123"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

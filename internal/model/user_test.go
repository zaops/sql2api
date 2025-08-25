package model

import (
	"testing"
	"time"
)

func TestUser_SetPassword(t *testing.T) {
	user := &User{}

	// 测试正常密码
	err := user.SetPassword("password123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user.PasswordHash == "" {
		t.Error("Expected password hash to be set")
	}

	if user.PasswordHash == "password123" {
		t.Error("Expected password to be hashed, not stored as plain text")
	}

	// 测试密码太短
	err = user.SetPassword("123")
	if err == nil {
		t.Error("Expected error for short password")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "password123"

	// 设置密码
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("Failed to set password: %v", err)
	}

	// 测试正确密码
	if !user.CheckPassword(password) {
		t.Error("Expected password check to pass")
	}

	// 测试错误密码
	if user.CheckPassword("wrongpassword") {
		t.Error("Expected password check to fail")
	}
}

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:           1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Email:        "test@example.com",
		FullName:     "Test User",
		IsActive:     true,
		LastLoginAt:  &now,
	}

	response := user.ToResponse()

	// 验证响应结构
	if response.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, response.ID)
	}

	if response.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, response.Username)
	}

	if response.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, response.Email)
	}

	// 确保密码哈希不在响应中
	// 这里我们无法直接检查，因为 UserResponse 结构中没有 PasswordHash 字段
	// 这本身就是安全设计的体现
}

func TestUser_BeforeCreate(t *testing.T) {
	tests := []struct {
		name      string
		user      User
		expectErr bool
	}{
		{
			name: "Valid user",
			user: User{
				Username: "testuser",
				IsActive: true,
			},
			expectErr: false,
		},
		{
			name: "Empty username",
			user: User{
				Username: "",
			},
			expectErr: true,
		},
		{
			name: "User with default IsActive",
			user: User{
				Username: "testuser",
				// IsActive 未设置，应该默认为 true
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.BeforeCreate(nil)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// 检查 IsActive 默认值
			if !tt.expectErr && tt.user.Username != "" {
				if !tt.user.IsActive {
					t.Error("Expected IsActive to be true by default")
				}
			}
		})
	}
}

func TestUser_TableName(t *testing.T) {
	user := User{}
	expected := "users"

	if user.TableName() != expected {
		t.Errorf("Expected table name %s, got %s", expected, user.TableName())
	}
}

func TestUserCreateRequest_Validation(t *testing.T) {
	// 这里我们测试结构体的定义是否正确
	// 实际的验证会在 handler 层使用 validator 包进行

	req := UserCreateRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		FullName: "Test User",
	}

	// 验证字段是否正确设置
	if req.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", req.Username)
	}

	if req.Password != "password123" {
		t.Errorf("Expected password password123, got %s", req.Password)
	}

	if req.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", req.Email)
	}
}

func TestLoginRequest_Structure(t *testing.T) {
	req := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	if req.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", req.Username)
	}

	if req.Password != "password123" {
		t.Errorf("Expected password password123, got %s", req.Password)
	}
}

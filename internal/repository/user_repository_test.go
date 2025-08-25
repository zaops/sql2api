package repository

import (
	"testing"

	"sql2api/internal/model"

	"gorm.io/gorm"
)

func TestNewUserRepository(t *testing.T) {
	// 创建一个模拟的 GORM DB 实例（nil 也可以用于测试接口创建）
	var db *gorm.DB

	repo := NewUserRepository(db)
	if repo == nil {
		t.Error("Expected repository to be created")
	}

	// 验证返回的是正确的接口类型
	_, ok := repo.(UserRepository)
	if !ok {
		t.Error("Expected repository to implement UserRepository interface")
	}
}

func TestUserRepository_ValidateInputs(t *testing.T) {
	// 这些测试不需要实际的数据库连接，只测试输入验证逻辑
	var db *gorm.DB
	repo := NewUserRepository(db)

	// 测试 Create 方法的输入验证
	err := repo.Create(nil)
	if err == nil {
		t.Error("Expected error when creating nil user")
	}

	expectedError := "user cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByID 方法的输入验证
	_, err = repo.GetByID(0)
	if err == nil {
		t.Error("Expected error when getting user with ID 0")
	}

	expectedError = "invalid user ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByUsername 方法的输入验证
	_, err = repo.GetByUsername("")
	if err == nil {
		t.Error("Expected error when getting user with empty username")
	}

	expectedError = "username cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByEmail 方法的输入验证
	_, err = repo.GetByEmail("")
	if err == nil {
		t.Error("Expected error when getting user with empty email")
	}

	expectedError = "email cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Update 方法的输入验证
	err = repo.Update(nil)
	if err == nil {
		t.Error("Expected error when updating nil user")
	}

	expectedError = "user cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Update 方法的 ID 验证
	user := &model.User{ID: 0}
	err = repo.Update(user)
	if err == nil {
		t.Error("Expected error when updating user with ID 0")
	}

	expectedError = "user ID cannot be zero"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Delete 方法的输入验证
	err = repo.Delete(0)
	if err == nil {
		t.Error("Expected error when deleting user with ID 0")
	}

	expectedError = "invalid user ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 UpdateLastLogin 方法的输入验证
	err = repo.UpdateLastLogin(0)
	if err == nil {
		t.Error("Expected error when updating last login for user with ID 0")
	}

	expectedError = "invalid user ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 SetActive 方法的输入验证
	err = repo.SetActive(0, true)
	if err == nil {
		t.Error("Expected error when setting active status for user with ID 0")
	}

	expectedError = "invalid user ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestUserRepository_ListValidation(t *testing.T) {
	// 跳过需要数据库连接的测试
	t.Skip("Skipping database-dependent test - requires actual database connection")
}

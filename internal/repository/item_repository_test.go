package repository

import (
	"testing"

	"sql2api/internal/model"

	"gorm.io/gorm"
)

func TestNewItemRepository(t *testing.T) {
	// 创建一个模拟的 GORM DB 实例（nil 也可以用于测试接口创建）
	var db *gorm.DB

	repo := NewItemRepository(db)
	if repo == nil {
		t.Error("Expected repository to be created")
	}

	// 验证返回的是正确的接口类型
	_, ok := repo.(ItemRepository)
	if !ok {
		t.Error("Expected repository to implement ItemRepository interface")
	}
}

func TestItemRepository_ValidateInputs(t *testing.T) {
	// 这些测试不需要实际的数据库连接，只测试输入验证逻辑
	var db *gorm.DB
	repo := NewItemRepository(db)

	// 测试 Create 方法的输入验证
	err := repo.Create(nil)
	if err == nil {
		t.Error("Expected error when creating nil item")
	}

	expectedError := "item cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Create 方法的 CreatedBy 验证
	item := &model.Item{Name: "Test", CreatedBy: 0}
	err = repo.Create(item)
	if err == nil {
		t.Error("Expected error when creating item with CreatedBy 0")
	}

	expectedError = "creator ID cannot be zero"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByID 方法的输入验证
	_, err = repo.GetByID(0)
	if err == nil {
		t.Error("Expected error when getting item with ID 0")
	}

	expectedError = "invalid item ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByIDWithCreator 方法的输入验证
	_, err = repo.GetByIDWithCreator(0)
	if err == nil {
		t.Error("Expected error when getting item with creator with ID 0")
	}

	expectedError = "invalid item ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Update 方法的输入验证
	err = repo.Update(nil)
	if err == nil {
		t.Error("Expected error when updating nil item")
	}

	expectedError = "item cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Update 方法的 ID 验证
	item = &model.Item{ID: 0}
	err = repo.Update(item)
	if err == nil {
		t.Error("Expected error when updating item with ID 0")
	}

	expectedError = "item ID cannot be zero"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Delete 方法的输入验证
	err = repo.Delete(0)
	if err == nil {
		t.Error("Expected error when deleting item with ID 0")
	}

	expectedError = "invalid item ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 GetByCreator 方法的输入验证
	_, _, err = repo.GetByCreator(0, 0, 10)
	if err == nil {
		t.Error("Expected error when getting items by creator with ID 0")
	}

	expectedError = "invalid creator ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 SetActive 方法的输入验证
	err = repo.SetActive(0, true)
	if err == nil {
		t.Error("Expected error when setting active status for item with ID 0")
	}

	expectedError = "invalid item ID"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试 Search 方法的输入验证
	_, _, err = repo.Search("", 0, 10)
	if err == nil {
		t.Error("Expected error when searching with empty keyword")
	}

	expectedError = "search keyword cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestItemRepository_QueryRequestHandling(t *testing.T) {
	// 跳过需要数据库连接的测试
	t.Skip("Skipping database-dependent test - requires actual database connection")

	// 创建一个有效的查询请求
	query := &model.ItemQueryRequest{
		Name:     "test",
		Category: "test-category",
		Page:     1,
		PageSize: 10,
		OrderBy:  "name",
		Order:    "asc",
	}

	// 测试查询请求的方法
	if query.GetOffset() != 0 {
		t.Errorf("Expected offset 0, got %d", query.GetOffset())
	}

	if query.GetLimit() != 10 {
		t.Errorf("Expected limit 10, got %d", query.GetLimit())
	}

	if query.GetOrderBy() != "name" {
		t.Errorf("Expected order by 'name', got %s", query.GetOrderBy())
	}

	if query.GetOrder() != "asc" {
		t.Errorf("Expected order 'asc', got %s", query.GetOrder())
	}
}

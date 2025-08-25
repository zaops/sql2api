package model

import (
	"testing"
	"time"
)

func TestItem_BeforeCreate(t *testing.T) {
	tests := []struct {
		name      string
		item      Item
		expectErr bool
	}{
		{
			name: "Valid item",
			item: Item{
				Name:      "Test Item",
				Value:     100,
				CreatedBy: 1,
				IsActive:  true,
			},
			expectErr: false,
		},
		{
			name: "Empty name",
			item: Item{
				Name:      "",
				CreatedBy: 1,
			},
			expectErr: true,
		},
		{
			name: "No creator",
			item: Item{
				Name:      "Test Item",
				CreatedBy: 0,
			},
			expectErr: true,
		},
		{
			name: "Item with default IsActive",
			item: Item{
				Name:      "Test Item",
				CreatedBy: 1,
				// IsActive 未设置，应该默认为 true
			},
			expectErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.BeforeCreate(nil)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			// 检查 IsActive 默认值
			if !tt.expectErr && tt.item.Name != "" && tt.item.CreatedBy != 0 {
				if !tt.item.IsActive {
					t.Error("Expected IsActive to be true by default")
				}
			}
		})
	}
}

func TestItem_BeforeUpdate(t *testing.T) {
	tests := []struct {
		name      string
		item      Item
		expectErr bool
	}{
		{
			name: "Valid update",
			item: Item{
				Name:  "Updated Item",
				Value: 200,
			},
			expectErr: false,
		},
		{
			name: "Empty name",
			item: Item{
				Name: "",
			},
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.BeforeUpdate(nil)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestItem_ToResponse(t *testing.T) {
	now := time.Now()
	creator := User{
		ID:       1,
		Username: "creator",
		Email:    "creator@example.com",
	}
	
	item := &Item{
		ID:          1,
		CreatedAt:   now,
		UpdatedAt:   now,
		Name:        "Test Item",
		Value:       100,
		Description: "Test Description",
		Category:    "Test Category",
		Tags:        "tag1,tag2",
		IsActive:    true,
		CreatedBy:   1,
		Creator:     creator,
	}
	
	response := item.ToResponse()
	
	// 验证响应结构
	if response.ID != item.ID {
		t.Errorf("Expected ID %d, got %d", item.ID, response.ID)
	}
	
	if response.Name != item.Name {
		t.Errorf("Expected name %s, got %s", item.Name, response.Name)
	}
	
	if response.Value != item.Value {
		t.Errorf("Expected value %d, got %d", item.Value, response.Value)
	}
	
	if response.CreatedBy != item.CreatedBy {
		t.Errorf("Expected created_by %d, got %d", item.CreatedBy, response.CreatedBy)
	}
	
	// 验证创建者信息
	if response.Creator.ID != creator.ID {
		t.Errorf("Expected creator ID %d, got %d", creator.ID, response.Creator.ID)
	}
}

func TestItem_TableName(t *testing.T) {
	item := Item{}
	expected := "items"
	
	if item.TableName() != expected {
		t.Errorf("Expected table name %s, got %s", expected, item.TableName())
	}
}

func TestItemQueryRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		request  ItemQueryRequest
		expected int
	}{
		{
			name: "Default values",
			request: ItemQueryRequest{},
			expected: 0, // (1-1) * 10
		},
		{
			name: "Page 2, PageSize 10",
			request: ItemQueryRequest{
				Page:     2,
				PageSize: 10,
			},
			expected: 10, // (2-1) * 10
		},
		{
			name: "Page 3, PageSize 20",
			request: ItemQueryRequest{
				Page:     3,
				PageSize: 20,
			},
			expected: 40, // (3-1) * 20
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := tt.request.GetOffset()
			if offset != tt.expected {
				t.Errorf("Expected offset %d, got %d", tt.expected, offset)
			}
		})
	}
}

func TestItemQueryRequest_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		request  ItemQueryRequest
		expected int
	}{
		{
			name:     "Default value",
			request:  ItemQueryRequest{},
			expected: 10,
		},
		{
			name: "Custom page size",
			request: ItemQueryRequest{
				PageSize: 20,
			},
			expected: 20,
		},
		{
			name: "Exceeds maximum",
			request: ItemQueryRequest{
				PageSize: 200,
			},
			expected: 100, // 应该被限制为最大值
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := tt.request.GetLimit()
			if limit != tt.expected {
				t.Errorf("Expected limit %d, got %d", tt.expected, limit)
			}
		})
	}
}

func TestItemQueryRequest_GetOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		request  ItemQueryRequest
		expected string
	}{
		{
			name:     "Default order by",
			request:  ItemQueryRequest{},
			expected: "id",
		},
		{
			name: "Valid order by name",
			request: ItemQueryRequest{
				OrderBy: "name",
			},
			expected: "name",
		},
		{
			name: "Valid order by created_at",
			request: ItemQueryRequest{
				OrderBy: "created_at",
			},
			expected: "created_at",
		},
		{
			name: "Invalid order by",
			request: ItemQueryRequest{
				OrderBy: "invalid_field",
			},
			expected: "id", // 应该回退到默认值
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderBy := tt.request.GetOrderBy()
			if orderBy != tt.expected {
				t.Errorf("Expected order by %s, got %s", tt.expected, orderBy)
			}
		})
	}
}

func TestItemQueryRequest_GetOrder(t *testing.T) {
	tests := []struct {
		name     string
		request  ItemQueryRequest
		expected string
	}{
		{
			name:     "Default order",
			request:  ItemQueryRequest{},
			expected: "asc",
		},
		{
			name: "Descending order",
			request: ItemQueryRequest{
				Order: "desc",
			},
			expected: "desc",
		},
		{
			name: "Invalid order",
			request: ItemQueryRequest{
				Order: "invalid",
			},
			expected: "asc", // 应该回退到默认值
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := tt.request.GetOrder()
			if order != tt.expected {
				t.Errorf("Expected order %s, got %s", tt.expected, order)
			}
		})
	}
}

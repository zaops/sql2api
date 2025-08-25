package model

import (
	"fmt"
	"testing"
)

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	message := "Operation successful"

	response := NewSuccessResponse(data, message)

	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.Message != message {
		t.Errorf("Expected message %s, got %s", message, response.Message)
	}

	// 验证数据类型而不是直接比较 map
	if response.Data == nil {
		t.Error("Expected data to be set")
	}

	if response.Error != nil {
		t.Error("Expected error to be nil")
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewSuccessResponse_DefaultMessage(t *testing.T) {
	data := "test data"

	response := NewSuccessResponse(data)

	if response.Message != "Success" {
		t.Errorf("Expected default message 'Success', got %s", response.Message)
	}
}

func TestNewErrorResponse(t *testing.T) {
	code := 400
	message := "Bad Request"
	details := "Invalid input data"

	response := NewErrorResponse(code, message, details)

	if response.Success {
		t.Error("Expected success to be false")
	}

	if response.Data != nil {
		t.Error("Expected data to be nil")
	}

	if response.Error == nil {
		t.Fatal("Expected error to be set")
	}

	if response.Error.Code != code {
		t.Errorf("Expected error code %d, got %d", code, response.Error.Code)
	}

	if response.Error.Message != message {
		t.Errorf("Expected error message %s, got %s", message, response.Error.Message)
	}

	if response.Error.Details != details {
		t.Errorf("Expected error details %s, got %s", details, response.Error.Details)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewErrorResponse_NoDetails(t *testing.T) {
	code := 500
	message := "Internal Server Error"

	response := NewErrorResponse(code, message)

	if response.Error.Details != "" {
		t.Errorf("Expected empty details, got %s", response.Error.Details)
	}
}

func TestPaginationRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected int
	}{
		{
			name:     "Default values",
			request:  PaginationRequest{},
			expected: 0, // (1-1) * 10
		},
		{
			name: "Page 2, PageSize 10",
			request: PaginationRequest{
				Page:     2,
				PageSize: 10,
			},
			expected: 10, // (2-1) * 10
		},
		{
			name: "Page 3, PageSize 20",
			request: PaginationRequest{
				Page:     3,
				PageSize: 20,
			},
			expected: 40, // (3-1) * 20
		},
		{
			name: "Zero page",
			request: PaginationRequest{
				Page:     0,
				PageSize: 10,
			},
			expected: 0, // 应该被设置为 1，然后 (1-1) * 10 = 0
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

func TestPaginationRequest_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected int
	}{
		{
			name:     "Default value",
			request:  PaginationRequest{},
			expected: 10,
		},
		{
			name: "Custom page size",
			request: PaginationRequest{
				PageSize: 20,
			},
			expected: 20,
		},
		{
			name: "Exceeds maximum",
			request: PaginationRequest{
				PageSize: 200,
			},
			expected: 100, // 应该被限制为最大值
		},
		{
			name: "Zero page size",
			request: PaginationRequest{
				PageSize: 0,
			},
			expected: 10, // 应该使用默认值
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

func TestPaginationRequest_GetOrder(t *testing.T) {
	tests := []struct {
		name     string
		request  PaginationRequest
		expected string
	}{
		{
			name:     "Default order",
			request:  PaginationRequest{},
			expected: "asc",
		},
		{
			name: "Descending order",
			request: PaginationRequest{
				Order: "desc",
			},
			expected: "desc",
		},
		{
			name: "Invalid order",
			request: PaginationRequest{
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

func TestPaginationResponse_CalculateTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		response PaginationResponse
		expected int
	}{
		{
			name: "Exact division",
			response: PaginationResponse{
				Total:    100,
				PageSize: 10,
			},
			expected: 10,
		},
		{
			name: "With remainder",
			response: PaginationResponse{
				Total:    105,
				PageSize: 10,
			},
			expected: 11,
		},
		{
			name: "Zero total",
			response: PaginationResponse{
				Total:    0,
				PageSize: 10,
			},
			expected: 0,
		},
		{
			name: "Zero page size",
			response: PaginationResponse{
				Total:    100,
				PageSize: 0,
			},
			expected: 0, // 避免除零错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.response.CalculateTotalPages()
			if tt.response.TotalPages != tt.expected {
				t.Errorf("Expected total pages %d, got %d", tt.expected, tt.response.TotalPages)
			}
		})
	}
}

func TestValidateAction(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{"Valid create", "create", true},
		{"Valid read", "read", true},
		{"Valid update", "update", true},
		{"Valid delete", "delete", true},
		{"Invalid action", "invalid", false},
		{"Empty action", "", false},
		{"Case sensitive", "CREATE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAction(tt.action)
			if result != tt.expected {
				t.Errorf("Expected %v for action %s, got %v", tt.expected, tt.action, result)
			}
		})
	}
}

func TestValidateResource(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		expected bool
	}{
		{"Valid user", "user", true},
		{"Valid item", "item", true},
		{"Invalid resource", "invalid", false},
		{"Empty resource", "", false},
		{"Case sensitive", "USER", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateResource(tt.resource)
			if result != tt.expected {
				t.Errorf("Expected %v for resource %s, got %v", tt.expected, tt.resource, result)
			}
		})
	}
}

func TestDefaultModelMigrator_GetModels(t *testing.T) {
	migrator := &DefaultModelMigrator{}
	models := migrator.GetModels()

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// 验证模型类型
	expectedTypes := []string{"*model.User", "*model.Item"}
	for i, model := range models {
		modelType := fmt.Sprintf("%T", model)
		if modelType != expectedTypes[i] {
			t.Errorf("Expected model type %s, got %s", expectedTypes[i], modelType)
		}
	}
}

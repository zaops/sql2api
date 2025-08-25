package service

import (
	"errors"
	"testing"

	"sql2api/internal/model"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	users  map[uint]*model.User
	nextID uint
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[uint]*model.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(user *model.User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(id uint) (*model.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetByUsername(username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) Update(user *model.User) error {
	if _, exists := m.users[user.ID]; exists {
		m.users[user.ID] = user
		return nil
	}
	return errors.New("user not found")
}

func (m *MockUserRepository) Delete(id uint) error {
	if _, exists := m.users[id]; exists {
		delete(m.users, id)
		return nil
	}
	return errors.New("user not found")
}

func (m *MockUserRepository) List(offset, limit int, orderBy, order string) ([]*model.User, int64, error) {
	users := make([]*model.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

func (m *MockUserRepository) UpdateLastLogin(id uint) error {
	if _, exists := m.users[id]; exists {
		return nil
	}
	return errors.New("user not found")
}

func (m *MockUserRepository) SetActive(id uint, active bool) error {
	if user, exists := m.users[id]; exists {
		user.IsActive = active
		return nil
	}
	return errors.New("user not found")
}

func TestUserService_Register(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	tests := []struct {
		name        string
		req         *model.UserCreateRequest
		expectError bool
	}{
		{
			name: "Valid registration",
			req: &model.UserCreateRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				FullName: "Test User",
			},
			expectError: false,
		},
		{
			name:        "Nil request",
			req:         nil,
			expectError: true,
		},
		{
			name: "Empty username",
			req: &model.UserCreateRequest{
				Username: "",
				Password: "password123",
			},
			expectError: true,
		},
		{
			name: "Empty password",
			req: &model.UserCreateRequest{
				Username: "testuser",
				Password: "",
			},
			expectError: true,
		},
		{
			name: "Short password",
			req: &model.UserCreateRequest{
				Username: "testuser",
				Password: "123",
			},
			expectError: true,
		},
		{
			name: "Short username",
			req: &model.UserCreateRequest{
				Username: "ab",
				Password: "password123",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("Expected user to be created")
				return
			}

			if user.Username != tt.req.Username {
				t.Errorf("Expected username %s, got %s", tt.req.Username, user.Username)
			}

			if user.Email != tt.req.Email {
				t.Errorf("Expected email %s, got %s", tt.req.Email, user.Email)
			}

			if !user.IsActive {
				t.Error("Expected user to be active")
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewUserService(mockRepo)

	// 创建测试用户
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}
	testUser.SetPassword("password123")
	mockRepo.Create(testUser)

	// 创建非激活用户
	inactiveUser := &model.User{
		Username: "inactive",
		Email:    "inactive@example.com",
		IsActive: false,
	}
	inactiveUser.SetPassword("password123")
	mockRepo.Create(inactiveUser)

	tests := []struct {
		name        string
		username    string
		password    string
		expectError bool
	}{
		{
			name:        "Valid login",
			username:    "testuser",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Empty username",
			username:    "",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "Empty password",
			username:    "testuser",
			password:    "",
			expectError: true,
		},
		{
			name:        "Wrong password",
			username:    "testuser",
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "Non-existent user",
			username:    "nonexistent",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "Inactive user",
			username:    "inactive",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Login(tt.username, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("Expected user to be returned")
				return
			}

			if user.Username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, user.Username)
			}
		})
	}
}

package service

import (
	"errors"
	"fmt"

	"sql2api/internal/model"
	"sql2api/internal/repository"
)

// UserService 用户业务服务接口
type UserService interface {
	// 用户认证
	Login(username, password string) (*model.User, error)
	Register(req *model.UserCreateRequest) (*model.User, error)

	// 用户管理
	GetUserByID(id uint) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	UpdateUser(id uint, req *model.UserUpdateRequest) (*model.User, error)
	DeleteUser(id uint) error
	ListUsers(page, pageSize int, orderBy, order string) ([]*model.User, int64, error)

	// 用户状态管理
	SetUserActive(id uint, active bool) error
	UpdateLastLogin(id uint) error
	ChangePassword(id uint, oldPassword, newPassword string) error
}

// userService 用户业务服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户业务服务
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// Login 用户登录验证
func (s *userService) Login(username, password string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	// 根据用户名获取用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// 验证密码
	if !user.CheckPassword(password) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// Register 用户注册
func (s *userService) Register(req *model.UserCreateRequest) (*model.User, error) {
	if req == nil {
		return nil, errors.New("registration request cannot be nil")
	}

	// 验证必填字段
	if req.Username == "" {
		return nil, errors.New("username is required")
	}

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// 验证密码强度
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// 验证用户名长度
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return nil, errors.New("username must be between 3 and 50 characters")
	}

	// 创建用户对象
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: true,
	}

	// 设置密码哈希
	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 保存用户
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*model.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *userService) GetUserByUsername(username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(id uint, req *model.UserUpdateRequest) (*model.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}

	// 获取现有用户
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// 保存更新
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(page, pageSize int, orderBy, order string) ([]*model.User, int64, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 100 {
		pageSize = 100
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取用户列表
	users, total, err := s.userRepo.List(offset, pageSize, orderBy, order)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// SetUserActive 设置用户激活状态
func (s *userService) SetUserActive(id uint, active bool) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	if err := s.userRepo.SetActive(id, active); err != nil {
		return fmt.Errorf("failed to set user active status: %w", err)
	}

	return nil
}

// UpdateLastLogin 更新最后登录时间
func (s *userService) UpdateLastLogin(id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	if err := s.userRepo.UpdateLastLogin(id); err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(id uint, oldPassword, newPassword string) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}

	if oldPassword == "" {
		return errors.New("old password cannot be empty")
	}

	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	if len(newPassword) < 6 {
		return errors.New("new password must be at least 6 characters long")
	}

	// 获取用户
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证旧密码
	if !user.CheckPassword(oldPassword) {
		return errors.New("old password is incorrect")
	}

	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 保存更新
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

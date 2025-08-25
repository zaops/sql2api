package repository

import (
	"errors"
	"fmt"

	"sql2api/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(offset, limit int, orderBy, order string) ([]*model.User, int64, error)
	UpdateLastLogin(id uint) error
	SetActive(id uint, active bool) error
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	
	// 检查用户名是否已存在
	var existingUser model.User
	if err := r.db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
		return fmt.Errorf("username '%s' already exists", user.Username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check username uniqueness: %w", err)
	}
	
	// 检查邮箱是否已存在（如果提供了邮箱）
	if user.Email != "" {
		if err := r.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			return fmt.Errorf("email '%s' already exists", user.Email)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}
	
	// 创建用户
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(id uint) (*model.User, error) {
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}
	
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with username '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	
	if user.ID == 0 {
		return errors.New("user ID cannot be zero")
	}
	
	// 检查用户是否存在
	var existingUser model.User
	if err := r.db.First(&existingUser, user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user with ID %d not found", user.ID)
		}
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	
	// 检查用户名唯一性（如果用户名发生变化）
	if user.Username != existingUser.Username {
		var duplicateUser model.User
		if err := r.db.Where("username = ? AND id != ?", user.Username, user.ID).First(&duplicateUser).Error; err == nil {
			return fmt.Errorf("username '%s' already exists", user.Username)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check username uniqueness: %w", err)
		}
	}
	
	// 检查邮箱唯一性（如果邮箱发生变化且不为空）
	if user.Email != "" && user.Email != existingUser.Email {
		var duplicateUser model.User
		if err := r.db.Where("email = ? AND id != ?", user.Email, user.ID).First(&duplicateUser).Error; err == nil {
			return fmt.Errorf("email '%s' already exists", user.Email)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}
	
	// 更新用户
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}
	
	// 检查用户是否存在
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user with ID %d not found", id)
		}
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	
	// 软删除用户
	if err := r.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

// List 获取用户列表
func (r *userRepository) List(offset, limit int, orderBy, order string) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64
	
	// 验证排序字段
	validOrderBy := map[string]bool{
		"id":         true,
		"username":   true,
		"email":      true,
		"created_at": true,
		"updated_at": true,
	}
	
	if !validOrderBy[orderBy] {
		orderBy = "id"
	}
	
	if order != "desc" {
		order = "asc"
	}
	
	// 获取总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// 获取用户列表
	query := r.db.Order(fmt.Sprintf("%s %s", orderBy, order))
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	
	return users, total, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *userRepository) UpdateLastLogin(id uint) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}
	
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user with ID %d not found", id)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	if err := user.UpdateLastLogin(r.db); err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	
	return nil
}

// SetActive 设置用户激活状态
func (r *userRepository) SetActive(id uint, active bool) error {
	if id == 0 {
		return errors.New("invalid user ID")
	}
	
	if err := r.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", active).Error; err != nil {
		return fmt.Errorf("failed to set user active status: %w", err)
	}
	
	return nil
}

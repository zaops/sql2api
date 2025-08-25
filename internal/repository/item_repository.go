package repository

import (
	"errors"
	"fmt"

	"sql2api/internal/model"

	"gorm.io/gorm"
)

// ItemRepository 项目数据访问接口
type ItemRepository interface {
	Create(item *model.Item) error
	GetByID(id uint) (*model.Item, error)
	GetByIDWithCreator(id uint) (*model.Item, error)
	Update(item *model.Item) error
	Delete(id uint) error
	List(query *model.ItemQueryRequest) ([]*model.Item, int64, error)
	ListWithCreator(query *model.ItemQueryRequest) ([]*model.Item, int64, error)
	GetByCreator(creatorID uint, offset, limit int) ([]*model.Item, int64, error)
	SetActive(id uint, active bool) error
	Search(keyword string, offset, limit int) ([]*model.Item, int64, error)
}

// itemRepository 项目仓库实现
type itemRepository struct {
	db *gorm.DB
}

// NewItemRepository 创建项目仓库实例
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		db: db,
	}
}

// Create 创建项目
func (r *itemRepository) Create(item *model.Item) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	
	if item.CreatedBy == 0 {
		return errors.New("creator ID cannot be zero")
	}
	
	// 验证创建者是否存在
	var creator model.User
	if err := r.db.First(&creator, item.CreatedBy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("creator with ID %d not found", item.CreatedBy)
		}
		return fmt.Errorf("failed to verify creator: %w", err)
	}
	
	// 创建项目
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	
	return nil
}

// GetByID 根据ID获取项目
func (r *itemRepository) GetByID(id uint) (*model.Item, error) {
	if id == 0 {
		return nil, errors.New("invalid item ID")
	}
	
	var item model.Item
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("item with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get item by ID: %w", err)
	}
	
	return &item, nil
}

// GetByIDWithCreator 根据ID获取项目（包含创建者信息）
func (r *itemRepository) GetByIDWithCreator(id uint) (*model.Item, error) {
	if id == 0 {
		return nil, errors.New("invalid item ID")
	}
	
	var item model.Item
	if err := r.db.Preload("Creator").First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("item with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get item by ID: %w", err)
	}
	
	return &item, nil
}

// Update 更新项目
func (r *itemRepository) Update(item *model.Item) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	
	if item.ID == 0 {
		return errors.New("item ID cannot be zero")
	}
	
	// 检查项目是否存在
	var existingItem model.Item
	if err := r.db.First(&existingItem, item.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("item with ID %d not found", item.ID)
		}
		return fmt.Errorf("failed to check item existence: %w", err)
	}
	
	// 更新项目
	if err := r.db.Save(item).Error; err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	
	return nil
}

// Delete 删除项目（软删除）
func (r *itemRepository) Delete(id uint) error {
	if id == 0 {
		return errors.New("invalid item ID")
	}
	
	// 检查项目是否存在
	var item model.Item
	if err := r.db.First(&item, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("item with ID %d not found", id)
		}
		return fmt.Errorf("failed to check item existence: %w", err)
	}
	
	// 软删除项目
	if err := r.db.Delete(&item).Error; err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	
	return nil
}

// List 获取项目列表
func (r *itemRepository) List(query *model.ItemQueryRequest) ([]*model.Item, int64, error) {
	if query == nil {
		query = &model.ItemQueryRequest{}
	}
	
	var items []*model.Item
	var total int64
	
	// 构建查询条件
	db := r.db.Model(&model.Item{})
	
	// 添加筛选条件
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}
	
	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}
	
	// 添加排序和分页
	orderBy := query.GetOrderBy()
	order := query.GetOrder()
	db = db.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	offset := query.GetOffset()
	limit := query.GetLimit()
	if limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	
	// 获取项目列表
	if err := db.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get items: %w", err)
	}
	
	return items, total, nil
}

// ListWithCreator 获取项目列表（包含创建者信息）
func (r *itemRepository) ListWithCreator(query *model.ItemQueryRequest) ([]*model.Item, int64, error) {
	if query == nil {
		query = &model.ItemQueryRequest{}
	}
	
	var items []*model.Item
	var total int64
	
	// 构建查询条件
	db := r.db.Model(&model.Item{}).Preload("Creator")
	
	// 添加筛选条件
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}
	
	// 获取总数（不包含 Preload）
	countDB := r.db.Model(&model.Item{})
	if query.Name != "" {
		countDB = countDB.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	if query.Category != "" {
		countDB = countDB.Where("category = ?", query.Category)
	}
	if query.IsActive != nil {
		countDB = countDB.Where("is_active = ?", *query.IsActive)
	}
	
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}
	
	// 添加排序和分页
	orderBy := query.GetOrderBy()
	order := query.GetOrder()
	db = db.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	offset := query.GetOffset()
	limit := query.GetLimit()
	if limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	
	// 获取项目列表
	if err := db.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get items: %w", err)
	}
	
	return items, total, nil
}

// GetByCreator 根据创建者获取项目列表
func (r *itemRepository) GetByCreator(creatorID uint, offset, limit int) ([]*model.Item, int64, error) {
	if creatorID == 0 {
		return nil, 0, errors.New("invalid creator ID")
	}
	
	var items []*model.Item
	var total int64
	
	// 获取总数
	if err := r.db.Model(&model.Item{}).Where("created_by = ?", creatorID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count items by creator: %w", err)
	}
	
	// 获取项目列表
	query := r.db.Where("created_by = ?", creatorID).Order("created_at DESC")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get items by creator: %w", err)
	}
	
	return items, total, nil
}

// SetActive 设置项目激活状态
func (r *itemRepository) SetActive(id uint, active bool) error {
	if id == 0 {
		return errors.New("invalid item ID")
	}
	
	if err := r.db.Model(&model.Item{}).Where("id = ?", id).Update("is_active", active).Error; err != nil {
		return fmt.Errorf("failed to set item active status: %w", err)
	}
	
	return nil
}

// Search 搜索项目
func (r *itemRepository) Search(keyword string, offset, limit int) ([]*model.Item, int64, error) {
	if keyword == "" {
		return nil, 0, errors.New("search keyword cannot be empty")
	}
	
	var items []*model.Item
	var total int64
	
	// 构建搜索条件
	searchPattern := "%" + keyword + "%"
	db := r.db.Model(&model.Item{}).Where(
		"name ILIKE ? OR description ILIKE ? OR category ILIKE ? OR tags ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)
	
	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}
	
	// 获取搜索结果
	query := db.Order("created_at DESC")
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	
	if err := query.Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search items: %w", err)
	}
	
	return items, total, nil
}

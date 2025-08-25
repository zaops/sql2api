package service

import (
	"errors"
	"fmt"

	"sql2api/internal/model"
	"sql2api/internal/repository"
)

// ItemService 项目业务服务接口
type ItemService interface {
	// 项目管理
	CreateItem(req *model.ItemCreateRequest, creatorID uint) (*model.Item, error)
	GetItemByID(id uint, includeCreator bool) (*model.Item, error)
	UpdateItem(id uint, req *model.ItemUpdateRequest, userID uint) (*model.Item, error)
	DeleteItem(id uint, userID uint) error
	
	// 项目查询
	ListItems(query *model.ItemQueryRequest, includeCreator bool) (*model.ItemListResponse, error)
	SearchItems(keyword string, page, pageSize int) (*model.ItemListResponse, error)
	GetItemsByCreator(creatorID uint, page, pageSize int) (*model.ItemListResponse, error)
	
	// 项目状态管理
	SetItemActive(id uint, active bool, userID uint) error
	
	// 业务规则验证
	ValidateItemOwnership(itemID, userID uint) error
}

// itemService 项目业务服务实现
type itemService struct {
	itemRepo repository.ItemRepository
	userRepo repository.UserRepository
}

// NewItemService 创建项目业务服务
func NewItemService(itemRepo repository.ItemRepository, userRepo repository.UserRepository) ItemService {
	return &itemService{
		itemRepo: itemRepo,
		userRepo: userRepo,
	}
}

// CreateItem 创建项目
func (s *itemService) CreateItem(req *model.ItemCreateRequest, creatorID uint) (*model.Item, error) {
	if req == nil {
		return nil, errors.New("create request cannot be nil")
	}
	
	if creatorID == 0 {
		return nil, errors.New("creator ID cannot be zero")
	}
	
	// 验证必填字段
	if req.Name == "" {
		return nil, errors.New("item name is required")
	}
	
	// 验证名称长度
	if len(req.Name) > 255 {
		return nil, errors.New("item name cannot exceed 255 characters")
	}
	
	// 验证描述长度
	if len(req.Description) > 1000 {
		return nil, errors.New("description cannot exceed 1000 characters")
	}
	
	// 验证分类长度
	if len(req.Category) > 100 {
		return nil, errors.New("category cannot exceed 100 characters")
	}
	
	// 验证创建者是否存在
	_, err := s.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("creator not found: %w", err)
	}
	
	// 创建项目对象
	item := &model.Item{
		Name:        req.Name,
		Value:       req.Value,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		IsActive:    true,
		CreatedBy:   creatorID,
	}
	
	// 保存项目
	if err := s.itemRepo.Create(item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}
	
	return item, nil
}

// GetItemByID 根据ID获取项目
func (s *itemService) GetItemByID(id uint, includeCreator bool) (*model.Item, error) {
	if id == 0 {
		return nil, errors.New("invalid item ID")
	}
	
	var item *model.Item
	var err error
	
	if includeCreator {
		item, err = s.itemRepo.GetByIDWithCreator(id)
	} else {
		item, err = s.itemRepo.GetByID(id)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	
	return item, nil
}

// UpdateItem 更新项目
func (s *itemService) UpdateItem(id uint, req *model.ItemUpdateRequest, userID uint) (*model.Item, error) {
	if id == 0 {
		return nil, errors.New("invalid item ID")
	}
	
	if req == nil {
		return nil, errors.New("update request cannot be nil")
	}
	
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}
	
	// 验证项目所有权
	if err := s.ValidateItemOwnership(id, userID); err != nil {
		return nil, err
	}
	
	// 获取现有项目
	item, err := s.itemRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	
	// 更新字段
	if req.Name != nil {
		if *req.Name == "" {
			return nil, errors.New("item name cannot be empty")
		}
		if len(*req.Name) > 255 {
			return nil, errors.New("item name cannot exceed 255 characters")
		}
		item.Name = *req.Name
	}
	
	if req.Value != nil {
		item.Value = *req.Value
	}
	
	if req.Description != nil {
		if len(*req.Description) > 1000 {
			return nil, errors.New("description cannot exceed 1000 characters")
		}
		item.Description = *req.Description
	}
	
	if req.Category != nil {
		if len(*req.Category) > 100 {
			return nil, errors.New("category cannot exceed 100 characters")
		}
		item.Category = *req.Category
	}
	
	if req.Tags != nil {
		item.Tags = *req.Tags
	}
	
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}
	
	// 保存更新
	if err := s.itemRepo.Update(item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}
	
	return item, nil
}

// DeleteItem 删除项目
func (s *itemService) DeleteItem(id uint, userID uint) error {
	if id == 0 {
		return errors.New("invalid item ID")
	}
	
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	
	// 验证项目所有权
	if err := s.ValidateItemOwnership(id, userID); err != nil {
		return err
	}
	
	if err := s.itemRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	
	return nil
}

// ListItems 获取项目列表
func (s *itemService) ListItems(query *model.ItemQueryRequest, includeCreator bool) (*model.ItemListResponse, error) {
	if query == nil {
		query = &model.ItemQueryRequest{}
	}
	
	// 设置默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	
	var items []*model.Item
	var total int64
	var err error
	
	if includeCreator {
		items, total, err = s.itemRepo.ListWithCreator(query)
	} else {
		items, total, err = s.itemRepo.List(query)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	
	// 转换为响应格式
	itemResponses := make([]model.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}
	
	// 计算总页数
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	
	return &model.ItemListResponse{
		Items:      itemResponses,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchItems 搜索项目
func (s *itemService) SearchItems(keyword string, page, pageSize int) (*model.ItemListResponse, error) {
	if keyword == "" {
		return nil, errors.New("search keyword cannot be empty")
	}
	
	// 设置默认分页参数
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
	
	items, total, err := s.itemRepo.Search(keyword, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	
	// 转换为响应格式
	itemResponses := make([]model.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}
	
	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	return &model.ItemListResponse{
		Items:      itemResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetItemsByCreator 根据创建者获取项目列表
func (s *itemService) GetItemsByCreator(creatorID uint, page, pageSize int) (*model.ItemListResponse, error) {
	if creatorID == 0 {
		return nil, errors.New("invalid creator ID")
	}
	
	// 设置默认分页参数
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
	
	items, total, err := s.itemRepo.GetByCreator(creatorID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by creator: %w", err)
	}
	
	// 转换为响应格式
	itemResponses := make([]model.ItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}
	
	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	return &model.ItemListResponse{
		Items:      itemResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// SetItemActive 设置项目激活状态
func (s *itemService) SetItemActive(id uint, active bool, userID uint) error {
	if id == 0 {
		return errors.New("invalid item ID")
	}
	
	if userID == 0 {
		return errors.New("user ID cannot be zero")
	}
	
	// 验证项目所有权
	if err := s.ValidateItemOwnership(id, userID); err != nil {
		return err
	}
	
	if err := s.itemRepo.SetActive(id, active); err != nil {
		return fmt.Errorf("failed to set item active status: %w", err)
	}
	
	return nil
}

// ValidateItemOwnership 验证项目所有权
func (s *itemService) ValidateItemOwnership(itemID, userID uint) error {
	if itemID == 0 {
		return errors.New("invalid item ID")
	}
	
	if userID == 0 {
		return errors.New("invalid user ID")
	}
	
	item, err := s.itemRepo.GetByID(itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	
	if item.CreatedBy != userID {
		return errors.New("access denied: you can only modify your own items")
	}
	
	return nil
}

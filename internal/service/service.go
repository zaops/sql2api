package service

import (
	"sql2api/internal/repository"
)

// Services 服务集合
type Services struct {
	User UserService
	Item ItemService
}

// NewServices 创建服务集合
func NewServices(repos *repository.Repositories) *Services {
	return &Services{
		User: NewUserService(repos.User),
		Item: NewItemService(repos.Item, repos.User),
	}
}

// ServiceManager 服务管理器接口
type ServiceManager interface {
	GetUserService() UserService
	GetItemService() ItemService
}

// serviceManager 服务管理器实现
type serviceManager struct {
	services *Services
}

// NewServiceManager 创建服务管理器
func NewServiceManager(repos *repository.Repositories) ServiceManager {
	return &serviceManager{
		services: NewServices(repos),
	}
}

// GetUserService 获取用户服务
func (sm *serviceManager) GetUserService() UserService {
	return sm.services.User
}

// GetItemService 获取项目服务
func (sm *serviceManager) GetItemService() ItemService {
	return sm.services.Item
}

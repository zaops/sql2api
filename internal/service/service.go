package service

import (
	"fmt"

	"sql2api/internal/config"
	"sql2api/internal/repository"
)

// Services 服务集合
type Services struct {
	SQL SQLService
}

// NewServices 创建服务集合
func NewServices(repos *repository.Repositories, cfg *config.Config) (*Services, error) {
	// 创建 SQL 服务
	var sqlService SQLService
	var err error
	if cfg.SQL.Enabled {
		sqlService, err = NewSQLService(repos, &cfg.SQL)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQL service: %w", err)
		}
	}

	return &Services{
		SQL: sqlService,
	}, nil
}

// ServiceManager 服务管理器接口
type ServiceManager interface {
	GetSQLService() SQLService
}

// serviceManager 服务管理器实现
type serviceManager struct {
	services *Services
}

// NewServiceManager 创建服务管理器
func NewServiceManager(repos *repository.Repositories, cfg *config.Config) (ServiceManager, error) {
	services, err := NewServices(repos, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create services: %w", err)
	}

	return &serviceManager{
		services: services,
	}, nil
}

// GetSQLService 获取 SQL 服务
func (sm *serviceManager) GetSQLService() SQLService {
	return sm.services.SQL
}

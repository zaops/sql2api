package repository

import (
	"sql2api/internal/config"

	"gorm.io/gorm"
)

// Repositories 仓库集合
type Repositories struct {
	DB   *Database
	User UserRepository
	Item ItemRepository
}

// NewRepositories 创建仓库集合
func NewRepositories(cfg *config.DatabaseConfig) (*Repositories, error) {
	// 创建数据库连接
	db, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	// 执行数据库迁移
	if err := db.Migrate(); err != nil {
		db.Close()
		return nil, err
	}

	// 创建仓库实例
	repos := &Repositories{
		DB:   db,
		User: NewUserRepository(db.GetDB()),
		Item: NewItemRepository(db.GetDB()),
	}

	return repos, nil
}

// Close 关闭所有数据库连接
func (r *Repositories) Close() error {
	if r.DB != nil {
		return r.DB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func (r *Repositories) GetDB() *gorm.DB {
	if r.DB != nil {
		return r.DB.GetDB()
	}
	return nil
}

// HealthCheck 健康检查
func (r *Repositories) HealthCheck() error {
	if r.DB != nil {
		return r.DB.HealthCheck()
	}
	return nil
}

// Transaction 执行事务
func (r *Repositories) Transaction(fn func(*gorm.DB) error) error {
	if r.DB != nil {
		return r.DB.Transaction(fn)
	}
	return nil
}

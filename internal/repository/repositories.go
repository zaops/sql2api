package repository

import (
	"sql2api/internal/config"

	"gorm.io/gorm"
)

// Repositories 仓库集合
type Repositories struct {
	db *Database
}

// NewRepositories 创建仓库集合
func NewRepositories(cfg *config.DatabaseConfig) (*Repositories, error) {
	// 创建数据库连接
	database, err := NewDatabase(cfg)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		db: database,
	}, nil
}

// GetDB 获取数据库连接
func (r *Repositories) GetDB() *gorm.DB {
	if r.db == nil {
		return nil
	}
	return r.db.GetDB()
}

// GetDatabase 获取数据库实例
func (r *Repositories) GetDatabase() *Database {
	return r.db
}

// Close 关闭所有连接
func (r *Repositories) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// HealthCheck 健康检查
func (r *Repositories) HealthCheck() error {
	if r.db != nil {
		return r.db.HealthCheck()
	}
	return nil
}

// Migrate 执行数据库迁移
func (r *Repositories) Migrate() error {
	if r.db != nil {
		return r.db.Migrate()
	}
	return nil
}

package repository

import (
	"fmt"
	"log"
	"time"

	"sql2api/internal/config"
	"sql2api/internal/model"

	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库接口
type Database struct {
	DB     *gorm.DB
	Config *config.DatabaseConfig
}

// NewDatabase 创建数据库实例
func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	db := &Database{
		Config: cfg,
	}
	
	if err := db.Connect(); err != nil {
		return nil, err
	}
	
	return db, nil
}

// Connect 连接数据库
func (d *Database) Connect() error {
	var dialector gorm.Dialector
	
	// 根据数据库类型选择方言
	switch d.Config.Type {
	case "postgres":
		dialector = postgres.Open(d.Config.GetDSN())
	case "oracle":
		dialector = oracle.Open(d.Config.GetDSN())
	default:
		return fmt.Errorf("unsupported database type: %s", d.Config.Type)
	}
	
	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	}
	
	// 连接数据库
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 获取底层 SQL DB 实例
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// 配置连接池
	sqlDB.SetMaxOpenConns(d.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(d.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(d.Config.MaxLifetime) * time.Minute)
	
	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	
	d.DB = db
	log.Printf("Successfully connected to %s database", d.Config.Type)
	
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Migrate 执行数据库迁移
func (d *Database) Migrate() error {
	migrator := &model.DefaultModelMigrator{}
	models := migrator.GetModels()
	
	log.Println("Starting database migration...")
	
	for _, model := range models {
		if err := d.DB.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		log.Printf("Successfully migrated model: %T", model)
	}
	
	log.Println("Database migration completed successfully")
	return nil
}

// GetDB 获取 GORM DB 实例
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// HealthCheck 健康检查
func (d *Database) HealthCheck() error {
	if d.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}

// GetConnectionStats 获取连接池统计信息
func (d *Database) GetConnectionStats() (map[string]interface{}, error) {
	if d.DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	
	sqlDB, err := d.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	stats := sqlDB.Stats()
	
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// Transaction 执行事务
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

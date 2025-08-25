package repository

import (
	"testing"

	"sql2api/internal/config"

	"gorm.io/gorm"
)

func TestNewDatabase(t *testing.T) {
	// 测试配置
	cfg := &config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Username:     "test",
		Password:     "test",
		Database:     "test",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 10,
		MaxLifetime:  60,
	}

	// 注意：这个测试需要实际的数据库连接，在 CI/CD 环境中可能需要跳过
	t.Skip("Skipping database connection test - requires actual database")

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.DB == nil {
		t.Error("Expected DB to be initialized")
	}

	if db.Config != cfg {
		t.Error("Expected config to match input")
	}
}

func TestDatabase_UnsupportedType(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Type: "mysql", // 不支持的数据库类型
	}

	_, err := NewDatabase(cfg)
	if err == nil {
		t.Error("Expected error for unsupported database type")
	}

	expectedError := "unsupported database type: mysql"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDatabase_HealthCheck(t *testing.T) {
	db := &Database{}

	// 测试空连接
	err := db.HealthCheck()
	if err == nil {
		t.Error("Expected error for nil database connection")
	}

	expectedError := "database connection is nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDatabase_GetConnectionStats(t *testing.T) {
	db := &Database{}

	// 测试空连接
	_, err := db.GetConnectionStats()
	if err == nil {
		t.Error("Expected error for nil database connection")
	}

	expectedError := "database connection is nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDatabase_Close(t *testing.T) {
	db := &Database{}

	// 测试关闭空连接
	err := db.Close()
	if err != nil {
		t.Errorf("Expected no error when closing nil connection, got: %v", err)
	}
}

// 集成测试示例（需要实际数据库）
func TestDatabase_Integration(t *testing.T) {
	t.Skip("Skipping integration test - requires actual database")

	cfg := &config.DatabaseConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Username:     "postgres",
		Password:     "password",
		Database:     "sql2api_test",
		SSLMode:      "disable",
		MaxOpenConns: 5,
		MaxIdleConns: 2,
		MaxLifetime:  30,
	}

	// 创建数据库连接
	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 测试健康检查
	if err := db.HealthCheck(); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	// 测试连接统计
	stats, err := db.GetConnectionStats()
	if err != nil {
		t.Errorf("Failed to get connection stats: %v", err)
	}

	if stats == nil {
		t.Error("Expected connection stats to be returned")
	}

	// 测试迁移
	if err := db.Migrate(); err != nil {
		t.Errorf("Migration failed: %v", err)
	}

	// 测试事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些操作
		return nil
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}
}

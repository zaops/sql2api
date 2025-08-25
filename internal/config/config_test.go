package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 测试加载默认配置
	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证默认值
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", config.Server.Port)
	}

	if config.Database.Type != "postgres" {
		t.Errorf("Expected database type 'postgres', got %s", config.Database.Type)
	}

	if config.JWT.Issuer != "sql2api" {
		t.Errorf("Expected JWT issuer 'sql2api', got %s", config.JWT.Issuer)
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("SQL2API_SERVER_PORT", "9090")
	os.Setenv("SQL2API_DATABASE_TYPE", "oracle")
	defer func() {
		os.Unsetenv("SQL2API_SERVER_PORT")
		os.Unsetenv("SQL2API_DATABASE_TYPE")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证环境变量覆盖
	if config.Server.Port != 9090 {
		t.Errorf("Expected server port 9090 from env var, got %d", config.Server.Port)
	}

	if config.Database.Type != "oracle" {
		t.Errorf("Expected database type 'oracle' from env var, got %s", config.Database.Type)
	}
}

func TestGetDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "PostgreSQL DSN",
			config: DatabaseConfig{
				Type:     "postgres",
				Host:     "localhost",
				Port:     5432,
				Username: "user",
				Password: "pass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=user password=pass dbname=testdb sslmode=disable",
		},
		{
			name: "Oracle DSN with service",
			config: DatabaseConfig{
				Type:     "oracle",
				Host:     "localhost",
				Port:     1521,
				Username: "user",
				Password: "pass",
				Service:  "XE",
			},
			expected: "oracle://user:pass@localhost:1521/XE",
		},
		{
			name: "Oracle DSN with database",
			config: DatabaseConfig{
				Type:     "oracle",
				Host:     "localhost",
				Port:     1521,
				Username: "user",
				Password: "pass",
				Database: "testdb",
			},
			expected: "oracle://user:pass@localhost:1521/testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.GetDSN()
			if dsn != tt.expected {
				t.Errorf("Expected DSN %s, got %s", tt.expected, dsn)
			}
		})
	}
}

func TestGetServerAddress(t *testing.T) {
	config := ServerConfig{
		Host: "127.0.0.1",
		Port: 8080,
	}

	expected := "127.0.0.1:8080"
	address := config.GetServerAddress()
	if address != expected {
		t.Errorf("Expected server address %s, got %s", expected, address)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Database: DatabaseConfig{Type: "postgres"},
				JWT: JWTConfig{Secret: "this-is-a-very-long-secret-key-for-testing"},
				Log: LogConfig{Level: "info"},
			},
			expectErr: false,
		},
		{
			name: "Invalid database type",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Database: DatabaseConfig{Type: "mysql"},
				JWT: JWTConfig{Secret: "this-is-a-very-long-secret-key-for-testing"},
				Log: LogConfig{Level: "info"},
			},
			expectErr: true,
		},
		{
			name: "Invalid port",
			config: Config{
				Server: ServerConfig{Port: 70000},
				Database: DatabaseConfig{Type: "postgres"},
				JWT: JWTConfig{Secret: "this-is-a-very-long-secret-key-for-testing"},
				Log: LogConfig{Level: "info"},
			},
			expectErr: true,
		},
		{
			name: "Short JWT secret",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Database: DatabaseConfig{Type: "postgres"},
				JWT: JWTConfig{Secret: "short"},
				Log: LogConfig{Level: "info"},
			},
			expectErr: true,
		},
		{
			name: "Invalid log level",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Database: DatabaseConfig{Type: "postgres"},
				JWT: JWTConfig{Secret: "this-is-a-very-long-secret-key-for-testing"},
				Log: LogConfig{Level: "invalid"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

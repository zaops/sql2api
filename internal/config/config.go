package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用程序配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
	APIKeys  APIKeyConfig   `mapstructure:"api_keys"`
	SQL      SQLConfig      `mapstructure:"sql"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`          // debug, release, test
	ReadTimeout  int    `mapstructure:"read_timeout"`  // 秒
	WriteTimeout int    `mapstructure:"write_timeout"` // 秒
	IdleTimeout  int    `mapstructure:"idle_timeout"`  // 秒
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type         string `mapstructure:"type"` // postgres, oracle
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Service      string `mapstructure:"service"`  // Oracle service name
	SSLMode      string `mapstructure:"ssl_mode"` // PostgreSQL SSL mode
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"` // 分钟
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	IPWhitelist []string `mapstructure:"ip_whitelist"`
	EnableCORS  bool     `mapstructure:"enable_cors"`
	CORSOrigins []string `mapstructure:"cors_origins"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
	Output string `mapstructure:"output"` // stdout, file
	File   string `mapstructure:"file"`   // 日志文件路径
}

// SwaggerConfig Swagger 文档配置
type SwaggerConfig struct {
	Title       string   `mapstructure:"title"`        // API 标题
	Version     string   `mapstructure:"version"`      // API 版本
	Description string   `mapstructure:"description"`  // API 描述
	Host        string   `mapstructure:"host"`         // API 主机地址
	BasePath    string   `mapstructure:"base_path"`    // API 基础路径
	Schemes     []string `mapstructure:"schemes"`      // 支持的协议
}

// APIKeyConfig API Key 配置
type APIKeyConfig struct {
	Enabled     bool              `mapstructure:"enabled"`      // 是否启用 API Key 认证
	Keys        []APIKeyItem      `mapstructure:"keys"`         // API Key 列表
	HeaderName  string            `mapstructure:"header_name"`  // API Key 请求头名称
	QueryParam  string            `mapstructure:"query_param"`  // API Key 查询参数名称
	AllowAnonymous bool           `mapstructure:"allow_anonymous"` // 是否允许匿名访问
}

// APIKeyItem API Key 项目
type APIKeyItem struct {
	Key         string   `mapstructure:"key"`         // API Key 值
	Name        string   `mapstructure:"name"`        // API Key 名称
	Description string   `mapstructure:"description"` // API Key 描述
	Permissions []string `mapstructure:"permissions"` // 权限列表
	Active      bool     `mapstructure:"active"`      // 是否激活
}

// SQLConfig SQL 功能配置
type SQLConfig struct {
	Enabled            bool     `mapstructure:"enabled"`              // 是否启用 SQL 功能
	AllowedTables      []string `mapstructure:"allowed_tables"`       // 允许访问的表列表
	AllowedActions     []string `mapstructure:"allowed_actions"`      // 允许的操作类型
	MaxQueryTime       int      `mapstructure:"max_query_time"`       // 最大查询时间（秒）
	MaxResultSize      int      `mapstructure:"max_result_size"`      // 最大结果集大小（行数）
	EnableRawSQL       bool     `mapstructure:"enable_raw_sql"`       // 是否允许原生 SQL
	EnableBatch        bool     `mapstructure:"enable_batch"`         // 是否启用批量操作
	EnableTransactions bool     `mapstructure:"enable_transactions"`  // 是否启用事务支持
}

// Load 加载配置
func Load() (*Config, error) {
	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/sql2api")

	// 设置环境变量前缀
	viper.SetEnvPrefix("SQL2API")
	viper.AutomaticEnv()

	// 替换环境变量中的点号为下划线
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认配置
			fmt.Println("Warning: Config file not found, using default configuration")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	// 数据库默认配置
	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.database", "sql2api")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 60)

	// 安全默认配置
	viper.SetDefault("security.ip_whitelist", []string{"127.0.0.1", "::1"})
	viper.SetDefault("security.enable_cors", true)
	viper.SetDefault("security.cors_origins", []string{"*"})

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.file", "sql2api.log")

	// Swagger 默认配置
	viper.SetDefault("swagger.title", "SQL2API Server")
	viper.SetDefault("swagger.version", "1.0.0")
	viper.SetDefault("swagger.description", "SQL2API 是一个现代化的 RESTful API 服务，支持统一的 CRUD 操作、JWT 认证、IP 白名单等功能")
	viper.SetDefault("swagger.host", "localhost:8081")
	viper.SetDefault("swagger.base_path", "/api/v1")
	viper.SetDefault("swagger.schemes", []string{"http", "https"})

	// API Key 默认配置
	viper.SetDefault("api_keys.enabled", false)
	viper.SetDefault("api_keys.header_name", "X-API-Key")
	viper.SetDefault("api_keys.query_param", "api_key")
	viper.SetDefault("api_keys.allow_anonymous", false)
	viper.SetDefault("api_keys.keys", []APIKeyItem{})

	// SQL 功能默认配置
	viper.SetDefault("sql.enabled", true)
	viper.SetDefault("sql.allowed_tables", []string{"items"})
	viper.SetDefault("sql.allowed_actions", []string{"select", "insert", "update", "delete"})
	viper.SetDefault("sql.max_query_time", 30)
	viper.SetDefault("sql.max_result_size", 1000)
	viper.SetDefault("sql.enable_raw_sql", true)
	viper.SetDefault("sql.enable_batch", true)
	viper.SetDefault("sql.enable_transactions", true)
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证数据库类型
	if config.Database.Type != "postgres" && config.Database.Type != "oracle" {
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	// 验证服务器端口
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLevel := false
	for _, level := range validLogLevels {
		if config.Log.Level == level {
			isValidLevel = true
			break
		}
	}
	if !isValidLevel {
		return fmt.Errorf("invalid log level: %s", config.Log.Level)
	}

	// 验证 SQL 配置
	if config.SQL.Enabled {
		// 验证查询时间限制
		if config.SQL.MaxQueryTime <= 0 || config.SQL.MaxQueryTime > 300 {
			return fmt.Errorf("invalid max_query_time: %d (must be between 1 and 300 seconds)", config.SQL.MaxQueryTime)
		}

		// 验证结果集大小限制
		if config.SQL.MaxResultSize <= 0 || config.SQL.MaxResultSize > 10000 {
			return fmt.Errorf("invalid max_result_size: %d (must be between 1 and 10000 rows)", config.SQL.MaxResultSize)
		}

		// 验证允许的操作类型
		validActions := map[string]bool{
			"select": true,
			"insert": true,
			"update": true,
			"delete": true,
		}
		for _, action := range config.SQL.AllowedActions {
			if !validActions[action] {
				return fmt.Errorf("invalid SQL action: %s", action)
			}
		}
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "oracle":
		if c.Service != "" {
			return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
				c.Username, c.Password, c.Host, c.Port, c.Service)
		}
		return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

// GetServerAddress 获取服务器监听地址
func (c *ServerConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

package sql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"sql2api/internal/config"
)

// SecurityValidator SQL 安全验证器
type SecurityValidator struct {
	config           *config.SQLConfig
	allowedTables    map[string]bool
	allowedActions   map[string]bool
	dangerousKeywords []string
	sqlInjectionPatterns []*regexp.Regexp
}

// NewSecurityValidator 创建安全验证器
func NewSecurityValidator(cfg *config.SQLConfig) *SecurityValidator {
	validator := &SecurityValidator{
		config:         cfg,
		allowedTables:  make(map[string]bool),
		allowedActions: make(map[string]bool),
	}

	// 构建允许的表映射
	for _, table := range cfg.AllowedTables {
		validator.allowedTables[strings.ToLower(table)] = true
	}

	// 构建允许的操作映射
	for _, action := range cfg.AllowedActions {
		validator.allowedActions[strings.ToLower(action)] = true
	}

	// 初始化危险关键字列表
	validator.dangerousKeywords = []string{
		"drop", "truncate", "alter", "create", "grant", "revoke",
		"exec", "execute", "sp_", "xp_", "fn_", "information_schema",
		"pg_", "mysql", "sys", "master", "msdb", "tempdb",
	}

	// 初始化 SQL 注入检测模式
	validator.initSQLInjectionPatterns()

	return validator
}

// ValidateQuery 验证查询的安全性
func (v *SecurityValidator) ValidateQuery(query string, params map[string]interface{}) error {
	if query == "" {
		return errors.New("query cannot be empty")
	}

	// 清理查询字符串
	cleanQuery := strings.TrimSpace(strings.ToLower(query))

	// 检查 SQL 注入
	if err := v.checkSQLInjection(query); err != nil {
		return fmt.Errorf("SQL injection detected: %w", err)
	}

	// 检查危险关键字
	if err := v.checkDangerousKeywords(cleanQuery); err != nil {
		return fmt.Errorf("dangerous keywords detected: %w", err)
	}

	// 检查操作权限
	if err := v.checkActionPermission(cleanQuery); err != nil {
		return fmt.Errorf("action not allowed: %w", err)
	}

	// 检查表访问权限
	if err := v.checkTablePermission(cleanQuery); err != nil {
		return fmt.Errorf("table access denied: %w", err)
	}

	// 验证参数
	if err := v.validateParameters(params); err != nil {
		return fmt.Errorf("parameter validation failed: %w", err)
	}

	return nil
}

// IsSelectQuery 检查是否为查询操作
func (v *SecurityValidator) IsSelectQuery(query string) bool {
	cleanQuery := strings.TrimSpace(strings.ToLower(query))
	return strings.HasPrefix(cleanQuery, "select")
}

// checkSQLInjection 检查 SQL 注入
func (v *SecurityValidator) checkSQLInjection(query string) error {
	for _, pattern := range v.sqlInjectionPatterns {
		if pattern.MatchString(query) {
			return fmt.Errorf("potential SQL injection pattern detected")
		}
	}
	return nil
}

// checkDangerousKeywords 检查危险关键字
func (v *SecurityValidator) checkDangerousKeywords(query string) error {
	for _, keyword := range v.dangerousKeywords {
		if strings.Contains(query, keyword) {
			return fmt.Errorf("dangerous keyword '%s' not allowed", keyword)
		}
	}
	return nil
}

// checkActionPermission 检查操作权限
func (v *SecurityValidator) checkActionPermission(query string) error {
	// 提取 SQL 操作类型
	action := v.extractSQLAction(query)
	if action == "" {
		return errors.New("unable to determine SQL action")
	}

	// 检查是否允许该操作
	if !v.allowedActions[action] {
		return fmt.Errorf("action '%s' is not allowed", action)
	}

	return nil
}

// checkTablePermission 检查表访问权限
func (v *SecurityValidator) checkTablePermission(query string) error {
	// 提取查询中涉及的表名
	tables := v.extractTableNames(query)
	if len(tables) == 0 {
		return errors.New("no tables found in query")
	}

	// 检查每个表是否在白名单中
	for _, table := range tables {
		if !v.allowedTables[strings.ToLower(table)] {
			return fmt.Errorf("access to table '%s' is not allowed", table)
		}
	}

	return nil
}

// validateParameters 验证参数
func (v *SecurityValidator) validateParameters(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	// 检查参数数量限制
	if len(params) > 100 {
		return errors.New("too many parameters (max: 100)")
	}

	// 检查参数值
	for key, value := range params {
		if err := v.validateParameterValue(key, value); err != nil {
			return fmt.Errorf("invalid parameter '%s': %w", key, err)
		}
	}

	return nil
}

// validateParameterValue 验证单个参数值
func (v *SecurityValidator) validateParameterValue(key string, value interface{}) error {
	if key == "" {
		return errors.New("parameter key cannot be empty")
	}

	// 检查字符串参数的长度和内容
	if str, ok := value.(string); ok {
		if len(str) > 10000 {
			return errors.New("string parameter too long (max: 10000 characters)")
		}

		// 检查是否包含潜在的 SQL 注入内容
		if v.containsSQLInjection(str) {
			return errors.New("parameter contains potential SQL injection")
		}
	}

	return nil
}

// containsSQLInjection 检查字符串是否包含 SQL 注入
func (v *SecurityValidator) containsSQLInjection(str string) bool {
	dangerous := []string{
		"'", "\"", ";", "--", "/*", "*/", "union", "select", "insert", "update", "delete",
		"drop", "create", "alter", "exec", "execute", "script", "javascript",
	}

	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerous {
		if strings.Contains(lowerStr, pattern) {
			return true
		}
	}
	return false
}

// extractSQLAction 提取 SQL 操作类型
func (v *SecurityValidator) extractSQLAction(query string) string {
	words := strings.Fields(query)
	if len(words) == 0 {
		return ""
	}

	firstWord := strings.ToLower(words[0])
	switch firstWord {
	case "select":
		return "select"
	case "insert":
		return "insert"
	case "update":
		return "update"
	case "delete":
		return "delete"
	case "with": // CTE 查询
		// 查找 CTE 后的实际操作
		for i, word := range words {
			lowerWord := strings.ToLower(word)
			if lowerWord == "select" || lowerWord == "insert" || lowerWord == "update" || lowerWord == "delete" {
				return lowerWord
			}
			if i > 10 { // 避免无限循环
				break
			}
		}
		return "select" // 默认为查询
	default:
		return firstWord
	}
}

// extractTableNames 提取查询中的表名
func (v *SecurityValidator) extractTableNames(query string) []string {
	var tables []string

	// 简单的表名提取逻辑
	// 实际应用中可能需要更复杂的 SQL 解析
	words := strings.Fields(query)
	
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		
		// 查找 FROM、JOIN、INTO、UPDATE 等关键字后的表名
		if (lowerWord == "from" || lowerWord == "join" || lowerWord == "into" || lowerWord == "update") && i+1 < len(words) {
			tableName := strings.ToLower(words[i+1])
			// 清理表名（移除标点符号）
			tableName = strings.Trim(tableName, "(),;")
			if tableName != "" && !v.isKeyword(tableName) {
				tables = append(tables, tableName)
			}
		}
	}

	return tables
}

// isKeyword 检查是否为 SQL 关键字
func (v *SecurityValidator) isKeyword(word string) bool {
	keywords := map[string]bool{
		"select": true, "from": true, "where": true, "and": true, "or": true,
		"order": true, "by": true, "group": true, "having": true, "limit": true,
		"offset": true, "join": true, "inner": true, "left": true, "right": true,
		"full": true, "outer": true, "on": true, "as": true, "distinct": true,
		"count": true, "sum": true, "avg": true, "max": true, "min": true,
	}
	return keywords[strings.ToLower(word)]
}

// initSQLInjectionPatterns 初始化 SQL 注入检测模式
func (v *SecurityValidator) initSQLInjectionPatterns() {
	patterns := []string{
		`(?i)(\s|^)(union\s+select)`,                    // UNION SELECT 注入
		`(?i)(\s|^)(or\s+1\s*=\s*1)`,                   // OR 1=1 注入
		`(?i)(\s|^)(and\s+1\s*=\s*1)`,                  // AND 1=1 注入
		`(?i)(\s|^)(or\s+\w+\s*=\s*\w+)`,              // OR 字段=字段 注入
		`(?i)(\s|^)(and\s+\w+\s*=\s*\w+)`,             // AND 字段=字段 注入
		`(?i)(;|\s)(drop|create|alter|truncate)\s+`,     // DDL 语句
		`(?i)(exec|execute|sp_|xp_)`,                    // 存储过程执行
		`(?i)(script|javascript|vbscript)`,              // 脚本注入
		`(?i)(information_schema|sys\.|pg_|mysql\.)`,    // 系统表访问
	}

	v.sqlInjectionPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			v.sqlInjectionPatterns = append(v.sqlInjectionPatterns, regex)
		}
	}
}

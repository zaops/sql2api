package sql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// QueryValidator 查询验证器
type QueryValidator struct {
	maxQueryLength    int
	maxComplexity     int
	allowedFunctions  map[string]bool
	forbiddenPatterns []*regexp.Regexp
}

// NewQueryValidator 创建查询验证器
func NewQueryValidator() *QueryValidator {
	validator := &QueryValidator{
		maxQueryLength: 10000, // 最大查询长度
		maxComplexity:  100,   // 最大查询复杂度
		allowedFunctions: map[string]bool{
			// 聚合函数
			"count": true, "sum": true, "avg": true, "max": true, "min": true,
			// 字符串函数
			"upper": true, "lower": true, "trim": true, "length": true, "substr": true,
			"substring": true, "concat": true, "replace": true, "like": true,
			// 数学函数
			"abs": true, "round": true, "ceil": true, "floor": true, "mod": true,
			// 日期函数
			"now": true, "current_date": true, "current_time": true, "current_timestamp": true,
			"date": true, "time": true, "year": true, "month": true, "day": true,
			// 条件函数
			"case": true, "when": true, "then": true, "else": true, "end": true,
			"coalesce": true, "nullif": true, "isnull": true, "ifnull": true,
		},
	}

	// 初始化禁止的模式
	validator.initForbiddenPatterns()

	return validator
}

// ValidateQueryStructure 验证查询结构
func (v *QueryValidator) ValidateQueryStructure(query string) error {
	if query == "" {
		return errors.New("query cannot be empty")
	}

	// 检查查询长度
	if len(query) > v.maxQueryLength {
		return fmt.Errorf("query too long: %d characters (max: %d)", len(query), v.maxQueryLength)
	}

	// 检查查询复杂度
	complexity := v.calculateQueryComplexity(query)
	if complexity > v.maxComplexity {
		return fmt.Errorf("query too complex: complexity %d (max: %d)", complexity, v.maxComplexity)
	}

	// 检查禁止的模式
	if err := v.checkForbiddenPatterns(query); err != nil {
		return fmt.Errorf("forbidden pattern detected: %w", err)
	}

	// 检查函数使用
	if err := v.validateFunctions(query); err != nil {
		return fmt.Errorf("function validation failed: %w", err)
	}

	// 检查语法基本结构
	if err := v.validateBasicSyntax(query); err != nil {
		return fmt.Errorf("syntax validation failed: %w", err)
	}

	return nil
}

// calculateQueryComplexity 计算查询复杂度
func (v *QueryValidator) calculateQueryComplexity(query string) int {
	complexity := 0
	lowerQuery := strings.ToLower(query)

	// 基础复杂度
	complexity += 1

	// JOIN 操作增加复杂度
	complexity += strings.Count(lowerQuery, "join") * 5
	complexity += strings.Count(lowerQuery, "left join") * 3
	complexity += strings.Count(lowerQuery, "right join") * 3
	complexity += strings.Count(lowerQuery, "inner join") * 3
	complexity += strings.Count(lowerQuery, "outer join") * 4

	// 子查询增加复杂度
	complexity += strings.Count(lowerQuery, "select") * 3

	// WHERE 条件增加复杂度
	complexity += strings.Count(lowerQuery, "where") * 2
	complexity += strings.Count(lowerQuery, "and") * 1
	complexity += strings.Count(lowerQuery, "or") * 2

	// GROUP BY 和 ORDER BY 增加复杂度
	complexity += strings.Count(lowerQuery, "group by") * 3
	complexity += strings.Count(lowerQuery, "order by") * 2
	complexity += strings.Count(lowerQuery, "having") * 3

	// 聚合函数增加复杂度
	complexity += strings.Count(lowerQuery, "count(") * 2
	complexity += strings.Count(lowerQuery, "sum(") * 2
	complexity += strings.Count(lowerQuery, "avg(") * 2
	complexity += strings.Count(lowerQuery, "max(") * 2
	complexity += strings.Count(lowerQuery, "min(") * 2

	// UNION 操作增加复杂度
	complexity += strings.Count(lowerQuery, "union") * 5

	// CTE (WITH) 增加复杂度
	complexity += strings.Count(lowerQuery, "with") * 4

	return complexity
}

// checkForbiddenPatterns 检查禁止的模式
func (v *QueryValidator) checkForbiddenPatterns(query string) error {
	for _, pattern := range v.forbiddenPatterns {
		if pattern.MatchString(query) {
			return fmt.Errorf("query contains forbidden pattern")
		}
	}
	return nil
}

// validateFunctions 验证函数使用
func (v *QueryValidator) validateFunctions(query string) error {
	// 提取查询中使用的函数
	functions := v.extractFunctions(query)

	// 检查每个函数是否被允许
	for _, function := range functions {
		if !v.allowedFunctions[strings.ToLower(function)] {
			return fmt.Errorf("function '%s' is not allowed", function)
		}
	}

	return nil
}

// extractFunctions 提取查询中的函数
func (v *QueryValidator) extractFunctions(query string) []string {
	// 使用正则表达式匹配函数调用模式
	funcPattern := regexp.MustCompile(`(\w+)\s*\(`)
	matches := funcPattern.FindAllStringSubmatch(query, -1)

	var functions []string
	for _, match := range matches {
		if len(match) > 1 {
			funcName := strings.ToLower(match[1])
			// 过滤掉 SQL 关键字
			if !v.isSQLKeyword(funcName) {
				functions = append(functions, funcName)
			}
		}
	}

	return functions
}

// isSQLKeyword 检查是否为 SQL 关键字
func (v *QueryValidator) isSQLKeyword(word string) bool {
	keywords := map[string]bool{
		"select": true, "from": true, "where": true, "and": true, "or": true,
		"order": true, "by": true, "group": true, "having": true, "limit": true,
		"offset": true, "join": true, "inner": true, "left": true, "right": true,
		"full": true, "outer": true, "on": true, "as": true, "distinct": true,
		"union": true, "all": true, "exists": true, "in": true, "not": true,
		"is": true, "null": true, "like": true, "between": true, "case": true,
		"when": true, "then": true, "else": true, "end": true, "with": true,
	}
	return keywords[strings.ToLower(word)]
}

// validateBasicSyntax 验证基本语法
func (v *QueryValidator) validateBasicSyntax(query string) error {
	// 检查括号匹配
	if err := v.checkParenthesesBalance(query); err != nil {
		return fmt.Errorf("parentheses mismatch: %w", err)
	}

	// 检查引号匹配
	if err := v.checkQuotesBalance(query); err != nil {
		return fmt.Errorf("quotes mismatch: %w", err)
	}

	// 检查基本的 SQL 结构
	if err := v.checkBasicSQLStructure(query); err != nil {
		return fmt.Errorf("invalid SQL structure: %w", err)
	}

	return nil
}

// checkParenthesesBalance 检查括号平衡
func (v *QueryValidator) checkParenthesesBalance(query string) error {
	count := 0
	for _, char := range query {
		switch char {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return errors.New("unmatched closing parenthesis")
			}
		}
	}

	if count != 0 {
		return errors.New("unmatched opening parenthesis")
	}

	return nil
}

// checkQuotesBalance 检查引号平衡
func (v *QueryValidator) checkQuotesBalance(query string) error {
	singleQuoteCount := 0
	doubleQuoteCount := 0
	inSingleQuote := false
	inDoubleQuote := false

	for i, char := range query {
		switch char {
		case '\'':
			if !inDoubleQuote {
				if i > 0 && query[i-1] == '\\' {
					continue // 转义的引号
				}
				inSingleQuote = !inSingleQuote
				singleQuoteCount++
			}
		case '"':
			if !inSingleQuote {
				if i > 0 && query[i-1] == '\\' {
					continue // 转义的引号
				}
				inDoubleQuote = !inDoubleQuote
				doubleQuoteCount++
			}
		}
	}

	if singleQuoteCount%2 != 0 {
		return errors.New("unmatched single quote")
	}

	if doubleQuoteCount%2 != 0 {
		return errors.New("unmatched double quote")
	}

	return nil
}

// checkBasicSQLStructure 检查基本 SQL 结构
func (v *QueryValidator) checkBasicSQLStructure(query string) error {
	lowerQuery := strings.ToLower(strings.TrimSpace(query))

	// 检查是否以有效的 SQL 关键字开始
	validStarters := []string{"select", "insert", "update", "delete", "with"}
	hasValidStarter := false
	for _, starter := range validStarters {
		if strings.HasPrefix(lowerQuery, starter) {
			hasValidStarter = true
			break
		}
	}

	if !hasValidStarter {
		return errors.New("query must start with a valid SQL keyword")
	}

	// 对于 SELECT 查询，检查是否有 FROM 子句（除非是简单的表达式查询）
	if strings.HasPrefix(lowerQuery, "select") {
		// 检查是否包含表名或 FROM 子句
		if !strings.Contains(lowerQuery, "from") && !v.isSimpleExpressionQuery(lowerQuery) {
			return errors.New("SELECT query must include FROM clause or be a simple expression")
		}
	}

	return nil
}

// isSimpleExpressionQuery 检查是否为简单表达式查询
func (v *QueryValidator) isSimpleExpressionQuery(query string) bool {
	// 简单表达式查询示例：SELECT 1, SELECT NOW(), SELECT 'hello'
	// 这些查询不需要 FROM 子句
	simplePatterns := []string{
		"select\\s+\\d+",           // SELECT 数字
		"select\\s+now\\(\\)",      // SELECT NOW()
		"select\\s+'[^']*'",        // SELECT '字符串'
		"select\\s+\"[^\"]*\"",     // SELECT "字符串"
		"select\\s+current_",       // SELECT CURRENT_*
	}

	for _, pattern := range simplePatterns {
		if matched, _ := regexp.MatchString(pattern, query); matched {
			return true
		}
	}

	return false
}

// initForbiddenPatterns 初始化禁止的模式
func (v *QueryValidator) initForbiddenPatterns() {
	patterns := []string{
		`(?i)(benchmark|sleep|waitfor|delay)\s*\(`,     // 时间延迟攻击
		`(?i)(load_file|into\s+outfile|into\s+dumpfile)`, // 文件操作
		`(?i)(@@|@\w+)`,                                // 系统变量访问
		`(?i)(char|ascii|hex|unhex|bin|oct)\s*\(`,      // 编码函数（可能用于绕过）
		`(?i)(version|user|database|schema)\s*\(\s*\)`, // 系统信息函数
		`(?i)(concat_ws|group_concat|string_agg)\s*\(`, // 可能用于数据泄露的函数
		`(?i)(if|ifnull|nullif|case)\s*\(.*select`,     // 条件语句中的子查询
		`(?i)(\|\||&&|<<|>>)`,                          // 位运算符（可能用于绕过）
	}

	v.forbiddenPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			v.forbiddenPatterns = append(v.forbiddenPatterns, regex)
		}
	}
}

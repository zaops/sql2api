package sql

import (
	"regexp"
	"strings"

	"sql2api/internal/model"
)

// DatabaseErrorMapper 数据库错误映射器
type DatabaseErrorMapper struct {
	dbType string
}

// NewDatabaseErrorMapper 创建数据库错误映射器
func NewDatabaseErrorMapper(dbType string) *DatabaseErrorMapper {
	return &DatabaseErrorMapper{
		dbType: dbType,
	}
}

// MapError 将数据库错误映射为 SQL 错误
func (m *DatabaseErrorMapper) MapError(err error) *model.SQLError {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())
	
	switch m.dbType {
	case "postgres":
		return m.mapPostgreSQLError(errMsg, err.Error())
	case "oracle":
		return m.mapOracleError(errMsg, err.Error())
	default:
		return m.mapGenericError(errMsg, err.Error())
	}
}

// mapPostgreSQLError 映射 PostgreSQL 错误
func (m *DatabaseErrorMapper) mapPostgreSQLError(errMsg, originalErr string) *model.SQLError {
	// PostgreSQL 错误码模式
	patterns := map[*regexp.Regexp]model.SQLError{
		// 语法错误
		regexp.MustCompile(`syntax error|invalid syntax|parse error`): {
			Code:    model.SQLErrorSyntax,
			Message: "SQL syntax error",
		},
		// 权限错误
		regexp.MustCompile(`permission denied|access denied|insufficient privilege`): {
			Code:    model.SQLErrorPermission,
			Message: "Permission denied",
		},
		// 连接错误
		regexp.MustCompile(`connection|connect|network|timeout|unreachable`): {
			Code:    model.SQLErrorConnection,
			Message: "Database connection error",
		},
		// 参数错误
		regexp.MustCompile(`invalid parameter|missing parameter|parameter.*required`): {
			Code:    model.SQLErrorParams,
			Message: "Parameter error",
		},
		// 事务错误
		regexp.MustCompile(`transaction|deadlock|serialization failure`): {
			Code:    model.SQLErrorTransaction,
			Message: "Transaction error",
		},
		// 超时错误
		regexp.MustCompile(`timeout|time.*out|deadline exceeded`): {
			Code:    model.SQLErrorTimeout,
			Message: "Query timeout",
		},
	}

	// PostgreSQL 特定错误码
	pgErrorCodes := map[string]model.SQLError{
		"42601": {Code: model.SQLErrorSyntax, Message: "Syntax error"},
		"42501": {Code: model.SQLErrorPermission, Message: "Insufficient privilege"},
		"08000": {Code: model.SQLErrorConnection, Message: "Connection exception"},
		"08003": {Code: model.SQLErrorConnection, Message: "Connection does not exist"},
		"08006": {Code: model.SQLErrorConnection, Message: "Connection failure"},
		"40001": {Code: model.SQLErrorTransaction, Message: "Serialization failure"},
		"40P01": {Code: model.SQLErrorTransaction, Message: "Deadlock detected"},
		"57014": {Code: model.SQLErrorTimeout, Message: "Query canceled"},
	}

	// 检查 PostgreSQL 错误码
	for code, sqlErr := range pgErrorCodes {
		if strings.Contains(errMsg, code) {
			sqlErr.Details = originalErr
			return &sqlErr
		}
	}

	// 检查错误模式
	for pattern, sqlErr := range patterns {
		if pattern.MatchString(errMsg) {
			sqlErr.Details = originalErr
			return &sqlErr
		}
	}

	// 默认错误
	return &model.SQLError{
		Code:    model.SQLErrorSyntax,
		Message: "Database error",
		Details: originalErr,
	}
}

// mapOracleError 映射 Oracle 错误
func (m *DatabaseErrorMapper) mapOracleError(errMsg, originalErr string) *model.SQLError {
	// Oracle 错误码模式
	oracleErrorCodes := map[string]model.SQLError{
		"ora-00900": {Code: model.SQLErrorSyntax, Message: "Invalid SQL statement"},
		"ora-00901": {Code: model.SQLErrorSyntax, Message: "Invalid CREATE command"},
		"ora-00902": {Code: model.SQLErrorSyntax, Message: "Invalid datatype"},
		"ora-00903": {Code: model.SQLErrorSyntax, Message: "Invalid table name"},
		"ora-00904": {Code: model.SQLErrorSyntax, Message: "Invalid identifier"},
		"ora-00942": {Code: model.SQLErrorPermission, Message: "Table or view does not exist"},
		"ora-00955": {Code: model.SQLErrorSyntax, Message: "Name is already used by an existing object"},
		"ora-01017": {Code: model.SQLErrorPermission, Message: "Invalid username/password"},
		"ora-01031": {Code: model.SQLErrorPermission, Message: "Insufficient privileges"},
		"ora-01034": {Code: model.SQLErrorConnection, Message: "Oracle not available"},
		"ora-01089": {Code: model.SQLErrorConnection, Message: "Immediate shutdown in progress"},
		"ora-03113": {Code: model.SQLErrorConnection, Message: "End-of-file on communication channel"},
		"ora-03114": {Code: model.SQLErrorConnection, Message: "Not connected to Oracle"},
		"ora-12154": {Code: model.SQLErrorConnection, Message: "TNS: could not resolve the connect identifier"},
		"ora-12170": {Code: model.SQLErrorTimeout, Message: "TNS: connect timeout occurred"},
		"ora-12571": {Code: model.SQLErrorTimeout, Message: "TNS: packet writer failure"},
		"ora-00060": {Code: model.SQLErrorTransaction, Message: "Deadlock detected while waiting for resource"},
		"ora-08177": {Code: model.SQLErrorTransaction, Message: "Cannot serialize access for this transaction"},
	}

	// 检查 Oracle 错误码
	for code, sqlErr := range oracleErrorCodes {
		if strings.Contains(errMsg, code) {
			sqlErr.Details = originalErr
			return &sqlErr
		}
	}

	// 通用模式检查
	patterns := map[*regexp.Regexp]model.SQLError{
		regexp.MustCompile(`syntax|invalid.*statement|parse`): {
			Code:    model.SQLErrorSyntax,
			Message: "SQL syntax error",
		},
		regexp.MustCompile(`privilege|permission|access.*denied`): {
			Code:    model.SQLErrorPermission,
			Message: "Permission denied",
		},
		regexp.MustCompile(`connection|connect|tns|network`): {
			Code:    model.SQLErrorConnection,
			Message: "Database connection error",
		},
		regexp.MustCompile(`timeout|time.*out`): {
			Code:    model.SQLErrorTimeout,
			Message: "Query timeout",
		},
		regexp.MustCompile(`deadlock|transaction|serialize`): {
			Code:    model.SQLErrorTransaction,
			Message: "Transaction error",
		},
	}

	for pattern, sqlErr := range patterns {
		if pattern.MatchString(errMsg) {
			sqlErr.Details = originalErr
			return &sqlErr
		}
	}

	// 默认错误
	return &model.SQLError{
		Code:    model.SQLErrorSyntax,
		Message: "Database error",
		Details: originalErr,
	}
}

// mapGenericError 映射通用数据库错误
func (m *DatabaseErrorMapper) mapGenericError(errMsg, originalErr string) *model.SQLError {
	patterns := map[*regexp.Regexp]model.SQLError{
		regexp.MustCompile(`syntax|parse|invalid.*sql`): {
			Code:    model.SQLErrorSyntax,
			Message: "SQL syntax error",
		},
		regexp.MustCompile(`permission|privilege|access.*denied|unauthorized`): {
			Code:    model.SQLErrorPermission,
			Message: "Permission denied",
		},
		regexp.MustCompile(`connection|connect|network|unreachable`): {
			Code:    model.SQLErrorConnection,
			Message: "Database connection error",
		},
		regexp.MustCompile(`parameter|argument|missing.*value`): {
			Code:    model.SQLErrorParams,
			Message: "Parameter error",
		},
		regexp.MustCompile(`timeout|time.*out|deadline`): {
			Code:    model.SQLErrorTimeout,
			Message: "Query timeout",
		},
		regexp.MustCompile(`transaction|deadlock|lock.*timeout`): {
			Code:    model.SQLErrorTransaction,
			Message: "Transaction error",
		},
		regexp.MustCompile(`too.*large|limit.*exceeded|memory`): {
			Code:    model.SQLErrorResultSize,
			Message: "Result set too large",
		},
	}

	for pattern, sqlErr := range patterns {
		if pattern.MatchString(errMsg) {
			sqlErr.Details = originalErr
			return &sqlErr
		}
	}

	// 默认错误
	return &model.SQLError{
		Code:    model.SQLErrorSyntax,
		Message: "Database error",
		Details: originalErr,
	}
}

// GetSQLState 获取数据库特定的 SQL 状态码
func (m *DatabaseErrorMapper) GetSQLState(err error) string {
	if err == nil {
		return ""
	}

	errMsg := strings.ToLower(err.Error())

	switch m.dbType {
	case "postgres":
		// PostgreSQL SQLSTATE 提取
		if match := regexp.MustCompile(`sqlstate:\s*([0-9a-z]{5})`).FindStringSubmatch(errMsg); len(match) > 1 {
			return match[1]
		}
	case "oracle":
		// Oracle 错误码提取
		if match := regexp.MustCompile(`ora-(\d{5})`).FindStringSubmatch(errMsg); len(match) > 1 {
			return "ORA-" + match[1]
		}
	}

	return ""
}

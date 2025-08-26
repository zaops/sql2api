package sql

import (
	"fmt"
	"strings"

	"sql2api/internal/model"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	dbType  string
	dialect DatabaseDialect
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(dbType string) *QueryBuilder {
	factory := NewDialectFactory()
	dialect := factory.CreateDialect(dbType)
	
	return &QueryBuilder{
		dbType:  dbType,
		dialect: dialect,
	}
}

// BuildStructuredQuery 构建结构化查询
func (b *QueryBuilder) BuildStructuredQuery(query *model.StructuredQuery) (string, map[string]interface{}, error) {
	switch strings.ToLower(query.Action) {
	case "select":
		return b.buildSelectQuery(query)
	case "insert":
		return b.buildInsertQuery(query)
	case "update":
		return b.buildUpdateQuery(query)
	case "delete":
		return b.buildDeleteQuery(query)
	default:
		return "", nil, fmt.Errorf("unsupported action: %s", query.Action)
	}
}

// buildSelectQuery 构建 SELECT 查询
func (b *QueryBuilder) buildSelectQuery(query *model.StructuredQuery) (string, map[string]interface{}, error) {
	var sql strings.Builder
	params := make(map[string]interface{})
	paramIndex := 1
	
	// SELECT 子句
	sql.WriteString("SELECT ")
	if len(query.Fields) > 0 {
		sql.WriteString(strings.Join(query.Fields, ", "))
	} else {
		sql.WriteString("*")
	}
	
	// FROM 子句
	sql.WriteString(" FROM ")
	sql.WriteString(query.Table)
	
	// WHERE 子句
	if len(query.Where) > 0 {
		whereClause, whereParams, err := b.buildWhereClause(query.Where, paramIndex)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build WHERE clause: %w", err)
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
		
		// 合并参数
		for k, v := range whereParams {
			params[k] = v
		}
		paramIndex += len(whereParams)
	}
	
	// GROUP BY 子句
	if len(query.GroupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(query.GroupBy, ", "))
	}
	
	// HAVING 子句
	if len(query.Having) > 0 {
		havingClause, havingParams, err := b.buildWhereClause(query.Having, paramIndex)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build HAVING clause: %w", err)
		}
		sql.WriteString(" HAVING ")
		sql.WriteString(havingClause)
		
		// 合并参数
		for k, v := range havingParams {
			params[k] = v
		}
		paramIndex += len(havingParams)
	}
	
	// ORDER BY 子句
	if len(query.OrderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		var orderClauses []string
		for _, orderBy := range query.OrderBy {
			order := "ASC"
			if strings.ToUpper(orderBy.Order) == "DESC" {
				order = "DESC"
			}
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", orderBy.Field, order))
		}
		sql.WriteString(strings.Join(orderClauses, ", "))
	}
	
	// LIMIT 子句
	if query.Limit > 0 {
		sql.WriteString(" ")
		sql.WriteString(b.dialect.GetLimitQuery(query.Limit))
	}
	
	return sql.String(), params, nil
}

// buildInsertQuery 构建 INSERT 查询
func (b *QueryBuilder) buildInsertQuery(query *model.StructuredQuery) (string, map[string]interface{}, error) {
	if len(query.Data) == 0 {
		return "", nil, fmt.Errorf("no data provided for insert")
	}
	
	var sql strings.Builder
	params := make(map[string]interface{})
	
	// INSERT INTO 子句
	sql.WriteString("INSERT INTO ")
	sql.WriteString(query.Table)
	
	// 字段列表
	var fields []string
	var placeholders []string
	paramIndex := 1
	
	for field, value := range query.Data {
		fields = append(fields, field)
		placeholder := b.getParameterPlaceholder(paramIndex)
		placeholders = append(placeholders, placeholder)
		params[fmt.Sprintf("param_%d", paramIndex)] = value
		paramIndex++
	}
	
	sql.WriteString(" (")
	sql.WriteString(strings.Join(fields, ", "))
	sql.WriteString(") VALUES (")
	sql.WriteString(strings.Join(placeholders, ", "))
	sql.WriteString(")")
	
	return sql.String(), params, nil
}

// buildUpdateQuery 构建 UPDATE 查询
func (b *QueryBuilder) buildUpdateQuery(query *model.StructuredQuery) (string, map[string]interface{}, error) {
	if len(query.Data) == 0 {
		return "", nil, fmt.Errorf("no data provided for update")
	}
	
	var sql strings.Builder
	params := make(map[string]interface{})
	paramIndex := 1
	
	// UPDATE 子句
	sql.WriteString("UPDATE ")
	sql.WriteString(query.Table)
	sql.WriteString(" SET ")
	
	// SET 子句
	var setClauses []string
	for field, value := range query.Data {
		placeholder := b.getParameterPlaceholder(paramIndex)
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", field, placeholder))
		params[fmt.Sprintf("param_%d", paramIndex)] = value
		paramIndex++
	}
	sql.WriteString(strings.Join(setClauses, ", "))
	
	// WHERE 子句
	if len(query.Where) > 0 {
		whereClause, whereParams, err := b.buildWhereClause(query.Where, paramIndex)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build WHERE clause: %w", err)
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
		
		// 合并参数
		for k, v := range whereParams {
			params[k] = v
		}
	}
	
	return sql.String(), params, nil
}

// buildDeleteQuery 构建 DELETE 查询
func (b *QueryBuilder) buildDeleteQuery(query *model.StructuredQuery) (string, map[string]interface{}, error) {
	var sql strings.Builder
	params := make(map[string]interface{})
	
	// DELETE FROM 子句
	sql.WriteString("DELETE FROM ")
	sql.WriteString(query.Table)
	
	// WHERE 子句
	if len(query.Where) > 0 {
		whereClause, whereParams, err := b.buildWhereClause(query.Where, 1)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build WHERE clause: %w", err)
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
		
		// 合并参数
		for k, v := range whereParams {
			params[k] = v
		}
	} else {
		return "", nil, fmt.Errorf("WHERE clause is required for DELETE operation")
	}
	
	return sql.String(), params, nil
}

// buildWhereClause 构建 WHERE 子句
func (b *QueryBuilder) buildWhereClause(conditions map[string]interface{}, startIndex int) (string, map[string]interface{}, error) {
	if len(conditions) == 0 {
		return "", nil, nil
	}
	
	var clauses []string
	params := make(map[string]interface{})
	paramIndex := startIndex
	
	for field, value := range conditions {
		placeholder := b.getParameterPlaceholder(paramIndex)
		clauses = append(clauses, fmt.Sprintf("%s = %s", field, placeholder))
		params[fmt.Sprintf("param_%d", paramIndex)] = value
		paramIndex++
	}
	
	return strings.Join(clauses, " AND "), params, nil
}

// getParameterPlaceholder 获取参数占位符
func (b *QueryBuilder) getParameterPlaceholder(index int) string {
	switch b.dbType {
	case "postgres":
		return fmt.Sprintf("$%d", index)
	case "oracle":
		return fmt.Sprintf(":param_%d", index)
	default:
		return "?"
	}
}

// ApplyPagination 应用分页
func (b *QueryBuilder) ApplyPagination(query string, offset, limit int) string {
	return b.dialect.ApplyPagination(query, offset, limit)
}

// ApplySort 应用排序
func (b *QueryBuilder) ApplySort(query string, sortBy, sortOrder string) string {
	return b.dialect.ApplySort(query, sortBy, sortOrder)
}

// BuildInsertQuery 构建插入查询（用于 InsertRequest）
func (b *QueryBuilder) BuildInsertQuery(req *model.InsertRequest) (string, map[string]interface{}, error) {
	if len(req.Data) == 0 {
		return "", nil, fmt.Errorf("no data provided for insert")
	}

	var sql strings.Builder
	params := make(map[string]interface{})

	// INSERT INTO 子句
	sql.WriteString("INSERT INTO ")
	sql.WriteString(req.Table)

	// 字段列表
	var fields []string
	var placeholders []string
	paramIndex := 1

	for field, value := range req.Data {
		fields = append(fields, field)
		placeholder := b.getParameterPlaceholder(paramIndex)
		placeholders = append(placeholders, placeholder)
		params[fmt.Sprintf("param_%d", paramIndex)] = value
		paramIndex++
	}

	sql.WriteString(" (")
	sql.WriteString(strings.Join(fields, ", "))
	sql.WriteString(") VALUES (")
	sql.WriteString(strings.Join(placeholders, ", "))
	sql.WriteString(")")

	// 处理冲突
	if req.OnConflict != "" {
		conflictClause := b.buildOnConflictClause(req.OnConflict, fields)
		sql.WriteString(conflictClause)
	}

	// 返回字段
	if len(req.ReturnFields) > 0 {
		returnClause := b.buildReturningClause(req.ReturnFields)
		sql.WriteString(returnClause)
	}

	return sql.String(), params, nil
}

// BuildBatchInsertQuery 构建批量插入查询
func (b *QueryBuilder) BuildBatchInsertQuery(req *model.BatchInsertRequest) (string, map[string]interface{}, error) {
	if len(req.Data) == 0 {
		return "", nil, fmt.Errorf("no data provided for batch insert")
	}

	// 获取所有字段名（从第一条记录）
	var fields []string
	for field := range req.Data[0] {
		fields = append(fields, field)
	}

	var sql strings.Builder
	params := make(map[string]interface{})

	// INSERT INTO 子句
	sql.WriteString("INSERT INTO ")
	sql.WriteString(req.Table)
	sql.WriteString(" (")
	sql.WriteString(strings.Join(fields, ", "))
	sql.WriteString(") VALUES ")

	// VALUES 子句
	var valuesClauses []string
	paramIndex := 1

	for _, record := range req.Data {
		var placeholders []string
		for _, field := range fields {
			placeholder := b.getParameterPlaceholder(paramIndex)
			placeholders = append(placeholders, placeholder)
			params[fmt.Sprintf("param_%d", paramIndex)] = record[field]
			paramIndex++
		}
		valuesClauses = append(valuesClauses, "("+strings.Join(placeholders, ", ")+")")
	}

	sql.WriteString(strings.Join(valuesClauses, ", "))

	// 处理冲突
	if req.OnConflict != "" {
		conflictClause := b.buildOnConflictClause(req.OnConflict, fields)
		sql.WriteString(conflictClause)
	}

	// 返回字段
	if len(req.ReturnFields) > 0 {
		returnClause := b.buildReturningClause(req.ReturnFields)
		sql.WriteString(returnClause)
	}

	return sql.String(), params, nil
}

// buildOnConflictClause 构建冲突处理子句
func (b *QueryBuilder) buildOnConflictClause(onConflict string, fields []string) string {
	switch b.dbType {
	case "postgres":
		switch onConflict {
		case "ignore":
			return " ON CONFLICT DO NOTHING"
		case "update":
			var updateClauses []string
			for _, field := range fields {
				updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", field, field))
			}
			return " ON CONFLICT DO UPDATE SET " + strings.Join(updateClauses, ", ")
		}
	case "oracle":
		// Oracle 使用 MERGE 语句处理冲突，这里简化处理
		return ""
	}
	return ""
}

// buildReturningClause 构建返回字段子句
func (b *QueryBuilder) buildReturningClause(returnFields []string) string {
	switch b.dbType {
	case "postgres":
		return " RETURNING " + strings.Join(returnFields, ", ")
	case "oracle":
		return " RETURNING " + strings.Join(returnFields, ", ") + " INTO " + strings.Repeat(":out, ", len(returnFields)-1) + ":out"
	}
	return ""
}

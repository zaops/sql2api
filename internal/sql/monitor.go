package sql

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// QueryMetrics 查询指标
type QueryMetrics struct {
	QueryID       string        `json:"query_id"`
	DatabaseType  string        `json:"database_type"`
	QueryType     string        `json:"query_type"`     // select, insert, update, delete, batch
	SQL           string        `json:"sql,omitempty"`  // 脱敏后的 SQL
	ExecutionTime time.Duration `json:"execution_time"`
	AffectedRows  int64         `json:"affected_rows"`
	ResultRows    int64         `json:"result_rows"`
	Success       bool          `json:"success"`
	ErrorCode     int           `json:"error_code,omitempty"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
	ClientIP      string        `json:"client_ip,omitempty"`
	APIKey        string        `json:"api_key,omitempty"` // 脱敏后的 API Key
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	enabled     bool
	logQueries  bool
	logErrors   bool
	slowQueryMs int64 // 慢查询阈值（毫秒）
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(enabled, logQueries, logErrors bool, slowQueryMs int64) *PerformanceMonitor {
	return &PerformanceMonitor{
		enabled:     enabled,
		logQueries:  logQueries,
		logErrors:   logErrors,
		slowQueryMs: slowQueryMs,
	}
}

// StartQuery 开始查询监控
func (m *PerformanceMonitor) StartQuery(ctx context.Context, queryType, dbType, sql string) *QueryContext {
	if !m.enabled {
		return &QueryContext{enabled: false}
	}

	queryID := m.generateQueryID()
	
	queryCtx := &QueryContext{
		enabled:      true,
		monitor:      m,
		QueryID:      queryID,
		DatabaseType: dbType,
		QueryType:    queryType,
		SQL:          m.sanitizeSQL(sql),
		StartTime:    time.Now(),
		Context:      ctx,
	}

	// 记录查询开始
	if m.logQueries {
		log.Printf("[SQL-MONITOR] Query started - ID: %s, Type: %s, DB: %s", 
			queryID, queryType, dbType)
	}

	return queryCtx
}

// QueryContext 查询上下文
type QueryContext struct {
	enabled      bool
	monitor      *PerformanceMonitor
	QueryID      string
	DatabaseType string
	QueryType    string
	SQL          string
	StartTime    time.Time
	Context      context.Context
	ClientIP     string
	APIKey       string
}

// SetClientInfo 设置客户端信息
func (qc *QueryContext) SetClientInfo(clientIP, apiKey string) {
	if !qc.enabled {
		return
	}
	qc.ClientIP = clientIP
	qc.APIKey = qc.monitor.sanitizeAPIKey(apiKey)
}

// Finish 完成查询监控
func (qc *QueryContext) Finish(success bool, affectedRows, resultRows int64, err error) {
	if !qc.enabled {
		return
	}

	executionTime := time.Since(qc.StartTime)
	
	metrics := QueryMetrics{
		QueryID:       qc.QueryID,
		DatabaseType:  qc.DatabaseType,
		QueryType:     qc.QueryType,
		SQL:           qc.SQL,
		ExecutionTime: executionTime,
		AffectedRows:  affectedRows,
		ResultRows:    resultRows,
		Success:       success,
		Timestamp:     qc.StartTime,
		ClientIP:      qc.ClientIP,
		APIKey:        qc.APIKey,
	}

	if err != nil {
		metrics.ErrorMessage = err.Error()
		// 这里可以根据错误类型设置错误码
	}

	// 记录指标
	qc.monitor.recordMetrics(&metrics)

	// 记录日志
	qc.monitor.logQuery(&metrics)
}

// recordMetrics 记录查询指标
func (m *PerformanceMonitor) recordMetrics(metrics *QueryMetrics) {
	// 这里可以集成到监控系统，如 Prometheus、InfluxDB 等
	// 目前只是简单记录到日志
	
	if metrics.ExecutionTime.Milliseconds() > m.slowQueryMs {
		log.Printf("[SQL-MONITOR] SLOW QUERY - ID: %s, Time: %dms, Type: %s", 
			metrics.QueryID, metrics.ExecutionTime.Milliseconds(), metrics.QueryType)
	}
}

// logQuery 记录查询日志
func (m *PerformanceMonitor) logQuery(metrics *QueryMetrics) {
	if !metrics.Success && m.logErrors {
		log.Printf("[SQL-ERROR] Query failed - ID: %s, Error: %s, Time: %dms", 
			metrics.QueryID, metrics.ErrorMessage, metrics.ExecutionTime.Milliseconds())
	} else if metrics.Success && m.logQueries {
		log.Printf("[SQL-SUCCESS] Query completed - ID: %s, Type: %s, Time: %dms, Affected: %d, Results: %d", 
			metrics.QueryID, metrics.QueryType, metrics.ExecutionTime.Milliseconds(), 
			metrics.AffectedRows, metrics.ResultRows)
	}
}

// generateQueryID 生成查询 ID
func (m *PerformanceMonitor) generateQueryID() string {
	return fmt.Sprintf("sql_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// sanitizeSQL 脱敏 SQL 语句
func (m *PerformanceMonitor) sanitizeSQL(sql string) string {
	if sql == "" {
		return ""
	}

	// 移除敏感信息
	sanitized := sql
	
	// 替换字符串字面量
	sanitized = strings.ReplaceAll(sanitized, "'", "?")
	sanitized = strings.ReplaceAll(sanitized, "\"", "?")
	
	// 限制长度
	if len(sanitized) > 200 {
		sanitized = sanitized[:200] + "..."
	}
	
	return sanitized
}

// sanitizeAPIKey 脱敏 API Key
func (m *PerformanceMonitor) sanitizeAPIKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	
	if len(apiKey) <= 8 {
		return "****"
	}
	
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// GetMetricsSummary 获取指标摘要
func (m *PerformanceMonitor) GetMetricsSummary() map[string]interface{} {
	// 这里可以返回聚合的指标数据
	// 目前返回基本信息
	return map[string]interface{}{
		"enabled":        m.enabled,
		"log_queries":    m.logQueries,
		"log_errors":     m.logErrors,
		"slow_query_ms":  m.slowQueryMs,
		"monitor_status": "active",
	}
}

// MemoryOptimizer 内存优化器
type MemoryOptimizer struct {
	maxResultSize int
	batchSize     int
}

// NewMemoryOptimizer 创建内存优化器
func NewMemoryOptimizer(maxResultSize, batchSize int) *MemoryOptimizer {
	return &MemoryOptimizer{
		maxResultSize: maxResultSize,
		batchSize:     batchSize,
	}
}

// OptimizeResultSet 优化结果集内存使用
func (mo *MemoryOptimizer) OptimizeResultSet(rows []map[string]interface{}) []map[string]interface{} {
	if len(rows) <= mo.maxResultSize {
		return rows
	}

	// 如果结果集过大，只返回前 N 条记录
	log.Printf("[MEMORY-OPTIMIZER] Result set truncated from %d to %d rows", len(rows), mo.maxResultSize)
	return rows[:mo.maxResultSize]
}

// ShouldUseBatch 判断是否应该使用批量处理
func (mo *MemoryOptimizer) ShouldUseBatch(itemCount int) bool {
	return itemCount > mo.batchSize
}

// GetBatchSize 获取批量处理大小
func (mo *MemoryOptimizer) GetBatchSize() int {
	return mo.batchSize
}

// EstimateMemoryUsage 估算内存使用量
func (mo *MemoryOptimizer) EstimateMemoryUsage(rowCount, avgRowSize int) int64 {
	// 简单估算：行数 * 平均行大小 * 2（考虑 Go 的内存开销）
	return int64(rowCount * avgRowSize * 2)
}

// LogMemoryUsage 记录内存使用情况
func (mo *MemoryOptimizer) LogMemoryUsage(operation string, beforeMB, afterMB float64) {
	log.Printf("[MEMORY-OPTIMIZER] %s - Memory usage: %.2fMB -> %.2fMB (%.2fMB diff)", 
		operation, beforeMB, afterMB, afterMB-beforeMB)
}

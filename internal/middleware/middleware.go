package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"sql2api/internal/config"
	"sql2api/internal/model"

	"github.com/gin-gonic/gin"
)

// APIKeyManager API Key 管理器
type APIKeyManager struct {
	config *config.APIKeyConfig
	keyMap map[string]*config.APIKeyItem // Key -> APIKeyItem 映射，用于快速查找
}

// NewAPIKeyManager 创建 API Key 管理器
func NewAPIKeyManager(cfg *config.APIKeyConfig) *APIKeyManager {
	manager := &APIKeyManager{
		config: cfg,
		keyMap: make(map[string]*config.APIKeyItem),
	}

	// 构建 Key 映射表
	for i := range cfg.Keys {
		key := &cfg.Keys[i]
		if key.Active && key.Key != "" {
			manager.keyMap[key.Key] = key
		}
	}

	return manager
}

// ValidateAPIKey 验证 API Key
func (m *APIKeyManager) ValidateAPIKey(apiKey string) (*config.APIKeyItem, error) {
	if !m.config.Enabled {
		return nil, errors.New("API Key authentication is disabled")
	}

	if apiKey == "" {
		return nil, errors.New("API Key cannot be empty")
	}

	// 查找 API Key
	keyItem, exists := m.keyMap[apiKey]
	if !exists {
		return nil, errors.New("invalid API Key")
	}

	if !keyItem.Active {
		return nil, errors.New("API Key is inactive")
	}

	return keyItem, nil
}

// HasPermission 检查 API Key 是否有指定权限
func (m *APIKeyManager) HasPermission(apiKey string, permission string) bool {
	keyItem, err := m.ValidateAPIKey(apiKey)
	if err != nil {
		return false
	}

	// 如果没有配置权限，默认允许所有操作
	if len(keyItem.Permissions) == 0 {
		return true
	}

	// 检查是否有通配符权限
	for _, perm := range keyItem.Permissions {
		if perm == "*" || perm == "all" {
			return true
		}
		if perm == permission {
			return true
		}
		// 支持前缀匹配，如 "sql.*" 匹配 "sql.query", "sql.insert" 等
		if strings.HasSuffix(perm, "*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}

	return false
}

// IsAnonymousAllowed 检查是否允许匿名访问
func (m *APIKeyManager) IsAnonymousAllowed() bool {
	return m.config.AllowAnonymous
}

// IsEnabled 检查 API Key 认证是否启用
func (m *APIKeyManager) IsEnabled() bool {
	return m.config.Enabled
}

// GetHeaderName 获取 API Key 请求头名称
func (m *APIKeyManager) GetHeaderName() string {
	return m.config.HeaderName
}

// GetQueryParam 获取 API Key 查询参数名称
func (m *APIKeyManager) GetQueryParam() string {
	return m.config.QueryParam
}

// GetKeyInfo 获取 API Key 信息（不包含敏感信息）
func (m *APIKeyManager) GetKeyInfo(apiKey string) map[string]interface{} {
	keyItem, err := m.ValidateAPIKey(apiKey)
	if err != nil {
		return nil
	}

	return map[string]interface{}{
		"name":        keyItem.Name,
		"description": keyItem.Description,
		"permissions": keyItem.Permissions,
		"active":      keyItem.Active,
	}
}

// ListActiveKeys 列出所有活跃的 API Key 信息（不包含实际的 Key 值）
func (m *APIKeyManager) ListActiveKeys() []map[string]interface{} {
	var keys []map[string]interface{}

	for _, keyItem := range m.config.Keys {
		if keyItem.Active {
			keys = append(keys, map[string]interface{}{
				"name":        keyItem.Name,
				"description": keyItem.Description,
				"permissions": keyItem.Permissions,
				"active":      keyItem.Active,
			})
		}
	}

	return keys
}

// ValidatePermissions 验证权限格式
func ValidatePermissions(permissions []string) error {
	validPermissions := map[string]bool{
		"*":          true,
		"all":        true,
		"sql.query":  true,
		"sql.insert": true,
		"sql.update": true,
		"sql.delete": true,
		"sql.batch":  true,
		"sql.*":      true,
		"admin":      true,
		"read":       true,
		"write":      true,
	}

	for _, perm := range permissions {
		if !validPermissions[perm] {
			return fmt.Errorf("invalid permission: %s", perm)
		}
	}

	return nil
}

// SimpleAuthMiddleware 简化认证中间件
func SimpleAuthMiddleware(apiKeyManager *APIKeyManager, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将 API Key 管理器存储到上下文中
		c.Set("api_key_manager", apiKeyManager)

		// 如果不要求认证且允许匿名访问，直接通过
		if !required && apiKeyManager.IsAnonymousAllowed() {
			c.Next()
			return
		}

		// 尝试 API Key 认证
		if apiKeyManager.IsEnabled() {
			if authenticated := tryAPIKeyAuth(c, apiKeyManager); authenticated {
				c.Next()
				return
			}
		}

		// 如果不要求认证，允许通过
		if !required {
			c.Next()
			return
		}

		// 认证失败
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"Authentication required",
			"Please provide a valid API Key",
		))
		c.Abort()
	}
}

// tryAPIKeyAuth 尝试 API Key 认证
func tryAPIKeyAuth(c *gin.Context, apiKeyManager *APIKeyManager) bool {
	// 从请求头获取 API Key
	apiKey := c.GetHeader(apiKeyManager.GetHeaderName())
	
	// 如果请求头中没有，尝试从查询参数获取
	if apiKey == "" {
		apiKey = c.Query(apiKeyManager.GetQueryParam())
	}

	if apiKey == "" {
		return false
	}

	// 验证 API Key
	keyItem, err := apiKeyManager.ValidateAPIKey(apiKey)
	if err != nil {
		return false
	}

	// 将 API Key 信息存储到上下文中
	c.Set("auth_type", "api_key")
	c.Set("api_key", apiKey)
	c.Set("api_key_name", keyItem.Name)
	c.Set("api_key_permissions", keyItem.Permissions)
	c.Set("authenticated", true)

	return true
}



// RequireAPIKeyAuth 要求 API Key 认证的中间件
func RequireAPIKeyAuth(apiKeyManager *APIKeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !apiKeyManager.IsEnabled() {
			c.JSON(http.StatusServiceUnavailable, model.NewErrorResponse(
				http.StatusServiceUnavailable,
				"API Key authentication is disabled",
			))
			c.Abort()
			return
		}

		if !tryAPIKeyAuth(c, apiKeyManager) {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Valid API Key required",
				"Please provide a valid API Key in header or query parameter",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission 要求特定权限的中间件
func RequirePermission(apiKeyManager *APIKeyManager, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authType, exists := c.Get("auth_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authentication required",
			))
			c.Abort()
			return
		}

		// 检查 API Key 权限
		if authType == "api_key" {
			apiKey, _ := c.Get("api_key")
			if apiKey == nil {
				c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
					http.StatusUnauthorized,
					"API Key not found in context",
				))
				c.Abort()
				return
			}

			if !apiKeyManager.HasPermission(apiKey.(string), permission) {
				c.JSON(http.StatusForbidden, model.NewErrorResponse(
					http.StatusForbidden,
					"Insufficient permissions",
					"Required permission: "+permission,
				))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// GetAuthInfo 获取认证信息的辅助函数
func GetAuthInfo(c *gin.Context) map[string]interface{} {
	authType, exists := c.Get("auth_type")
	if !exists {
		return map[string]interface{}{
			"authenticated": false,
		}
	}

	info := map[string]interface{}{
		"authenticated": true,
		"auth_type":     authType,
	}

	if authType == "api_key" {
		if name, exists := c.Get("api_key_name"); exists {
			info["api_key_name"] = name
		}
		if permissions, exists := c.Get("api_key_permissions"); exists {
			info["permissions"] = permissions
		}
	}

	return info
}

// IPWhitelistManager IP 白名单管理器
type IPWhitelistManager struct {
	allowedIPs   []net.IP
	allowedCIDRs []*net.IPNet
	enabled      bool
}

// NewIPWhitelistManager 创建 IP 白名单管理器
func NewIPWhitelistManager(cfg *config.SecurityConfig) (*IPWhitelistManager, error) {
	manager := &IPWhitelistManager{
		enabled: len(cfg.IPWhitelist) > 0,
	}

	if !manager.enabled {
		return manager, nil
	}

	for _, ipStr := range cfg.IPWhitelist {
		ipStr = strings.TrimSpace(ipStr)
		if ipStr == "" {
			continue
		}

		// 检查是否是 CIDR 格式
		if strings.Contains(ipStr, "/") {
			_, cidr, err := net.ParseCIDR(ipStr)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR format '%s': %w", ipStr, err)
			}
			manager.allowedCIDRs = append(manager.allowedCIDRs, cidr)
		} else {
			// 单个 IP 地址
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP address '%s'", ipStr)
			}
			manager.allowedIPs = append(manager.allowedIPs, ip)
		}
	}

	return manager, nil
}

// IsAllowed 检查 IP 是否在白名单中
func (m *IPWhitelistManager) IsAllowed(clientIP string) bool {
	if !m.enabled {
		return true // 如果未启用白名单，允许所有 IP
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false // 无效的 IP 地址
	}

	// 检查单个 IP
	for _, allowedIP := range m.allowedIPs {
		if ip.Equal(allowedIP) {
			return true
		}
	}

	// 检查 CIDR 范围
	for _, cidr := range m.allowedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// IsEnabled 检查 IP 白名单是否启用
func (m *IPWhitelistManager) IsEnabled() bool {
	return m.enabled
}

// GetAllowedIPs 获取允许的 IP 列表（用于调试）
func (m *IPWhitelistManager) GetAllowedIPs() []string {
	var ips []string

	for _, ip := range m.allowedIPs {
		ips = append(ips, ip.String())
	}

	for _, cidr := range m.allowedCIDRs {
		ips = append(ips, cidr.String())
	}

	return ips
}

// IPWhitelistMiddleware IP 白名单中间件
func IPWhitelistMiddleware(manager *IPWhitelistManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager == nil || !manager.IsEnabled() {
			c.Next()
			return
		}

		clientIP := GetClientIP(c)
		if !manager.IsAllowed(clientIP) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(
				http.StatusForbidden,
				"Access denied",
				fmt.Sprintf("IP address %s is not in the whitelist", clientIP),
			))
			c.Abort()
			return
		}

		// 将客户端 IP 存储到上下文中，供后续使用
		c.Set("client_ip", clientIP)
		c.Next()
	}
}

// GetClientIP 获取客户端真实 IP 地址
func GetClientIP(c *gin.Context) string {
	// 优先级顺序：X-Forwarded-For -> X-Real-IP -> RemoteAddr

	// 1. 检查 X-Forwarded-For 头（可能包含多个 IP，取第一个）
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 2. 检查 X-Real-IP 头
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && net.ParseIP(xRealIP) != nil {
		return xRealIP
	}

	// 3. 使用 RemoteAddr（可能包含端口号）
	remoteAddr := c.Request.RemoteAddr
	if remoteAddr != "" {
		ip, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			// 如果没有端口号，直接返回
			if net.ParseIP(remoteAddr) != nil {
				return remoteAddr
			}
		} else {
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	return "unknown"
}

package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"sql2api/internal/config"
	"sql2api/internal/model"

	"github.com/gin-gonic/gin"
)

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
func (m *IPWhitelistManager) IsAllowed(ipStr string) bool {
	if !m.enabled {
		return true // 如果未启用白名单，允许所有 IP
	}

	if ipStr == "" {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
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

// GetClientIP 获取客户端真实 IP 地址
func GetClientIP(c *gin.Context) string {
	// 优先级顺序：X-Forwarded-For > X-Real-IP > RemoteAddr

	// 1. 检查 X-Forwarded-For 头部（可能包含多个 IP，第一个是客户端 IP）
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个 IP，格式：client, proxy1, proxy2
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" && net.ParseIP(clientIP) != nil {
				return clientIP
			}
		}
	}

	// 2. 检查 X-Real-IP 头部
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" && net.ParseIP(xRealIP) != nil {
		return xRealIP
	}

	// 3. 检查 X-Forwarded 的其他格式
	xForwarded := c.GetHeader("X-Forwarded")
	if xForwarded != "" {
		// 格式可能是：for=192.168.1.1
		if strings.HasPrefix(xForwarded, "for=") {
			ip := strings.TrimPrefix(xForwarded, "for=")
			ip = strings.TrimSpace(ip)
			if ip != "" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 4. 使用 Gin 的 ClientIP 方法作为后备
	clientIP := c.ClientIP()
	if clientIP != "" {
		return clientIP
	}

	// 5. 最后从 RemoteAddr 提取 IP
	if c.Request != nil && c.Request.RemoteAddr != "" {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil && host != "" {
			return host
		}
		// 如果没有端口，直接返回地址
		if net.ParseIP(c.Request.RemoteAddr) != nil {
			return c.Request.RemoteAddr
		}
	}

	return "127.0.0.1" // 默认返回本地地址
}

// IPWhitelistMiddleware IP 白名单中间件
func IPWhitelistMiddleware(manager *IPWhitelistManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !manager.enabled {
			c.Next()
			return
		}

		clientIP := GetClientIP(c)

		if !manager.IsAllowed(clientIP) {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(
				http.StatusForbidden,
				"Access denied",
				fmt.Sprintf("IP address %s is not allowed", clientIP),
			))
			c.Abort()
			return
		}

		// 将客户端 IP 存储到上下文中，供后续使用
		c.Set("client_ip", clientIP)
		c.Next()
	}
}

// LoggingIPMiddleware IP 记录中间件（用于调试和审计）
func LoggingIPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := GetClientIP(c)

		// 记录所有相关的 IP 头部信息
		headers := map[string]string{
			"X-Forwarded-For": c.GetHeader("X-Forwarded-For"),
			"X-Real-IP":       c.GetHeader("X-Real-IP"),
			"X-Forwarded":     c.GetHeader("X-Forwarded"),
			"RemoteAddr":      c.Request.RemoteAddr,
			"ClientIP":        clientIP,
		}

		// 存储到上下文中
		c.Set("ip_headers", headers)
		c.Set("client_ip", clientIP)

		c.Next()
	}
}

// GetClientIPFromContext 从上下文获取客户端 IP
func GetClientIPFromContext(c *gin.Context) string {
	if ip, exists := c.Get("client_ip"); exists {
		return ip.(string)
	}
	return GetClientIP(c)
}

// GetIPHeadersFromContext 从上下文获取 IP 头部信息
func GetIPHeadersFromContext(c *gin.Context) map[string]string {
	if headers, exists := c.Get("ip_headers"); exists {
		return headers.(map[string]string)
	}
	return nil
}

// CreateIPInfoEndpoint 创建 IP 信息查看端点（用于调试）
func CreateIPInfoEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := GetClientIP(c)

		info := map[string]interface{}{
			"client_ip": clientIP,
			"headers": map[string]string{
				"X-Forwarded-For": c.GetHeader("X-Forwarded-For"),
				"X-Real-IP":       c.GetHeader("X-Real-IP"),
				"X-Forwarded":     c.GetHeader("X-Forwarded"),
				"User-Agent":      c.GetHeader("User-Agent"),
			},
			"remote_addr": c.Request.RemoteAddr,
			"method":      c.Request.Method,
			"url":         c.Request.URL.String(),
		}

		c.JSON(http.StatusOK, model.NewSuccessResponse(info, "IP information retrieved"))
	}
}

// ValidateIPWhitelist 验证 IP 白名单配置
func ValidateIPWhitelist(ipList []string) error {
	for _, ipStr := range ipList {
		ipStr = strings.TrimSpace(ipStr)
		if ipStr == "" {
			continue
		}

		// 检查是否是 CIDR 格式
		if strings.Contains(ipStr, "/") {
			_, _, err := net.ParseCIDR(ipStr)
			if err != nil {
				return fmt.Errorf("invalid CIDR format '%s': %w", ipStr, err)
			}
		} else {
			// 单个 IP 地址
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return fmt.Errorf("invalid IP address '%s'", ipStr)
			}
		}
	}
	return nil
}

// IsPrivateIP 检查是否是私有 IP 地址
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// IPv4 私有地址范围
	private4 := []string{
		"10.0.0.0/8",     // Class A
		"172.16.0.0/12",  // Class B
		"192.168.0.0/16", // Class C
		"127.0.0.0/8",    // Loopback
	}

	// IPv6 私有地址范围
	private6 := []string{
		"::1/128",   // Loopback
		"fc00::/7",  // Unique local
		"fe80::/10", // Link local
	}

	allPrivate := append(private4, private6...)

	for _, cidrStr := range allPrivate {
		_, cidr, err := net.ParseCIDR(cidrStr)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// GetPublicIP 获取公网 IP（从代理头部中提取）
func GetPublicIP(c *gin.Context) string {
	clientIP := GetClientIP(c)
	ip := net.ParseIP(clientIP)

	if ip != nil && !IsPrivateIP(ip) {
		return clientIP
	}

	// 如果客户端 IP 是私有地址，尝试从其他头部获取公网 IP
	headers := []string{"X-Forwarded-For", "X-Real-IP"}

	for _, header := range headers {
		value := c.GetHeader(header)
		if value == "" {
			continue
		}

		ips := strings.Split(value, ",")
		for _, ipStr := range ips {
			ipStr = strings.TrimSpace(ipStr)
			ip := net.ParseIP(ipStr)
			if ip != nil && !IsPrivateIP(ip) {
				return ipStr
			}
		}
	}

	return clientIP // 返回原始 IP，即使是私有地址
}

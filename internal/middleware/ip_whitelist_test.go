package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"sql2api/internal/config"

	"github.com/gin-gonic/gin"
)

func TestNewIPWhitelistManager(t *testing.T) {
	tests := []struct {
		name        string
		ipWhitelist []string
		expectError bool
		enabled     bool
	}{
		{
			name:        "Empty whitelist",
			ipWhitelist: []string{},
			expectError: false,
			enabled:     false,
		},
		{
			name:        "Valid single IP",
			ipWhitelist: []string{"192.168.1.1"},
			expectError: false,
			enabled:     true,
		},
		{
			name:        "Valid CIDR",
			ipWhitelist: []string{"192.168.0.0/16"},
			expectError: false,
			enabled:     true,
		},
		{
			name:        "Mixed valid IPs and CIDRs",
			ipWhitelist: []string{"127.0.0.1", "192.168.0.0/16", "10.0.0.0/8"},
			expectError: false,
			enabled:     true,
		},
		{
			name:        "Invalid IP",
			ipWhitelist: []string{"invalid.ip.address"},
			expectError: true,
			enabled:     true,
		},
		{
			name:        "Invalid CIDR",
			ipWhitelist: []string{"192.168.0.0/33"},
			expectError: true,
			enabled:     true,
		},
		{
			name:        "IPv6 addresses",
			ipWhitelist: []string{"::1", "2001:db8::/32"},
			expectError: false,
			enabled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SecurityConfig{
				IPWhitelist: tt.ipWhitelist,
			}

			manager, err := NewIPWhitelistManager(cfg)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if manager.enabled != tt.enabled {
				t.Errorf("Expected enabled=%v, got %v", tt.enabled, manager.enabled)
			}
		})
	}
}

func TestIPWhitelistManager_IsAllowed(t *testing.T) {
	cfg := &config.SecurityConfig{
		IPWhitelist: []string{
			"127.0.0.1",
			"192.168.1.100",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"::1",
		},
	}

	manager, err := NewIPWhitelistManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Localhost IPv4", "127.0.0.1", true},
		{"Specific allowed IP", "192.168.1.100", true},
		{"IP in CIDR range 10.x", "10.1.2.3", true},
		{"IP in CIDR range 172.16.x", "172.16.1.1", true},
		{"Localhost IPv6", "::1", true},
		{"Disallowed IP", "8.8.8.8", false},
		{"Invalid IP", "invalid", false},
		{"Empty IP", "", false},
		{"IP outside CIDR", "192.168.2.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.IsAllowed(tt.ip)
			if result != tt.expected {
				t.Errorf("IsAllowed(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestIPWhitelistManager_Disabled(t *testing.T) {
	// 测试禁用状态的管理器
	cfg := &config.SecurityConfig{
		IPWhitelist: []string{}, // 空白名单，应该禁用
	}

	manager, err := NewIPWhitelistManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.enabled {
		t.Error("Expected manager to be disabled")
	}

	// 禁用状态下应该允许所有 IP
	testIPs := []string{"127.0.0.1", "8.8.8.8", "192.168.1.1", "invalid"}
	for _, ip := range testIPs {
		if !manager.IsAllowed(ip) {
			t.Errorf("Disabled manager should allow all IPs, but rejected %s", ip)
		}
	}
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
			},
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1, 172.16.0.1",
			},
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			expectedIP: "192.168.1.200",
		},
		{
			name: "X-Forwarded format",
			headers: map[string]string{
				"X-Forwarded": "for=192.168.1.200",
			},
			expectedIP: "192.168.1.200",
		},
		{
			name: "Priority test - X-Forwarded-For wins",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100",
				"X-Real-IP":       "192.168.1.200",
			},
			expectedIP: "192.168.1.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req, _ := http.NewRequest("GET", "/test", nil)

			// 设置头部
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			// 创建测试上下文
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 获取客户端 IP
			clientIP := GetClientIP(c)

			if clientIP != tt.expectedIP {
				t.Errorf("GetClientIP() = %s, expected %s", clientIP, tt.expectedIP)
			}
		})
	}
}

func TestIPWhitelistMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.SecurityConfig{
		IPWhitelist: []string{"127.0.0.1", "192.168.1.0/24"},
	}

	manager, err := NewIPWhitelistManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	middleware := IPWhitelistMiddleware(manager)

	tests := []struct {
		name           string
		clientIP       string
		expectedStatus int
		shouldAbort    bool
	}{
		{
			name:           "Allowed IP",
			clientIP:       "127.0.0.1",
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name:           "IP in allowed CIDR",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name:           "Disallowed IP",
			clientIP:       "8.8.8.8",
			expectedStatus: http.StatusForbidden,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(middleware)
			router.GET("/test", func(c *gin.Context) {
				// 检查客户端 IP 是否正确存储
				if !tt.shouldAbort {
					storedIP, exists := c.Get("client_ip")
					if !exists {
						t.Error("Expected client_ip to be stored in context")
					} else if storedIP.(string) != tt.clientIP {
						t.Errorf("Expected stored IP %s, got %s", tt.clientIP, storedIP.(string))
					}
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// 创建测试请求
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("X-Forwarded-For", tt.clientIP)

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应状态码
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestValidateIPWhitelist(t *testing.T) {
	tests := []struct {
		name        string
		ipList      []string
		expectError bool
	}{
		{
			name:        "Valid IPs and CIDRs",
			ipList:      []string{"127.0.0.1", "192.168.0.0/16", "::1"},
			expectError: false,
		},
		{
			name:        "Invalid IP",
			ipList:      []string{"invalid.ip"},
			expectError: true,
		},
		{
			name:        "Invalid CIDR",
			ipList:      []string{"192.168.0.0/33"},
			expectError: true,
		},
		{
			name:        "Empty list",
			ipList:      []string{},
			expectError: false,
		},
		{
			name:        "Empty strings",
			ipList:      []string{"", "  ", "127.0.0.1"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPWhitelist(tt.ipList)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Localhost IPv4", "127.0.0.1", true},
		{"Private Class A", "10.1.2.3", true},
		{"Private Class B", "172.16.1.1", true},
		{"Private Class C", "192.168.1.1", true},
		{"Public IP", "8.8.8.8", false},
		{"Localhost IPv6", "::1", true},
		{"Link local IPv6", "fe80::1", true},
		{"Public IPv6", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			result := IsPrivateIP(ip)
			if result != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sql2api/internal/config"
	"sql2api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTManager(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-for-testing-purposes",
		Issuer:     "test-issuer",
		Expiration: 24,
	}

	manager := NewJWTManager(cfg)
	if manager == nil {
		t.Error("Expected JWT manager to be created")
	}

	if string(manager.secretKey) != cfg.Secret {
		t.Errorf("Expected secret key %s, got %s", cfg.Secret, string(manager.secretKey))
	}

	if manager.issuer != cfg.Issuer {
		t.Errorf("Expected issuer %s, got %s", cfg.Issuer, manager.issuer)
	}

	expectedExpiry := time.Duration(cfg.Expiration) * time.Hour
	if manager.expiry != expectedExpiry {
		t.Errorf("Expected expiry %v, got %v", expectedExpiry, manager.expiry)
	}
}

func TestJWTManager_GenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-for-testing-purposes",
		Issuer:     "test-issuer",
		Expiration: 24,
	}

	manager := NewJWTManager(cfg)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// 测试正常生成令牌
	token, expiresAt, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	if expiresAt.IsZero() {
		t.Error("Expected expiration time to be set")
	}

	// 验证过期时间是否正确
	expectedExpiry := time.Now().Add(24 * time.Hour)
	if expiresAt.Before(expectedExpiry.Add(-time.Minute)) || expiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("Expected expiration time around %v, got %v", expectedExpiry, expiresAt)
	}

	// 测试空用户
	_, _, err = manager.GenerateToken(nil)
	if err == nil {
		t.Error("Expected error when generating token for nil user")
	}

	expectedError := "user cannot be nil"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJWTManager_ValidateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-for-testing-purposes",
		Issuer:     "test-issuer",
		Expiration: 24,
	}

	manager := NewJWTManager(cfg)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// 生成有效令牌
	token, _, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 测试验证有效令牌
	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.Issuer != cfg.Issuer {
		t.Errorf("Expected issuer %s, got %s", cfg.Issuer, claims.Issuer)
	}

	// 测试空令牌
	_, err = manager.ValidateToken("")
	if err == nil {
		t.Error("Expected error when validating empty token")
	}

	expectedError := "token cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}

	// 测试无效令牌
	_, err = manager.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("Expected error when validating invalid token")
	}

	// 测试过期令牌
	expiredClaims := JWTClaims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(cfg.Secret))

	_, err = manager.ValidateToken(expiredTokenString)
	if err == nil {
		t.Error("Expected error when validating expired token")
	}
}

func TestJWTAuthMiddleware(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret-key-for-testing-purposes",
		Issuer:     "test-issuer",
		Expiration: 24,
	}

	manager := NewJWTManager(cfg)
	middleware := JWTAuthMiddleware(manager)

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, _, err := manager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		shouldAbort    bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name:           "Invalid bearer format",
			authHeader:     "InvalidFormat " + token,
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试路由
			router := gin.New()
			router.Use(middleware)
			router.GET("/test", func(c *gin.Context) {
				// 如果中间件通过，检查用户信息是否正确设置
				if !tt.shouldAbort {
					userID, exists := c.Get("user_id")
					if !exists {
						t.Error("Expected user_id to be set in context")
					} else if userID.(uint) != user.ID {
						t.Errorf("Expected user ID %d, got %d", user.ID, userID.(uint))
					}

					username, exists := c.Get("username")
					if !exists {
						t.Error("Expected username to be set in context")
					} else if username.(string) != user.Username {
						t.Errorf("Expected username %s, got %s", user.Username, username.(string))
					}
				}

				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// 创建测试请求
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

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

func TestGetCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试有用户信息的情况
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", uint(1))
	c.Set("username", "testuser")
	c.Set("email", "test@example.com")

	userID, username, email, exists := GetCurrentUser(c)
	if !exists {
		t.Error("Expected user to exist in context")
	}

	if userID != 1 {
		t.Errorf("Expected user ID 1, got %d", userID)
	}

	if username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", username)
	}

	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", email)
	}

	// 测试没有用户信息的情况
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, _, _, exists = GetCurrentUser(c2)
	if exists {
		t.Error("Expected user not to exist in context")
	}
}

func TestGetCurrentUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试有用户ID的情况
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", uint(1))

	userID, exists := GetCurrentUserID(c)
	if !exists {
		t.Error("Expected user ID to exist in context")
	}

	if userID != 1 {
		t.Errorf("Expected user ID 1, got %d", userID)
	}

	// 测试没有用户ID的情况
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, exists = GetCurrentUserID(c2)
	if exists {
		t.Error("Expected user ID not to exist in context")
	}
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 测试已认证的情况
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("user_id", uint(1))

	if !RequireAuth(c) {
		t.Error("Expected user to be authenticated")
	}

	// 测试未认证的情况
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	if RequireAuth(c2) {
		t.Error("Expected user not to be authenticated")
	}
}

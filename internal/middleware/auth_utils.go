package middleware

import (
	"errors"
	"net/http"

	"sql2api/internal/model"

	"github.com/gin-gonic/gin"
)

// AuthResponse 认证响应结构
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt string    `json:"expires_at"`
	User      model.UserResponse `json:"user"`
}

// LoginHandler 登录处理器类型
type LoginHandler func(username, password string) (*model.User, error)

// CreateLoginEndpoint 创建登录端点
func CreateLoginEndpoint(jwtManager *JWTManager, loginHandler LoginHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq model.LoginRequest
		
		// 绑定请求数据
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(
				http.StatusBadRequest,
				"Invalid request format",
				err.Error(),
			))
			return
		}
		
		// 验证用户凭据
		user, err := loginHandler(loginReq.Username, loginReq.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Invalid credentials",
				err.Error(),
			))
			return
		}
		
		// 生成 JWT 令牌
		token, expiresAt, err := jwtManager.GenerateToken(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
				http.StatusInternalServerError,
				"Failed to generate token",
				err.Error(),
			))
			return
		}
		
		// 返回登录响应
		response := AuthResponse{
			Token:     token,
			ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
			User:      user.ToResponse(),
		}
		
		c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Login successful"))
	}
}

// CreateRefreshEndpoint 创建令牌刷新端点
func CreateRefreshEndpoint(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取当前令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authorization header is required",
			))
			return
		}
		
		// 检查 Bearer 前缀
		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authorization header must start with 'Bearer '",
			))
			return
		}
		
		// 提取令牌
		tokenString := authHeader[len(bearerPrefix):]
		
		// 刷新令牌
		newToken, expiresAt, err := jwtManager.RefreshToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Failed to refresh token",
				err.Error(),
			))
			return
		}
		
		// 返回新令牌
		response := map[string]interface{}{
			"token":      newToken,
			"expires_at": expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		
		c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Token refreshed successfully"))
	}
}

// CreateLogoutEndpoint 创建登出端点（可选实现）
func CreateLogoutEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// JWT 是无状态的，登出主要是客户端删除令牌
		// 这里可以实现令牌黑名单功能（如果需要的话）
		
		c.JSON(http.StatusOK, model.NewSuccessResponse(nil, "Logout successful"))
	}
}

// CreateProfileEndpoint 创建获取用户信息端点
func CreateProfileEndpoint() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户信息
		userID, username, email, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"User not authenticated",
			))
			return
		}
		
		// 构造用户响应
		userResponse := model.UserResponse{
			ID:       userID,
			Username: username,
			Email:    email,
		}
		
		c.JSON(http.StatusOK, model.NewSuccessResponse(userResponse, "Profile retrieved successfully"))
	}
}

// ValidateUserPermission 验证用户权限的辅助函数
func ValidateUserPermission(c *gin.Context, requiredUserID uint) error {
	currentUserID, exists := GetCurrentUserID(c)
	if !exists {
		return errors.New("user not authenticated")
	}
	
	if currentUserID != requiredUserID {
		return errors.New("insufficient permissions")
	}
	
	return nil
}

// RequireAuthMiddleware 要求认证的中间件包装器
func RequireAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !RequireAuth(c) {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authentication required",
			))
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminOnlyMiddleware 仅管理员访问的中间件（示例）
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现管理员权限检查
		// 例如检查用户角色或权限
		
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authentication required",
			))
			c.Abort()
			return
		}
		
		// 示例：假设用户ID为1的是管理员
		if userID != 1 {
			c.JSON(http.StatusForbidden, model.NewErrorResponse(
				http.StatusForbidden,
				"Admin access required",
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否允许该来源
		allowed := false
		for _, allowedOrigin := range origins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

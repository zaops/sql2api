package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sql2api/internal/config"
	"sql2api/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT 声明结构
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey: []byte(cfg.Secret),
		issuer:    cfg.Issuer,
		expiry:    time.Duration(cfg.Expiration) * time.Hour,
	}
}

// GenerateToken 生成 JWT 令牌
func (j *JWTManager) GenerateToken(user *model.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, errors.New("user cannot be nil")
	}

	now := time.Now()
	expiresAt := now.Add(j.expiry)

	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Audience:  []string{"sql2api"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d-%d", user.ID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken 验证 JWT 令牌
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// 验证发行者
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", j.issuer, claims.Issuer)
	}

	// 验证过期时间
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	// 验证生效时间
	if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
		return nil, errors.New("token not yet valid")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *JWTManager) RefreshToken(tokenString string) (string, time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// 检查令牌是否即将过期（在过期前1小时内可以刷新）
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", time.Time{}, errors.New("token is not eligible for refresh yet")
	}

	// 创建新的声明
	now := time.Now()
	expiresAt := now.Add(j.expiry)

	newClaims := JWTClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   claims.Subject,
			Audience:  claims.Audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d-%d", claims.UserID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refreshed token: %w", err)
	}

	return newTokenString, expiresAt, nil
}

// JWTAuthMiddleware JWT 认证中间件
func JWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authorization header is required",
			))
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Authorization header must start with 'Bearer '",
			))
			c.Abort()
			return
		}

		// 提取令牌
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Token is required",
			))
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				http.StatusUnauthorized,
				"Invalid token",
				err.Error(),
			))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("jwt_claims", claims)

		// 继续处理请求
		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选的 JWT 认证中间件（不强制要求认证）
func OptionalJWTAuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization header 获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有认证信息，继续处理请求
			c.Next()
			return
		}

		// 检查 Bearer 前缀
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			// 格式错误，继续处理请求（不设置用户信息）
			c.Next()
			return
		}

		// 提取令牌
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenString == "" {
			// 空令牌，继续处理请求
			c.Next()
			return
		}

		// 验证令牌
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// 令牌无效，继续处理请求（不设置用户信息）
			c.Next()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("jwt_claims", claims)

		// 继续处理请求
		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(c *gin.Context) (uint, string, string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, "", "", false
	}

	username, _ := c.Get("username")
	email, _ := c.Get("email")

	return userID.(uint), username.(string), email.(string), true
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// RequireAuth 检查是否已认证的辅助函数
func RequireAuth(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

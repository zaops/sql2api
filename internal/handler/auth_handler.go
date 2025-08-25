package handler

import (
	"net/http"

	"sql2api/internal/middleware"
	"sql2api/internal/model"
	"sql2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	userService service.UserService
	jwtManager  *middleware.JWTManager
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(userService service.UserService, jwtManager *middleware.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户使用用户名和密码登录，返回 JWT 令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录请求"
// @Success 200 {object} model.SuccessResponse "登录成功"
// @Failure 400 {object} model.ErrorResponse "请求格式错误"
// @Failure 401 {object} model.ErrorResponse "认证失败"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 验证用户凭据
	user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"Login failed",
			err.Error(),
		))
		return
	}

	// 生成 JWT 令牌
	token, expiresAt, err := h.jwtManager.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			http.StatusInternalServerError,
			"Failed to generate token",
			err.Error(),
		))
		return
	}

	// 更新最后登录时间
	if err := h.userService.UpdateLastLogin(user.ID); err != nil {
		// 记录错误但不影响登录流程
		// 在实际应用中可以使用日志记录
	}

	// 返回登录响应
	response := middleware.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		User:      user.ToResponse(),
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Login successful"))
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.UserCreateRequest true "注册请求"
// @Success 201 {object} model.SuccessResponse "注册成功"
// @Failure 400 {object} model.ErrorResponse "请求格式错误"
// @Failure 409 {object} model.ErrorResponse "用户已存在"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.UserCreateRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 注册用户
	user, err := h.userService.Register(&req)
	if err != nil {
		// 根据错误类型返回不同状态码
		statusCode := http.StatusBadRequest
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, model.NewErrorResponse(
			statusCode,
			"Registration failed",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(user.ToResponse(), "Registration successful"))
}

// RefreshToken 刷新令牌
// @Summary 刷新 JWT 令牌
// @Description 使用当前令牌获取新的令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse "刷新成功"
// @Failure 401 {object} model.ErrorResponse "令牌无效"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
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
	newToken, expiresAt, err := h.jwtManager.RefreshToken(tokenString)
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

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出（客户端删除令牌）
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse "登出成功"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 是无状态的，登出主要是客户端删除令牌
	// 这里可以实现令牌黑名单功能（如果需要的话）

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, "Logout successful"))
}

// Profile 获取用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse "获取成功"
// @Failure 401 {object} model.ErrorResponse "未认证"
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) Profile(c *gin.Context) {
	// 从上下文获取用户信息
	userID, username, email, exists := middleware.GetCurrentUser(c)
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

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} model.SuccessResponse "修改成功"
// @Failure 400 {object} model.ErrorResponse "请求格式错误"
// @Failure 401 {object} model.ErrorResponse "未认证或旧密码错误"
// @Failure 500 {object} model.ErrorResponse "服务器内部错误"
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 获取当前用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			http.StatusUnauthorized,
			"User not authenticated",
		))
		return
	}

	// 修改密码
	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "old password is incorrect" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, model.NewErrorResponse(
			statusCode,
			"Failed to change password",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil, "Password changed successfully"))
}

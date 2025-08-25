package handler

import (
	"sql2api/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(
	router *gin.Engine,
	handlers *Handlers,
	jwtManager *middleware.JWTManager,
	ipManager *middleware.IPWhitelistManager,
) {
	// 应用全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// IP 白名单中间件
	if ipManager != nil {
		router.Use(middleware.IPWhitelistMiddleware(ipManager))
	}

	// CORS 中间件
	router.Use(middleware.CORSMiddleware([]string{"*"}))

	// API 版本组
	v1 := router.Group("/api/v1")

	// 认证相关路由（不需要 JWT 认证）
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)
		auth.POST("/register", handlers.Auth.Register)

		// 需要 JWT 认证的认证路由
		authProtected := auth.Group("")
		authProtected.Use(middleware.JWTAuthMiddleware(jwtManager))
		{
			authProtected.POST("/refresh", handlers.Auth.RefreshToken)
			authProtected.POST("/logout", handlers.Auth.Logout)
			authProtected.GET("/profile", handlers.Auth.Profile)
			authProtected.POST("/change-password", handlers.Auth.ChangePassword)
		}
	}

	// 统一资源路由（需要 JWT 认证）
	resource := v1.Group("/resource")
	resource.Use(middleware.JWTAuthMiddleware(jwtManager))
	{
		resource.POST("", handlers.Resource.HandleResource)
	}

	// 健康检查路由（不需要认证）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": gin.H{"now": "2023-01-01T00:00:00Z"},
			"version":   "1.0.0",
		})
	})

	// IP 信息查看路由（调试用）
	router.GET("/debug/ip", middleware.CreateIPInfoEndpoint())

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// SetupRoutesWithOptionalAuth 设置带可选认证的路由
func SetupRoutesWithOptionalAuth(
	router *gin.Engine,
	handlers *Handlers,
	jwtManager *middleware.JWTManager,
	ipManager *middleware.IPWhitelistManager,
) {
	// 应用全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// IP 白名单中间件
	if ipManager != nil {
		router.Use(middleware.IPWhitelistMiddleware(ipManager))
	}

	// CORS 中间件
	router.Use(middleware.CORSMiddleware([]string{"*"}))

	// 可选 JWT 认证中间件（用于某些可以匿名访问的端点）
	router.Use(middleware.OptionalJWTAuthMiddleware(jwtManager))

	// API 版本组
	v1 := router.Group("/api/v1")

	// 认证相关路由
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handlers.Auth.Login)
		auth.POST("/register", handlers.Auth.Register)
		auth.POST("/refresh", handlers.Auth.RefreshToken)
		auth.POST("/logout", handlers.Auth.Logout)

		// 需要强制认证的路由
		authRequired := auth.Group("")
		authRequired.Use(middleware.RequireAuthMiddleware())
		{
			authRequired.GET("/profile", handlers.Auth.Profile)
			authRequired.POST("/change-password", handlers.Auth.ChangePassword)
		}
	}

	// 统一资源路由
	resource := v1.Group("/resource")
	resource.Use(middleware.RequireAuthMiddleware()) // 资源操作需要强制认证
	{
		resource.POST("", handlers.Resource.HandleResource)
	}

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": gin.H{"now": "2023-01-01T00:00:00Z"},
			"version":   "1.0.0",
		})
	})

	// IP 信息查看路由
	router.GET("/debug/ip", middleware.CreateIPInfoEndpoint())

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// RouteInfo 路由信息结构
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	Middlewares []string `json:"middlewares,omitempty"`
	Description string   `json:"description,omitempty"`
}

// GetRouteInfo 获取路由信息
func GetRouteInfo() []RouteInfo {
	return []RouteInfo{
		{
			Method:      "POST",
			Path:        "/api/v1/auth/login",
			Handler:     "AuthHandler.Login",
			Description: "用户登录",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/auth/register",
			Handler:     "AuthHandler.Register",
			Description: "用户注册",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/auth/refresh",
			Handler:     "AuthHandler.RefreshToken",
			Middlewares: []string{"JWTAuth"},
			Description: "刷新令牌",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/auth/logout",
			Handler:     "AuthHandler.Logout",
			Middlewares: []string{"JWTAuth"},
			Description: "用户登出",
		},
		{
			Method:      "GET",
			Path:        "/api/v1/auth/profile",
			Handler:     "AuthHandler.Profile",
			Middlewares: []string{"JWTAuth"},
			Description: "获取用户信息",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/auth/change-password",
			Handler:     "AuthHandler.ChangePassword",
			Middlewares: []string{"JWTAuth"},
			Description: "修改密码",
		},
		{
			Method:      "POST",
			Path:        "/api/v1/resource",
			Handler:     "ResourceHandler.HandleResource",
			Middlewares: []string{"JWTAuth"},
			Description: "统一资源操作",
		},
		{
			Method:      "GET",
			Path:        "/health",
			Handler:     "HealthCheck",
			Description: "健康检查",
		},
		{
			Method:      "GET",
			Path:        "/debug/ip",
			Handler:     "IPInfo",
			Description: "IP 信息查看",
		},
		{
			Method:      "GET",
			Path:        "/swagger/*any",
			Handler:     "SwaggerUI",
			Description: "Swagger API 文档",
		},
	}
}

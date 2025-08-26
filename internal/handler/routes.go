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
	ipManager *middleware.IPWhitelistManager,
	apiKeyManager *middleware.APIKeyManager,
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

	// 系统健康检查路由（公开访问）
	auth := v1.Group("/auth")
	{
		// 公开的健康检查端点
		auth.POST("/health", handlers.Auth.Health)
	}

	// 统一资源路由（需要 API Key 认证）
	resource := v1.Group("/resource")
	resource.Use(middleware.SimpleAuthMiddleware(apiKeyManager, true))
	{
		resource.POST("", handlers.Resource.HandleResource)
	}

	// SQL 操作路由（需要认证和相应权限）
	if handlers.SQL != nil {
		sql := v1.Group("/sql")
		sql.Use(middleware.SimpleAuthMiddleware(apiKeyManager, true))
		{
			// 通用 SQL 查询端点
			sql.POST("", handlers.SQL.HandleSQL)

			// 批量 SQL 操作端点
			sql.POST("/batch", handlers.SQL.HandleBatchSQL)

			// 便捷插入端点
			sql.POST("/insert", handlers.SQL.HandleInsertSQL)

			// 批量插入端点
			sql.POST("/batch-insert", handlers.SQL.HandleBatchInsert)
		}
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



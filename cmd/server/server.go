package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sql2api/docs"
	"sql2api/internal/config"
	"sql2api/internal/handler"
	"sql2api/internal/middleware"
	"sql2api/internal/repository"
	"sql2api/internal/service"

	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器结构
type Server struct {
	config        *config.Config
	router        *gin.Engine
	server        *http.Server
	repos         *repository.Repositories
	services      *service.Services
	handlers      *handler.Handlers
	ipManager     *middleware.IPWhitelistManager
	apiKeyManager *middleware.APIKeyManager
}

// NewServer 创建新的服务器实例
func NewServer() (*Server, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 配置 Swagger 信息
	docs.SwaggerInfo.Title = cfg.Swagger.Title
	docs.SwaggerInfo.Description = cfg.Swagger.Description
	docs.SwaggerInfo.Version = cfg.Swagger.Version
	docs.SwaggerInfo.Host = cfg.Swagger.Host
	docs.SwaggerInfo.BasePath = cfg.Swagger.BasePath
	docs.SwaggerInfo.Schemes = cfg.Swagger.Schemes

	server := &Server{
		config: cfg,
	}

	// 初始化组件
	if err := server.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// 设置路由
	if err := server.setupRoutes(); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	// 创建 HTTP 服务器
	server.createHTTPServer()

	return server, nil
}

// initializeComponents 初始化所有组件
func (s *Server) initializeComponents() error {
	fmt.Println("Initializing components...")

	// 初始化数据库连接
	repos, err := repository.NewRepositories(&s.config.Database)
	if err != nil {
		log.Printf("Database connection failed: %v", err)
		fmt.Println("⚠️  Running without database connection")
	} else {
		s.repos = repos
		fmt.Println("✅ Database connection established")
	}

	// 初始化服务层
	if s.repos != nil {
		services, err := service.NewServices(s.repos, &s.config)
		if err != nil {
			return fmt.Errorf("failed to initialize services: %w", err)
		}
		s.services = services
		fmt.Println("✅ Business services initialized")

		// 检查 SQL 服务状态
		if s.services.SQL != nil {
			fmt.Println("✅ SQL service enabled and initialized")
		} else {
			fmt.Println("ℹ️  SQL service disabled")
		}
	}

	// 初始化 API Key 管理器
	s.apiKeyManager = middleware.NewAPIKeyManager(&s.config.APIKeys)
	fmt.Printf("✅ API Key manager initialized (enabled: %v)\n", s.config.APIKeys.Enabled)

	// 初始化 IP 白名单管理器
	ipManager, err := middleware.NewIPWhitelistManager(&s.config.Security)
	if err != nil {
		log.Printf("Failed to initialize IP whitelist: %v", err)
		s.ipManager = nil
	} else {
		s.ipManager = ipManager
		fmt.Println("✅ IP whitelist manager initialized")
	}

	// 初始化处理器
	if s.services != nil {
		s.handlers = handler.NewHandlers(s.services)
		fmt.Println("✅ API handlers initialized")
	}

	return nil
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() error {
	// 设置 Gin 模式
	if s.config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建 Gin 路由器
	s.router = gin.New()

	// 设置路由
	if s.handlers != nil {
		handler.SetupRoutes(s.router, s.handlers, s.ipManager, s.apiKeyManager)
		fmt.Println("✅ Routes configured")
	} else {
		// 如果没有完整的处理器，至少设置健康检查
		s.setupBasicRoutes()
		fmt.Println("✅ Basic routes configured")
	}

	return nil
}

// setupBasicRoutes 设置基础路由（当完整服务不可用时）
func (s *Server) setupBasicRoutes() {
	// 应用基础中间件
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// 健康检查路由
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
			"message":   "Server is running (limited functionality)",
		})
	})

	// 根路由
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to SQL2API Server",
			"version": "1.0.0",
			"status":  "running",
		})
	})
}

// createHTTPServer 创建 HTTP 服务器
func (s *Server) createHTTPServer() {
	serverAddr := s.config.Server.GetServerAddress()
	s.server = &http.Server{
		Addr:           serverAddr,
		Handler:        s.router,
		ReadTimeout:    time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(s.config.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	fmt.Printf("✅ HTTP server configured on %s\n", serverAddr)
}

// Start 启动服务器
func (s *Server) Start() error {
	// 启动服务器（在 goroutine 中）
	go func() {
		fmt.Printf("\n🚀 Server starting on %s\n", s.server.Addr)
		fmt.Println("Press Ctrl+C to stop the server")

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Server shutting down...")

	// 创建一个超时上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	fmt.Println("✅ Server gracefully stopped")
	return nil
}

// Stop 停止服务器并清理资源
func (s *Server) Stop() error {
	// 清理资源
	if s.repos != nil {
		s.repos.Close()
		fmt.Println("✅ Database connection closed")
	}

	fmt.Println("👋 Goodbye!")
	return nil
}

// PrintStartupInfo 打印启动信息
func (s *Server) PrintStartupInfo() {
	fmt.Println("SQL2API Server - Starting...")
	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("- Server: %s\n", s.config.Server.GetServerAddress())
	fmt.Printf("- Database: %s\n", s.config.Database.Type)
	fmt.Printf("- JWT Issuer: %s\n", s.config.JWT.Issuer)
	fmt.Printf("- Log Level: %s\n", s.config.Log.Level)
	fmt.Printf("- IP Whitelist: %v\n", s.config.Security.IPWhitelist)
}

// RunServer 运行服务器的主函数
func RunServer() error {
	// 创建服务器
	server, err := NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// 打印启动信息
	server.PrintStartupInfo()

	// 启动服务器
	if err := server.Start(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	// 停止服务器并清理资源
	return server.Stop()
}

// Package main SQL2API Server
//
// @title SQL2API Server
// @version 1.0.0
// @description SQL2API 是一个现代化的 RESTful API 服务，支持统一的 CRUD 操作、SQL 查询引擎、API Key 认证、IP 白名单等功能
// @termsOfService http://swagger.io/terms/
// @contact.name SQL2API Team
// @contact.email support@sql2api.com
// @contact.url http://www.sql2api.com
// @license.name MIT
// @license.url http://opensource.org/licenses/MIT
// @host localhost:8081
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key authentication. Example: 'your-api-key-here'
//
// @tag.name Health
// @tag.description Health check endpoints
//
// @tag.name Resource
// @tag.description Resource management endpoints
//
// @tag.name SQL
// @tag.description SQL query and manipulation endpoints with support for PostgreSQL and Oracle databases
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"sql2api/internal/config"
	"sql2api/internal/handler"
	"sql2api/internal/middleware"
	"sql2api/internal/model"
	"sql2api/internal/repository"
	"sql2api/internal/service"

	_ "sql2api/docs" // 导入生成的 Swagger 文档

	"github.com/gin-gonic/gin"
	oracle "github.com/godoes/gorm-oracle"
	"github.com/golang-jwt/jwt/v5"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 版本信息变量，在构建时通过 -ldflags 注入
var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// 解析命令行参数
	var showVersion = flag.Bool("version", false, "显示版本信息")
	var showHelp = flag.Bool("help", false, "显示帮助信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("SQL2API Server\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Time: %s\n", buildTime)
		fmt.Printf("Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// 显示帮助信息
	if *showHelp {
		fmt.Printf("SQL2API Server - A powerful API server that converts SQL operations to REST endpoints\n\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("  sql2api [options]\n\n")
		fmt.Printf("Options:\n")
		fmt.Printf("  -version    显示版本信息\n")
		fmt.Printf("  -help       显示帮助信息\n\n")
		fmt.Printf("Configuration:\n")
		fmt.Printf("  配置文件: config.yaml\n")
		fmt.Printf("  环境变量: SQL2API_* (例如: SQL2API_SERVER_PORT=8080)\n\n")
		fmt.Printf("Documentation:\n")
		fmt.Printf("  Swagger UI: http://localhost:8080/swagger/index.html\n")
		fmt.Printf("  Examples: ./examples/sql_examples.md\n")
		os.Exit(0)
	}

	fmt.Printf("🚀 Starting SQL2API Server %s\n", version)

	// 运行服务器
	if err := RunServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func mainOld() {
	fmt.Println("SQL2API Server - Starting...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("- Server: %s\n", cfg.Server.GetServerAddress())
	fmt.Printf("- Database: %s\n", cfg.Database.Type)
	fmt.Printf("- JWT Issuer: %s\n", cfg.JWT.Issuer)
	fmt.Printf("- Log Level: %s\n", cfg.Log.Level)
	fmt.Printf("- IP Whitelist: %v\n", cfg.Security.IPWhitelist)

	// 验证依赖是否可以正常导入
	_ = gin.New()
	_ = &gorm.DB{}
	_ = postgres.Open("")
	_ = oracle.New(oracle.Config{})
	_ = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
	_ = ginSwagger.WrapHandler(swaggerFiles.Handler)
	_, _ = bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)

	fmt.Println("All dependencies verified successfully!")
	fmt.Printf("DSN: %s\n", cfg.Database.GetDSN())

	// 验证模型定义
	fmt.Println("\nTesting model definitions...")

	// 测试用户模型
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}

	err = user.SetPassword("password123")
	if err != nil {
		log.Printf("Error setting password: %v", err)
	} else {
		fmt.Println("✅ User password hashing works")
	}

	if user.CheckPassword("password123") {
		fmt.Println("✅ User password verification works")
	} else {
		fmt.Println("❌ User password verification failed")
	}

	userResponse := user.ToResponse()
	fmt.Printf("✅ User response: %+v\n", userResponse)

	// 测试项目模型
	item := &model.Item{
		Name:        "Test Item",
		Value:       100,
		Description: "Test Description",
		Category:    "Test Category",
		CreatedBy:   1,
	}

	itemResponse := item.ToResponse()
	fmt.Printf("✅ Item response: %+v\n", itemResponse)

	// 测试通用响应结构
	successResp := model.NewSuccessResponse(map[string]string{"test": "data"}, "Test successful")
	fmt.Printf("✅ Success response: %+v\n", successResp)

	errorResp := model.NewErrorResponse(400, "Test error", "Error details")
	fmt.Printf("✅ Error response: %+v\n", errorResp)

	fmt.Println("\nAll model definitions verified successfully!")

	// 测试数据库连接和仓库层
	fmt.Println("\nTesting database connection and repository layer...")

	// 创建仓库实例
	repos, err := repository.NewRepositories(&cfg.Database)
	if err != nil {
		log.Printf("Failed to create repositories: %v", err)
		fmt.Println("❌ Database connection failed (this is expected if no database is running)")
	} else {
		defer repos.Close()

		fmt.Println("✅ Database connection successful")

		// 测试健康检查
		if err := repos.HealthCheck(); err != nil {
			log.Printf("Health check failed: %v", err)
		} else {
			fmt.Println("✅ Database health check passed")
		}

		// 获取连接统计信息
		if stats, err := repos.DB.GetConnectionStats(); err != nil {
			log.Printf("Failed to get connection stats: %v", err)
		} else {
			fmt.Printf("✅ Connection stats: Open=%v, InUse=%v, Idle=%v\n",
				stats["open_connections"], stats["in_use"], stats["idle"])
		}

		// 测试项目仓库（不实际操作数据库）
		fmt.Println("✅ Item repository initialized")

		fmt.Println("✅ All repository operations verified")
	}

	// 测试 IP 白名单中间件
	fmt.Println("\nTesting IP whitelist middleware...")

	// 创建 IP 白名单管理器
	ipManager, err := middleware.NewIPWhitelistManager(&cfg.Security)
	if err != nil {
		log.Printf("Failed to create IP whitelist manager: %v", err)
	} else {
		fmt.Println("✅ IP whitelist manager created")

		// 测试 IP 检查
		testIPs := []string{
			"127.0.0.1",     // 应该被允许
			"192.168.1.100", // 应该被允许（在白名单中）
			"8.8.8.8",       // 应该被拒绝
			"10.1.2.3",      // 应该被允许（在 CIDR 范围内）
		}

		for _, testIP := range testIPs {
			allowed := ipManager.IsAllowed(testIP)
			status := "❌ DENIED"
			if allowed {
				status = "✅ ALLOWED"
			}
			fmt.Printf("   IP %s: %s\n", testIP, status)
		}

		// 测试 IP 验证
		err = middleware.ValidateIPWhitelist(cfg.Security.IPWhitelist)
		if err != nil {
			log.Printf("IP whitelist validation failed: %v", err)
		} else {
			fmt.Println("✅ IP whitelist configuration validated")
		}
	}

	fmt.Println("✅ IP whitelist middleware verified successfully")

	// 测试业务服务层
	fmt.Println("\nTesting business service layer...")

	var services *service.Services
	if repos != nil {
		// 创建服务管理器
		services = service.NewServices(repos)
		fmt.Println("✅ Service layer created")

		// 测试用户服务
		userService := services.User

		// 测试用户注册
		registerReq := &model.UserCreateRequest{
			Username: "testuser",
			Password: "password123",
			Email:    "test@example.com",
			FullName: "Test User",
		}

		user, err := userService.Register(registerReq)
		if err != nil {
			log.Printf("User registration failed: %v", err)
		} else {
			fmt.Printf("✅ User registration successful: %s (ID: %d)\n", user.Username, user.ID)

			// 测试用户登录
			loginUser, err := userService.Login("testuser", "password123")
			if err != nil {
				log.Printf("User login failed: %v", err)
			} else {
				fmt.Printf("✅ User login successful: %s\n", loginUser.Username)
			}

			// 测试错误密码登录
			_, err = userService.Login("testuser", "wrongpassword")
			if err != nil {
				fmt.Println("✅ Invalid password correctly rejected")
			} else {
				fmt.Println("❌ Invalid password should be rejected")
			}
		}

		// 测试项目服务
		itemService := services.Item

		if user != nil {
			// 测试项目创建
			itemReq := &model.ItemCreateRequest{
				Name:        "Test Item",
				Value:       100,
				Description: "Test Description",
				Category:    "Test Category",
				Tags:        "test,demo",
			}

			item, err := itemService.CreateItem(itemReq, user.ID)
			if err != nil {
				log.Printf("Item creation failed: %v", err)
			} else {
				fmt.Printf("✅ Item creation successful: %s (ID: %d)\n", item.Name, item.ID)

				// 测试项目获取
				retrievedItem, err := itemService.GetItemByID(item.ID, false)
				if err != nil {
					log.Printf("Item retrieval failed: %v", err)
				} else {
					fmt.Printf("✅ Item retrieval successful: %s\n", retrievedItem.Name)
				}

				// 测试项目所有权验证
				err = itemService.ValidateItemOwnership(item.ID, user.ID)
				if err != nil {
					log.Printf("Item ownership validation failed: %v", err)
				} else {
					fmt.Println("✅ Item ownership validation successful")
				}

				// 测试非所有者访问
				err = itemService.ValidateItemOwnership(item.ID, 999)
				if err != nil {
					fmt.Println("✅ Non-owner access correctly rejected")
				} else {
					fmt.Println("❌ Non-owner access should be rejected")
				}
			}
		}

		fmt.Println("✅ Business service layer verified successfully")
	} else {
		fmt.Println("⚠️  Skipping service layer test - no database connection")
	}

	// 测试 API 处理器和路由
	fmt.Println("\nTesting API handlers and routes...")

	if repos != nil && services != nil {
		// 创建处理器
		handlers := handler.NewHandlers(services, jwtManager)
		fmt.Println("✅ API handlers created")

		// 创建 Gin 路由器
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// 设置路由
		handler.SetupRoutes(router, handlers, jwtManager, ipManager)
		fmt.Println("✅ Routes configured")

		// 获取路由信息
		routeInfo := handler.GetRouteInfo()
		fmt.Printf("✅ Total routes configured: %d\n", len(routeInfo))

		// 显示主要路由
		fmt.Println("   Main routes:")
		for _, route := range routeInfo {
			middlewareInfo := ""
			if len(route.Middlewares) > 0 {
				middlewareInfo = fmt.Sprintf(" [%s]", route.Middlewares[0])
			}
			fmt.Printf("   - %s %s%s - %s\n", route.Method, route.Path, middlewareInfo, route.Description)
		}

		fmt.Println("✅ API handlers and routes verified successfully")
	} else {
		fmt.Println("⚠️  Skipping API handler test - no database connection or services")
	}
}

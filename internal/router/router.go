package router

import (
	"gpt-load/internal/handler"
	"gpt-load/internal/i18n"
	"gpt-load/internal/middleware"
	"gpt-load/internal/proxy"
	"gpt-load/internal/services"
	"gpt-load/internal/types"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"

	"github.com/gin-gonic/gin"
)


func NewRouter(
	serverHandler *handler.Server,
	proxyServer *proxy.ProxyServer,
	configManager types.ConfigManager,
	groupManager *services.GroupManager,
	incrementalValidationHandler *handler.IncrementalValidationHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// 注册全局中间件
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.Logger(configManager.GetLogConfig()))
	router.Use(middleware.CORS(configManager.GetCORSConfig()))
	router.Use(middleware.RateLimiter(configManager.GetPerformanceConfig()))
	router.Use(middleware.SecurityHeaders())
	startTime := time.Now()
	router.Use(func(c *gin.Context) {
		c.Set("serverStartTime", startTime)
		c.Next()
	})

	// 注册路由
	registerSystemRoutes(router, serverHandler)
	registerAPIRoutes(router, serverHandler, configManager, incrementalValidationHandler)
	registerProxyRoutes(router, proxyServer, groupManager)

	// 添加全局中间件和错误处理
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	return router
}

// registerSystemRoutes 注册系统级路由
func registerSystemRoutes(router *gin.Engine, serverHandler *handler.Server) {
	router.GET("/health", serverHandler.Health)
}

// registerAPIRoutes 注册API路由
func registerAPIRoutes(
	router *gin.Engine,
	serverHandler *handler.Server,
	configManager types.ConfigManager,
	incrementalValidationHandler *handler.IncrementalValidationHandler,
) {
	api := router.Group("/api")
	api.Use(i18n.Middleware())

	authConfig := configManager.GetAuthConfig()

	// 公开
	registerPublicAPIRoutes(api, serverHandler)

	// 认证
	protectedAPI := api.Group("")
	protectedAPI.Use(middleware.Auth(authConfig))
	registerProtectedAPIRoutes(protectedAPI, serverHandler, incrementalValidationHandler)
}

// registerPublicAPIRoutes 公开API路由
func registerPublicAPIRoutes(api *gin.RouterGroup, serverHandler *handler.Server) {
	api.POST("/auth/login", serverHandler.Login)
}

// registerProtectedAPIRoutes 认证API路由
func registerProtectedAPIRoutes(api *gin.RouterGroup, serverHandler *handler.Server, incrementalValidationHandler *handler.IncrementalValidationHandler) {
	api.GET("/channel-types", serverHandler.CommonHandler.GetChannelTypes)

	groups := api.Group("/groups")
	{
		groups.POST("", serverHandler.CreateGroup)
		groups.GET("", serverHandler.ListGroups)
		groups.GET("/list", serverHandler.List)
		groups.GET("/config-options", serverHandler.GetGroupConfigOptions)
		groups.PUT("/:id", serverHandler.UpdateGroup)
		groups.DELETE("/:id", serverHandler.DeleteGroup)
		groups.GET("/:id/stats", serverHandler.GetGroupStats)
		groups.POST("/:id/copy", serverHandler.CopyGroup)
	}

	// Key Management Routes
	keys := api.Group("/keys")
	{
		keys.GET("", serverHandler.ListKeysInGroup)
		keys.GET("/export", serverHandler.ExportKeys)
		keys.POST("/add-multiple", serverHandler.AddMultipleKeys)
		keys.POST("/add-async", serverHandler.AddMultipleKeysAsync)
		keys.POST("/delete-multiple", serverHandler.DeleteMultipleKeys)
		keys.POST("/delete-async", serverHandler.DeleteMultipleKeysAsync)
		keys.POST("/restore-multiple", serverHandler.RestoreMultipleKeys)
		keys.POST("/restore-all-invalid", serverHandler.RestoreAllInvalidKeys)
		keys.POST("/clear-all-invalid", serverHandler.ClearAllInvalidKeys)
		keys.POST("/clear-all", serverHandler.ClearAllKeys)
		keys.POST("/validate-group", serverHandler.ValidateGroupKeys)
		keys.POST("/test-multiple", serverHandler.TestMultipleKeys)
	}

	// Tasks
	api.GET("/tasks/status", serverHandler.GetTaskStatus)

	// 仪表板和日志
	dashboard := api.Group("/dashboard")
	{
		dashboard.GET("/stats", serverHandler.Stats)
		dashboard.GET("/chart", serverHandler.Chart)
		dashboard.GET("/encryption-status", serverHandler.EncryptionStatus)
	}

	// 日志
	logs := api.Group("/logs")
	{
		logs.GET("", serverHandler.GetLogs)
		logs.GET("/export", serverHandler.ExportLogs)
	}

	// 设置
	settings := api.Group("/settings")
	{
		settings.GET("", serverHandler.GetSettings)
		settings.PUT("", serverHandler.UpdateSettings)
	}

	// 增量验证
	validation := api.Group("/validation")
	{
		validation.POST("/groups/:groupId", incrementalValidationHandler.ValidateGroup)
		validation.POST("/groups", incrementalValidationHandler.ValidateAllGroups)
		validation.GET("/groups/:groupId/history", incrementalValidationHandler.GetValidationHistory)
	}
}

// registerProxyRoutes 注册代理路由
func registerProxyRoutes(
	router *gin.Engine,
	proxyServer *proxy.ProxyServer,
	groupManager *services.GroupManager,
) {
	proxyGroup := router.Group("/proxy")

	proxyGroup.Use(middleware.ProxyAuth(groupManager))

	proxyGroup.Any("/:group_name/*path", proxyServer.HandleProxy)
}

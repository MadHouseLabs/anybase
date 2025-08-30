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

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	v1 "github.com/madhouselabs/anybase/api/v1"
	// _ "github.com/madhouselabs/anybase/docs" // Commented out for Docker build
	"github.com/madhouselabs/anybase/internal/accesskey"
	"github.com/madhouselabs/anybase/internal/auth"
	"github.com/madhouselabs/anybase/internal/collection"
	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/database"
	"github.com/madhouselabs/anybase/internal/governance"
	"github.com/madhouselabs/anybase/internal/middleware"
	"github.com/madhouselabs/anybase/internal/settings"
	"github.com/madhouselabs/anybase/internal/user"
)

// @title AnyBase API
// @version 1.0
// @description Firebase-like API layer on AWS DocumentDB
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@anybase.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	db := database.GetDB()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Close(ctx); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Create indexes
	ctx := context.Background()
	if err := db.CreateIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	// Initialize repositories and services
	userRepo := user.NewRepository(db)
	authService := auth.NewService(userRepo, &cfg.Auth)
	rbacService := governance.NewRBACService(db)
	collectionService := collection.NewService(db, rbacService)
	accessKeyRepo := accesskey.NewRepository(db)
	settingsService := settings.NewService(db.GetDatabase())

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.Auth, rbacService)
	accessKeyMiddleware := middleware.NewAccessKeyAuthMiddleware(accessKeyRepo)
	rateLimiter := middleware.NewPerIPRateLimiter(100, 10) // 100 requests per second, burst of 10

	// Setup Gin router
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORS())
	router.Use(rateLimiter.Limit())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		if err := db.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "Database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// MCP endpoint
	mcpHandler := v1.NewMCPHandler(collectionService)
	router.POST("/mcp", accessKeyMiddleware.Authenticate(), authMiddleware.RequireAuth(), mcpHandler.HandleMCPRequest)

	// API routes
	setupAPIRoutes(router, authService, authMiddleware, accessKeyMiddleware, rbacService, collectionService, userRepo, accessKeyRepo, settingsService)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupAPIRoutes(router *gin.Engine, authService auth.Service, authMiddleware *middleware.AuthMiddleware, accessKeyMiddleware *middleware.AccessKeyAuthMiddleware, rbacService governance.RBACService, collectionService collection.Service, userRepo user.Repository, accessKeyRepo accesskey.Repository, settingsService settings.Service) {
	// API v1 group
	api := router.Group("/api/v1")

	// Auth endpoints (public)
	authHandler := v1.NewAuthHandler(authService)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.GET("/verify-email", authHandler.VerifyEmail)
		authGroup.POST("/request-password-reset", authHandler.RequestPasswordReset)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	// Protected auth endpoints
	authProtected := api.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth())
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.POST("/change-password", authHandler.ChangePassword)
	}

	// User endpoints (protected)
	userHandler := v1.NewUserHandler(userRepo)
	userGroup := api.Group("/users")
	userGroup.Use(authMiddleware.RequireAuth())
	{
		userGroup.GET("/profile", userHandler.GetUserProfile)
		userGroup.PUT("/profile", userHandler.UpdateUserProfile)
	}

	// Collection & View endpoints (support both JWT and Access Key auth)
	collectionHandler := v1.NewCollectionHandler(collectionService)
	collectionsGroup := api.Group("/collections")
	collectionsGroup.Use(accessKeyMiddleware.Authenticate()) // Try access key first
	collectionsGroup.Use(authMiddleware.RequireAuth())
	{
		collectionsGroup.POST("", collectionHandler.CreateCollection)
		collectionsGroup.GET("", collectionHandler.ListCollections)
		collectionsGroup.GET("/:name", collectionHandler.GetCollection)
		collectionsGroup.PUT("/:name", collectionHandler.UpdateCollection)
		collectionsGroup.DELETE("/:name", collectionHandler.DeleteCollection)
		
		// Index management
		collectionsGroup.GET("/:name/indexes", collectionHandler.ListIndexes)
		collectionsGroup.POST("/:name/indexes", collectionHandler.CreateIndex)
		collectionsGroup.DELETE("/:name/indexes/:index", collectionHandler.DeleteIndex)
	}

	// View endpoints (support both JWT and Access Key auth)
	viewsGroup := api.Group("/views")
	viewsGroup.Use(accessKeyMiddleware.Authenticate())
	viewsGroup.Use(authMiddleware.RequireAuth())
	{
		viewsGroup.POST("", collectionHandler.CreateView)
		viewsGroup.GET("", collectionHandler.ListViews)
		viewsGroup.GET("/:name", collectionHandler.GetView)
		viewsGroup.GET("/:name/query", collectionHandler.QueryView)
		viewsGroup.PUT("/:name", collectionHandler.UpdateView)
		viewsGroup.DELETE("/:name", collectionHandler.DeleteView)
	}

	// Data endpoints (support both JWT and Access Key auth)
	dataGroup := api.Group("/data")
	dataGroup.Use(accessKeyMiddleware.Authenticate())
	dataGroup.Use(authMiddleware.RequireAuth())
	{
		dataGroup.GET("/:collection", collectionHandler.QueryDocuments)
		dataGroup.POST("/:collection", collectionHandler.InsertDocument)
		dataGroup.GET("/:collection/:id", collectionHandler.GetDocument)
		dataGroup.PUT("/:collection/:id", collectionHandler.UpdateDocument)
		dataGroup.DELETE("/:collection/:id", collectionHandler.DeleteDocument)
	}


	// Access key management (accessible by admin and developer)
	accessKeyHandler := v1.NewAccessKeyHandler(accessKeyRepo)
	accessKeysGroup := api.Group("/access-keys")
	accessKeysGroup.Use(authMiddleware.RequireRole("admin", "developer"))
	{
		accessKeysGroup.POST("", accessKeyHandler.CreateAccessKey)
		accessKeysGroup.GET("", accessKeyHandler.ListAccessKeys)
		accessKeysGroup.GET("/:id", accessKeyHandler.GetAccessKey)
		accessKeysGroup.PUT("/:id", accessKeyHandler.UpdateAccessKey)
		accessKeysGroup.DELETE("/:id", accessKeyHandler.DeleteAccessKey)
		accessKeysGroup.POST("/:id/regenerate", accessKeyHandler.RegenerateAccessKey)
	}

	// User management endpoints
	// Read operations - accessible by admin and developer
	usersReadGroup := api.Group("/admin/users")
	usersReadGroup.Use(authMiddleware.RequireRole("admin", "developer"))
	{
		usersReadGroup.GET("", userHandler.ListUsers)
		usersReadGroup.GET("/:id", userHandler.GetUser)
	}

	// Write operations - admin only
	usersWriteGroup := api.Group("/admin/users")
	usersWriteGroup.Use(authMiddleware.RequireRole("admin"))
	{
		usersWriteGroup.POST("", userHandler.CreateUser)
		usersWriteGroup.PUT("/:id", userHandler.UpdateUser)
		usersWriteGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// Settings endpoints
	settingsHandler := v1.NewSettingsHandler(settingsService)
	settingsGroup := api.Group("/settings")
	settingsGroup.Use(authMiddleware.RequireAuth())
	{
		settingsGroup.GET("/user", settingsHandler.GetUserSettings)
		settingsGroup.PUT("/user", settingsHandler.UpdateUserSettings)
		settingsGroup.GET("/system", settingsHandler.GetSystemSettings)
	}
	
	// System settings update - admin only
	settingsAdminGroup := api.Group("/settings")
	settingsAdminGroup.Use(authMiddleware.RequireRole("admin"))
	{
		settingsAdminGroup.PUT("/system", settingsHandler.UpdateSystemSettings)
	}
}


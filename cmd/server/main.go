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
	v1 "github.com/madhouselabs/anybase/api/v1"
	"github.com/madhouselabs/anybase/internal/accesskey"
	"github.com/madhouselabs/anybase/internal/auth"
	"github.com/madhouselabs/anybase/internal/collection"
	"github.com/madhouselabs/anybase/internal/config"
	"github.com/madhouselabs/anybase/internal/database"
	"github.com/madhouselabs/anybase/internal/governance"
	"github.com/madhouselabs/anybase/internal/middleware"
	"github.com/madhouselabs/anybase/internal/settings"
	"github.com/madhouselabs/anybase/internal/user"
	"github.com/madhouselabs/anybase/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


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
	
	// Initialize admin user if needed
	if err := initializeAdminUser(ctx, userRepo, authService); err != nil {
		log.Printf("Warning: Failed to initialize admin user: %v", err)
	}
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

// initializeAdminUser creates an initial admin user if no admin users exist
func initializeAdminUser(ctx context.Context, userRepo user.Repository, authService auth.Service) error {
	// Check if any admin users exist
	filter := bson.M{"role": "admin", "deleted_at": nil}
	users, err := userRepo.List(ctx, filter, nil)
	if err != nil && err != mongo.ErrNoDocuments {
		return fmt.Errorf("failed to check for existing admin users: %w", err)
	}
	
	// If admin users already exist, skip initialization
	if len(users) > 0 {
		log.Println("Admin user already exists, skipping initialization")
		return nil
	}
	
	// Get admin credentials from environment variables
	adminEmail := os.Getenv("INIT_ADMIN_EMAIL")
	adminPassword := os.Getenv("INIT_ADMIN_PASSWORD")
	
	// If not set, use defaults (but log a warning)
	if adminEmail == "" {
		adminEmail = "admin@anybase.local"
		log.Println("Warning: INIT_ADMIN_EMAIL not set, using default: admin@anybase.local")
	}
	
	if adminPassword == "" {
		adminPassword = "admin123"
		log.Println("Warning: INIT_ADMIN_PASSWORD not set, using default: admin123")
		log.Println("IMPORTANT: Please change the admin password after first login!")
	}
	
	// Create the admin user using UserRegistration model
	adminUser, err := authService.Register(ctx, &models.UserRegistration{
		Email:     adminEmail,
		Password:  adminPassword,
		FirstName: "Admin",
		LastName:  "User",
	})
	
	if err != nil {
		return fmt.Errorf("failed to create initial admin user: %w", err)
	}
	
	// Update the user role to admin
	updateData := bson.M{"$set": bson.M{"role": "admin"}}
	if err := userRepo.UpdateRaw(ctx, adminUser.ID, updateData); err != nil {
		return fmt.Errorf("failed to set admin role: %w", err)
	}
	
	log.Printf("Successfully created initial admin user: %s", adminEmail)
	log.Println("Please change the admin password after first login!")
	
	return nil
}


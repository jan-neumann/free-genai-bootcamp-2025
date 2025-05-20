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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"lang-portal/backend_go/internal/api"
	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/database"
	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/repository"
	"lang-portal/backend_go/internal/service"
)

const (
	defaultPort = "8081"
	dbPath      = "words.db"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize database
	db, err := initDatabase(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	wordRepo := repository.NewWordRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	studyRepo := repository.NewStudyRepository(db)

	// Initialize services
	baseService := service.NewBaseService(wordRepo, groupRepo, studyRepo)
	dashboardService := service.NewDashboardService(baseService)
	wordService := service.NewWordService(baseService)
	groupService := service.NewGroupService(baseService)
	studyService := service.NewStudyService(baseService)

	// Initialize router with middleware
	router := gin.New() // Use gin.New() instead of gin.Default() to have more control over middleware

	// Add security and stability middleware
	router.Use(middleware.Recovery())                // Handle panics
	router.Use(middleware.SecurityHeaders())         // Add security headers
	router.Use(middleware.CORS())                    // Handle CORS
	router.Use(middleware.RequestLogger())           // Log requests
	router.Use(middleware.RateLimit(100, 200))       // Rate limit: 100 requests per second, burst of 200
	router.Use(middleware.Timeout(30 * time.Second)) // Request timeout
	router.Use(gin.Logger())                         // Gin's built-in logger

	// Register API routes
	api.RegisterRoutes(router, &api.Services{
		Dashboard: dashboardService,
		Word:      wordService,
		Group:     groupService,
		Study:     studyService,
	})

	// Create HTTP server with timeouts
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exiting")
}

func initDatabase(logger *log.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Run seeds if database is empty
	var count int64
	if err := db.Model(&models.Word{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check word count: %w", err)
	}
	if count == 0 {
		if err := database.Seed(db); err != nil {
			return nil, fmt.Errorf("failed to run seeds: %w", err)
		}
	}

	return db, nil
}

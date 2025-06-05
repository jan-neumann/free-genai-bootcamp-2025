package api

import (
	"lang-portal/backend_go/internal/repository"
	"lang-portal/backend_go/internal/routes"
	"lang-portal/backend_go/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter configures and returns a new Gin engine with all routes
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Initialize repositories
	wordRepo := repository.NewWordRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	studyRepo := repository.NewStudyRepository(db)

	// Initialize services
	baseService := service.NewBaseService(wordRepo, groupRepo, studyRepo)
	dashboardService := service.NewDashboardService(baseService)
	studyService := service.NewStudyService(baseService)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Register dashboard routes
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/last-session", GetLastStudySession(dashboardService))
			dashboard.GET("/progress", GetStudyProgress(dashboardService))
			dashboard.GET("/stats", GetQuickStats(dashboardService))
		}

		// Register study routes
		routes.RegisterStudyRoutes(router, db, studyService)

		// Register group routes
		routes.RegisterGroupRoutes(router, db)
	}

	return router
}

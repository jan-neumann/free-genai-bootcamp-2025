package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/api/handlers"
	"lang-portal/backend_go/internal/service"
)

// RegisterStudyRoutes registers all study-related routes
func RegisterStudyRoutes(router *gin.Engine, db *gorm.DB, studyService *service.StudyService) {
	studyHandler := handlers.NewStudyHandler(studyService, db)

	// Study activities routes
	activities := router.Group("/api/study/activities")
	{
		activities.GET("", studyHandler.GetStudyActivities)
		activities.GET("/:id", studyHandler.GetStudyActivity)
	}

	// Study sessions routes
	sessions := router.Group("/api/study/sessions")
	{
		sessions.POST("", handlers.CreateStudySession(studyService))
		sessions.GET("/:id", studyHandler.GetStudySession)
		sessions.POST("/:id/reviews", studyHandler.AddWordReview)
	}

	// Group study sessions routes
	router.GET("/api/groups/:group_id/sessions", studyHandler.GetGroupStudySessions)
}

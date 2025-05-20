package api

import (
	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/service"

	"github.com/gin-gonic/gin"
)

// Services holds all service instances used by the API handlers
type Services struct {
	Dashboard *service.DashboardService
	Word      *service.WordService
	Group     *service.GroupService
	Study     *service.StudyService
}

// RegisterRoutes sets up all API routes and middleware
func RegisterRoutes(router *gin.Engine, services *Services) {
	// Create API group
	api := router.Group("/api")
	{
		// Register middleware
		api.Use(middleware.PaginationMiddleware())

		// Dashboard routes
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/last-session", GetLastStudySession(services.Dashboard))
			dashboard.GET("/progress", GetStudyProgress(services.Dashboard))
			dashboard.GET("/quick-stats", GetQuickStats(services.Dashboard))
		}

		// Word routes
		words := api.Group("/words")
		{
			words.GET("", ListWords(services.Word))
			words.GET("/:id", GetWord(services.Word))
			words.POST("", CreateWord(services.Word))
			words.PUT("/:id", UpdateWord(services.Word))
			words.DELETE("/:id", DeleteWord(services.Word))
			words.GET("/:id/groups", GetGroupsByWord(services.Group))
		}

		// Group routes
		groups := api.Group("/groups")
		{
			groups.GET("", ListGroups(services.Group))
			groups.GET("/:id", GetGroup(services.Group))
			groups.POST("", CreateGroup(services.Group))
			groups.PUT("/:id", UpdateGroup(services.Group))
			groups.DELETE("/:id", DeleteGroup(services.Group))
			groups.POST("/:id/words/:word_id", AddWordToGroup(services.Group))
			groups.DELETE("/:id/words/:word_id", RemoveWordFromGroup(services.Group))
			groups.GET("/:id/stats", GetGroupStudyStats(services.Group))
			groups.GET("/:id/words", GetWordsByGroup(services.Word))
		}

		// Study routes
		study := api.Group("/study")
		{
			// Study activities
			study.GET("/activities", ListStudyActivities(services.Study))
			study.GET("/activities/:id", GetStudyActivity(services.Study))
			study.POST("/activities", CreateStudyActivity(services.Study))

			// Study sessions
			study.GET("/sessions", ListStudySessions(services.Study))
			study.POST("/sessions", CreateStudySession(services.Study))
			study.GET("/sessions/:id", GetStudySession(services.Study))
			study.GET("/sessions/group/:group_id", GetStudySessionsByGroup(services.Study))
			study.GET("/sessions/activity/:activity_id", GetStudySessionsByActivity(services.Study))

			// Word reviews
			study.POST("/sessions/:id/reviews", AddWordReview(services.Study))
			study.GET("/sessions/:id/reviews", GetWordReviewsBySession(services.Study))

			// Study statistics
			study.GET("/stats", GetStudyStats(services.Study))
			study.GET("/streak", GetStudyStreak(services.Study))
			study.GET("/active-groups", GetActiveGroups(services.Study))
			study.POST("/reset", ResetStudyHistory(services.Study))
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
	}
}

package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/api/handlers"
)

// RegisterGroupRoutes registers all group-related routes
func RegisterGroupRoutes(router *gin.Engine, db *gorm.DB) {
	groupHandler := handlers.NewGroupHandler(db)
	groupDetailHandler := handlers.NewGroupDetailHandler(db)

	// Group routes
	groups := router.Group("/api/groups")
	{
		groups.GET("", groupHandler.GetGroups)
		groups.POST("", groupHandler.CreateGroup)
		groups.GET("/:id", groupHandler.GetGroup)
		groups.PUT("/:id", groupHandler.UpdateGroup)
		groups.DELETE("/:id", groupHandler.DeleteGroup)
		groups.POST("/:id/words", groupHandler.AddWordToGroup)
		groups.DELETE("/:id/words/:word_id", groupHandler.RemoveWordFromGroup)

		// Raw words endpoint
		groups.GET("/:id/raw", groupDetailHandler.GetGroupWordsRaw)
	}

	// Group study sessions route (kept for backward compatibility)
	router.GET("/api/groups/:group_id/sessions", groupDetailHandler.GetGroupStudySessions)
}

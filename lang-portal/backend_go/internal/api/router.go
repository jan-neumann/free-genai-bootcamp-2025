package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter is deprecated, use RegisterRoutes instead
func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	// This function is kept for backward compatibility
	// New code should use RegisterRoutes directly
	return router
}

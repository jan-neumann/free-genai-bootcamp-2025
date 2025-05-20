package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct {
	db *gorm.DB
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

// UpdateTheme updates the application theme
func (h *SettingsHandler) UpdateTheme(c *gin.Context) {
	var req struct {
		Theme string `json:"theme" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate theme
	allowedThemes := map[string]bool{
		"light":  true,
		"dark":   true,
		"system": true,
	}

	if !allowedThemes[req.Theme] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid theme. Allowed themes are: %v", allowedThemes),
		})
		return
	}

	// In a real app, you'd persist this to a settings table.
	// For now, just echo back the theme as per spec.
	c.JSON(http.StatusOK, gin.H{
		"settings": gin.H{
			"theme": req.Theme,
		},
	})
}

// ResetHistory resets study history while keeping words and groups
func (h *SettingsHandler) ResetHistory(c *gin.Context) {
	// Begin transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Delete all word reviews and study sessions in a transaction
	if err := tx.Exec("DELETE FROM word_review_items").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete word reviews"})
		return
	}

	if err := tx.Exec("DELETE FROM study_sessions").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete study sessions"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Study history has been reset successfully"})
}

// FullReset performs a complete reset of the application
func (h *SettingsHandler) FullReset(c *gin.Context) {
	// Begin transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Delete all data in reverse order of dependencies
	tables := []string{
		"word_review_items", // Delete reviews first
		"study_sessions",    // Then study sessions
		"word_groups",       // Then word-group associations
		"groups",            // Then groups
		"words",             // Finally words
	}

	for _, table := range tables {
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete data from " + table,
				"details": err.Error(),
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Application has been fully reset"})
}

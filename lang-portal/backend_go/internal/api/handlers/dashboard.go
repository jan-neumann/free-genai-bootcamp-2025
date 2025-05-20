package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/models"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	db *gorm.DB
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetLastStudySession returns information about the most recent study session
func (h *DashboardHandler) GetLastStudySession(c *gin.Context) {
	var session models.StudySession
	if err := h.db.Preload("Group").
		Preload("Activity").
		Order("created_at DESC").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"id":                0,
				"group_id":          0,
				"created_at":        time.Time{},
				"study_activity_id": 0,
				"group_name":        "",
				"activity_name":     "",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch last study session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                session.ID,
		"group_id":          session.GroupID,
		"created_at":        session.CreatedAt,
		"study_activity_id": session.StudyActivityID,
		"group_name":        session.Group.Name,
		"activity_name":     session.Activity.Name,
	})
}

// GetStudyProgress returns study progress statistics
func (h *DashboardHandler) GetStudyProgress(c *gin.Context) {
	// Get total words studied (unique words that have been reviewed)
	var totalWordsStudied int64
	if err := h.db.Model(&models.WordReview{}).
		Distinct("word_id").
		Count(&totalWordsStudied).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count studied words"})
		return
	}

	// Get total available words
	var totalAvailableWords int64
	if err := h.db.Model(&models.Word{}).Count(&totalAvailableWords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count total words"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_words_studied":   totalWordsStudied,
		"total_available_words": totalAvailableWords,
	})
}

// GetQuickStats returns quick overview statistics
func (h *DashboardHandler) GetQuickStats(c *gin.Context) {
	// Get total study sessions
	var totalStudySessions int64
	if err := h.db.Model(&models.StudySession{}).Count(&totalStudySessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count study sessions"})
		return
	}

	// Get total active groups (groups that have been used in study sessions)
	var totalActiveGroups int64
	if err := h.db.Model(&models.StudySession{}).
		Distinct("group_id").
		Count(&totalActiveGroups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count active groups"})
		return
	}

	// Calculate success rate
	var totalReviews int64
	var correctReviews int64
	if err := h.db.Model(&models.WordReview{}).Count(&totalReviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count total reviews"})
		return
	}
	if err := h.db.Model(&models.WordReview{}).Where("correct = ?", true).Count(&correctReviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count correct reviews"})
		return
	}

	successRate := 0
	if totalReviews > 0 {
		successRate = int((float64(correctReviews) / float64(totalReviews)) * 100)
	}

	// Calculate study streak
	var studyStreakDays int
	var lastStudyDate time.Time
	var currentStreak int

	// Get all study sessions ordered by date
	var sessions []models.StudySession
	if err := h.db.Order("created_at DESC").Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study sessions"})
		return
	}

	if len(sessions) > 0 {
		lastStudyDate = sessions[0].CreatedAt
		currentStreak = 1

		// Check for consecutive days
		for i := 1; i < len(sessions); i++ {
			currentDate := sessions[i].CreatedAt
			daysDiff := lastStudyDate.Sub(currentDate).Hours() / 24

			if daysDiff <= 1 {
				currentStreak++
				lastStudyDate = currentDate
			} else {
				break
			}
		}

		// Check if the last study was today or yesterday
		now := time.Now()
		daysSinceLastStudy := now.Sub(lastStudyDate).Hours() / 24
		if daysSinceLastStudy > 1 {
			currentStreak = 0
		}
	}

	studyStreakDays = currentStreak

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"success_rate":         successRate,
			"total_study_sessions": totalStudySessions,
			"total_active_groups":  totalActiveGroups,
			"study_streak_days":    studyStreakDays,
		},
	})
}

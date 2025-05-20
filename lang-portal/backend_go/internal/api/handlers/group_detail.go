package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/models"
)

// GroupDetailHandler handles group-specific detail endpoints
type GroupDetailHandler struct {
	db *gorm.DB
}

// NewGroupDetailHandler creates a new group detail handler
func NewGroupDetailHandler(db *gorm.DB) *GroupDetailHandler {
	return &GroupDetailHandler{db: db}
}

// GetGroupWords returns all words in a group
func (h *GroupDetailHandler) GetGroupWords(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var words []models.Word
	var total int64

	// Get words through the word_groups join table
	query := h.db.Model(&models.Word{}).
		Joins("JOIN word_groups ON word_groups.word_id = words.id").
		Where("word_groups.group_id = ?", groupID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count words"})
		return
	}

	// Get paginated words with their reviews
	if err := query.
		Offset(offset).
		Limit(params.PageSize).
		Preload("Reviews").
		Find(&words).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch words"})
		return
	}

	// Transform words for response
	items := make([]interface{}, len(words))
	for i, word := range words {
		// Calculate correct and wrong counts
		correctCount := 0
		wrongCount := 0
		for _, review := range word.Reviews {
			if review.Correct {
				correctCount++
			} else {
				wrongCount++
			}
		}

		items[i] = gin.H{
			"id":            word.ID,
			"japanese":      word.Japanese,
			"romaji":        word.Romaji,
			"english":       word.English,
			"correct_count": correctCount,
			"wrong_count":   wrongCount,
		}
	}

	c.JSON(http.StatusOK, middleware.NewPaginatedResponse(items, int(total), params))
}

// GetGroupStudySessions returns all study sessions for a group
func (h *GroupDetailHandler) GetGroupStudySessions(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var sessions []models.StudySession
	var total int64

	// Get sessions that include words from this group
	query := h.db.Model(&models.StudySession{}).
		Joins("JOIN word_review_items ON word_review_items.study_session_id = study_sessions.id").
		Joins("JOIN word_groups ON word_groups.word_id = word_review_items.word_id").
		Where("word_groups.group_id = ?", groupID).
		Group("study_sessions.id") // Ensure unique sessions

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count sessions"})
		return
	}

	// Get paginated sessions with related data
	if err := query.
		Offset(offset).
		Limit(params.PageSize).
		Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}

	// Transform sessions for response
	items := make([]interface{}, len(sessions))
	for i, session := range sessions {
		// Calculate success rate for this group's words
		var wordCount int64
		var correctCount int64
		h.db.Model(&models.WordReview{}).
			Joins("JOIN word_groups ON word_groups.word_id = word_review_items.word_id").
			Where("word_review_items.study_session_id = ? AND word_groups.group_id = ?", session.ID, groupID).
			Count(&wordCount)

		h.db.Model(&models.WordReview{}).
			Joins("JOIN word_groups ON word_groups.word_id = word_review_items.word_id").
			Where("word_review_items.study_session_id = ? AND word_groups.group_id = ? AND word_review_items.correct = ?",
				session.ID, groupID, true).
			Count(&correctCount)

		successRate := 0.0
		if wordCount > 0 {
			successRate = float64(correctCount) / float64(wordCount) * 100
		}

		items[i] = gin.H{
			"id":                 session.ID,
			"activity_name":      session.Activity.Name,
			"group_name":         session.Group.Name,
			"start_time":         session.CreatedAt,
			"end_time":           session.CreatedAt,
			"review_items_count": len(session.Reviews),
			"success_rate":       successRate,
		}
	}

	c.JSON(http.StatusOK, middleware.NewPaginatedResponse(items, int(total), params))
}

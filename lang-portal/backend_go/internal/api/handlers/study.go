package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/models"
)

// StudyHandler handles study-related requests
type StudyHandler struct {
	db *gorm.DB
}

// NewStudyHandler creates a new study handler
func NewStudyHandler(db *gorm.DB) *StudyHandler {
	return &StudyHandler{db: db}
}

// GetStudyActivities returns all study activities
func (h *StudyHandler) GetStudyActivities(c *gin.Context) {
	var activities []models.StudyActivity
	if err := h.db.Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study activities"})
		return
	}

	// Transform activities to match spec format
	items := make([]gin.H, len(activities))
	for i, activity := range activities {
		items[i] = gin.H{
			"id":            activity.ID,
			"name":          activity.Name,
			"thumbnail_url": activity.ThumbnailURL,
			"description":   activity.Description,
		}
	}

	c.JSON(http.StatusOK, gin.H{"study_activities": items})
}

// GetStudyActivity returns a specific study activity
func (h *StudyHandler) GetStudyActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	var activity models.StudyActivity
	if err := h.db.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Study activity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study activity"})
		return
	}

	// Get available groups for this activity
	var groups []models.Group
	if err := h.db.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch available groups"})
		return
	}

	// Transform groups to match spec format
	availableGroups := make([]gin.H, len(groups))
	for i, group := range groups {
		availableGroups[i] = gin.H{
			"id":   group.ID,
			"name": group.Name,
		}
	}

	// Return activity with available groups
	c.JSON(http.StatusOK, gin.H{
		"id":               activity.ID,
		"name":             activity.Name,
		"thumbnail_url":    activity.ThumbnailURL,
		"description":      activity.Description,
		"available_groups": availableGroups,
	})
}

// CreateStudySession creates a new study session
func (h *StudyHandler) CreateStudySession(c *gin.Context) {
	var session models.StudySession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify that the group and activity exist
	var group models.Group
	if err := h.db.First(&group, session.GroupID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group not found"})
		return
	}

	var activity models.StudyActivity
	if err := h.db.First(&activity, session.StudyActivityID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Study activity not found"})
		return
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create study session"})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetStudySessions returns all study sessions
func (h *StudyHandler) GetStudySessions(c *gin.Context) {
	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var sessions []models.StudySession
	var total int64

	// Get total count
	if err := h.db.Model(&models.StudySession{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count study sessions"})
		return
	}

	// Get paginated sessions with reviews
	if err := h.db.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Offset(offset).
		Limit(params.PageSize).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study sessions"})
		return
	}

	// Transform sessions for response
	items := make([]interface{}, len(sessions))
	for i, session := range sessions {
		// Calculate success rate
		var correctCount int
		for _, review := range session.Reviews {
			if review.Correct {
				correctCount++
			}
		}
		successRate := 0
		if len(session.Reviews) > 0 {
			successRate = (correctCount * 100) / len(session.Reviews)
		}

		items[i] = gin.H{
			"id":                 session.ID,
			"activity_name":      session.Activity.Name,
			"group_name":         session.Group.Name,
			"start_time":         session.CreatedAt,
			"end_time":           session.CreatedAt, // Using CreatedAt as end_time for now
			"review_items_count": len(session.Reviews),
			"success_rate":       successRate,
		}
	}

	c.JSON(http.StatusOK, middleware.NewPaginatedResponse(items, int(total), params))
}

// GetStudySession returns a specific study session
func (h *StudyHandler) GetStudySession(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var session models.StudySession
	if err := h.db.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		First(&session, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Study session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study session"})
		return
	}

	// Calculate success rate
	var correctCount int
	for _, review := range session.Reviews {
		if review.Correct {
			correctCount++
		}
	}
	successRate := 0
	if len(session.Reviews) > 0 {
		successRate = (correctCount * 100) / len(session.Reviews)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                 session.ID,
		"activity_name":      session.Activity.Name,
		"group_name":         session.Group.Name,
		"start_time":         session.CreatedAt,
		"end_time":           session.CreatedAt, // Using CreatedAt as end_time for now
		"review_items_count": len(session.Reviews),
		"success_rate":       successRate,
	})
}

// AddWordReview adds a word review to a study session
func (h *StudyHandler) AddWordReview(c *gin.Context) {
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	wordID, err := strconv.ParseUint(c.Param("word_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	// Parse request body
	var requestBody struct {
		Correct bool `json:"correct" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify that the session and word exist
	var session models.StudySession
	if err := h.db.First(&session, sessionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Study session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study session"})
		return
	}

	var word models.Word
	if err := h.db.First(&word, wordID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch word"})
		return
	}

	// Create the review
	review := models.WordReview{
		StudySessionID: uint(sessionID),
		WordID:         uint(wordID),
		Correct:        requestBody.Correct,
	}

	if err := h.db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create word review"})
		return
	}

	// Return response in spec format
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"word_id":          wordID,
		"study_session_id": sessionID,
		"correct":          requestBody.Correct,
		"created_at":       review.CreatedAt,
	})
}

// GetGroupStudySessions returns all study sessions for a specific group
func (h *StudyHandler) GetGroupStudySessions(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("group_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var sessions []models.StudySession
	if err := h.db.Where("group_id = ?", groupID).
		Preload("Activity").
		Preload("Reviews").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study sessions"})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

// GetStudyActivitySessions returns study sessions for a specific activity
func (h *StudyHandler) GetStudyActivitySessions(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var sessions []models.StudySession
	var total int64

	// Get total count
	if err := h.db.Model(&models.StudySession{}).Where("study_activity_id = ?", activityID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count study sessions"})
		return
	}

	// Get paginated sessions
	if err := h.db.Where("study_activity_id = ?", activityID).
		Preload("Activity").
		Preload("Group").
		Offset(offset).
		Limit(params.PageSize).
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch study sessions"})
		return
	}

	// Transform sessions for response
	items := make([]interface{}, len(sessions))
	for i, session := range sessions {
		items[i] = gin.H{
			"id":                 session.ID,
			"activity_name":      session.Activity.Name,
			"group_name":         session.Group.Name,
			"start_time":         session.CreatedAt,
			"review_items_count": len(session.Reviews),
		}
	}

	c.JSON(http.StatusOK, middleware.NewPaginatedResponse(items, int(total), params))
}

// GetStudySessionWords returns words reviewed in a study session
func (h *StudyHandler) GetStudySessionWords(c *gin.Context) {
	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var reviews []models.WordReview
	var total int64

	// Get total count
	if err := h.db.Model(&models.WordReview{}).Where("study_session_id = ?", sessionID).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count reviews"})
		return
	}

	// Get paginated reviews with words
	if err := h.db.Where("study_session_id = ?", sessionID).
		Preload("Word").
		Offset(offset).
		Limit(params.PageSize).
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	// Transform reviews for response
	items := make([]interface{}, len(reviews))
	for i, review := range reviews {
		items[i] = gin.H{
			"id":          review.Word.ID,
			"japanese":    review.Word.Japanese,
			"romaji":      review.Word.Romaji,
			"english":     review.Word.English,
			"correct":     review.Correct,
			"reviewed_at": review.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, middleware.NewPaginatedResponse(items, int(total), params))
}

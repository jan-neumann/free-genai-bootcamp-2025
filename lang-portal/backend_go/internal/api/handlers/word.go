package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/models"
)

// WordHandler handles word-related requests
type WordHandler struct {
	db *gorm.DB
}

// NewWordHandler creates a new word handler
func NewWordHandler(db *gorm.DB) *WordHandler {
	return &WordHandler{db: db}
}

// GetWords returns a paginated list of words
func (h *WordHandler) GetWords(c *gin.Context) {
	params := middleware.GetPaginationParams(c)
	offset := (params.Page - 1) * params.PageSize

	var words []models.Word
	var total int64

	// Get total count
	if err := h.db.Model(&models.Word{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count words"})
		return
	}

	// Get paginated words with their reviews
	if err := h.db.Offset(offset).
		Limit(params.PageSize).
		Order("japanese ASC").
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

// GetWord returns a single word by ID
func (h *WordHandler) GetWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	var word models.Word
	if err := h.db.Preload("Reviews").
		Preload("Groups").
		First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch word"})
		return
	}

	// Calculate study statistics
	correctCount := 0
	wrongCount := 0
	for _, review := range word.Reviews {
		if review.Correct {
			correctCount++
		} else {
			wrongCount++
		}
	}

	// Transform groups for response
	groups := make([]gin.H, len(word.Groups))
	for i, group := range word.Groups {
		groups[i] = gin.H{
			"id":   group.ID,
			"name": group.Name,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"word": gin.H{
			"id":       word.ID,
			"japanese": word.Japanese,
			"romaji":   word.Romaji,
			"english":  word.English,
			"study_stats": gin.H{
				"correct_count": correctCount,
				"wrong_count":   wrongCount,
			},
			"groups": groups,
		},
	})
}

// CreateWord creates a new word
func (h *WordHandler) CreateWord(c *gin.Context) {
	var word models.Word
	if err := c.ShouldBindJSON(&word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&word).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create word"})
		return
	}

	c.JSON(http.StatusCreated, word)
}

// UpdateWord updates an existing word
func (h *WordHandler) UpdateWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	var word models.Word
	if err := h.db.First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch word"})
		return
	}

	if err := c.ShouldBindJSON(&word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&word).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
		return
	}

	c.JSON(http.StatusOK, word)
}

// DeleteWord deletes a word
func (h *WordHandler) DeleteWord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
		return
	}

	if err := h.db.Delete(&models.Word{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete word"})
		return
	}

	c.Status(http.StatusNoContent)
}

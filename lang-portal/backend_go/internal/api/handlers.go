package api

import (
	"net/http"
	"strconv"

	"lang-portal/backend_go/internal/api/middleware"
	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/service"

	"github.com/gin-gonic/gin"
)

// Dashboard Handlers

func GetLastStudySession(s *service.DashboardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := s.GetLastStudySession()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, session)
	}
}

func GetStudyProgress(s *service.DashboardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		progress, err := s.GetStudyProgress()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, progress)
	}
}

func GetQuickStats(s *service.DashboardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := s.GetQuickStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

// Word Handlers

func CreateWord(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var word models.Word
		if err := c.ShouldBindJSON(&word); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.CreateWord(&word); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, word)
	}
}

func GetWord(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		word, err := s.GetWord(uint(id))
		if err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, word)
	}
}

func ListWords(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.ListWords(serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func UpdateWord(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		var word models.Word
		if err := c.ShouldBindJSON(&word); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.UpdateWord(uint(id), &word); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func DeleteWord(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		if err := s.DeleteWord(uint(id)); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func GetWordsByGroup(s *service.WordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.GetWordsByGroup(uint(groupID), serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

// Group Handlers

func CreateGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var group models.Group
		if err := c.ShouldBindJSON(&group); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.CreateGroup(&group); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeInvalidInput {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, group)
	}
}

func GetGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		group, err := s.GetGroup(uint(id))
		if err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, group)
	}
}

func ListGroups(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.ListGroups(serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func UpdateGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		var group models.Group
		if err := c.ShouldBindJSON(&group); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.UpdateGroup(uint(id), &group); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
				return
			}
			if err.(*service.ServiceError).Code == service.ErrCodeInvalidInput {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func DeleteGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		if err := s.DeleteGroup(uint(id)); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func AddWordToGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		wordID, err := strconv.ParseUint(c.Param("word_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		if err := s.AddWordToGroup(uint(groupID), uint(wordID)); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group or word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func RemoveWordFromGroup(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		wordID, err := strconv.ParseUint(c.Param("word_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		if err := s.RemoveWordFromGroup(uint(groupID), uint(wordID)); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group or word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func GetGroupStudyStats(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		totalSessions, totalReviews, correctReviews, err := s.GetGroupStudyStats(uint(id))
		if err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_sessions":  totalSessions,
			"total_reviews":   totalReviews,
			"correct_reviews": correctReviews,
			"success_rate":    calculateSuccessRate(int64(totalReviews), int64(correctReviews)),
		})
	}
}

func GetGroupsByWord(s *service.GroupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		wordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID"})
			return
		}

		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.GetGroupsByWord(uint(wordID), serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

// Study Handlers

func CreateStudyActivity(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var activity models.StudyActivity
		if err := c.ShouldBindJSON(&activity); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.CreateStudyActivity(&activity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, activity)
	}
}

func GetStudyActivity(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
			return
		}

		activity, err := s.GetStudyActivity(uint(id))
		if err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Study activity not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, activity)
	}
}

func ListStudyActivities(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.ListStudyActivities(serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func CreateStudySession(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var session models.StudySession
		if err := c.ShouldBindJSON(&session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.CreateStudySession(&session); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Group or activity not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, session)
	}
}

func GetStudySession(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		session, err := s.GetStudySession(uint(id))
		if err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Study session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

func ListStudySessions(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginParams := middleware.GetPaginationParams(c)

		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.ListStudySessions(serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list study sessions: " + err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func GetStudySessionsByGroup(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("group_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}

		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.GetStudySessionsByGroup(uint(groupID), serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func GetStudySessionsByActivity(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activityID, err := strconv.ParseUint(c.Param("activity_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
			return
		}

		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.GetStudySessionsByActivity(uint(activityID), serviceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func AddWordReview(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		var review models.WordReview
		if err := c.ShouldBindJSON(&review); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := s.AddWordReview(uint(sessionID), &review); err != nil {
			if err.(*service.ServiceError).Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Session or word not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, review)
	}
}

func GetWordReviewsBySession(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		ginParams := middleware.GetPaginationParams(c)
		serviceParams := service.PaginationParams{
			Page:     ginParams.Page,
			PageSize: ginParams.PageSize,
		}

		servicePaginatedResult, err := s.GetWordReviewsBySession(uint(sessionID), serviceParams)
		if err != nil {
			if serr, ok := err.(*service.ServiceError); ok && serr.Code == service.ErrCodeNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Study session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		interfaceItems := make([]interface{}, len(servicePaginatedResult.Items))
		for i, item := range servicePaginatedResult.Items {
			interfaceItems[i] = item
		}

		response := middleware.NewPaginatedResponse(interfaceItems, int(servicePaginatedResult.TotalItems), ginParams)
		c.JSON(http.StatusOK, response)
	}
}

func GetStudyStats(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		totalSessions, totalReviews, correctReviews, err := s.GetStudyStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total_sessions":  totalSessions,
			"total_reviews":   totalReviews,
			"correct_reviews": correctReviews,
			"success_rate":    calculateSuccessRate(totalReviews, correctReviews),
		})
	}
}

func GetStudyStreak(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		streak, err := s.GetStudyStreak()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"streak_days": streak})
	}
}

func GetActiveGroups(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := s.GetActiveGroups()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"active_groups": count})
	}
}

func ResetStudyHistory(s *service.StudyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := s.ResetStudyHistory(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// Helper functions

func calculateSuccessRate(total, correct int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(correct) / float64(total) * 100
}

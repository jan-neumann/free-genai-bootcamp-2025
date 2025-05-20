package service

import (
	"lang-portal/backend_go/internal/repository"
	"time"
)

// DashboardService handles dashboard-related business logic
type DashboardService struct {
	*BaseService
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(base *BaseService) *DashboardService {
	return &DashboardService{BaseService: base}
}

// LastStudySession represents the last study session information
type LastStudySession struct {
	ID              uint      `json:"id"`
	GroupID         uint      `json:"group_id"`
	CreatedAt       time.Time `json:"created_at"`
	StudyActivityID uint      `json:"study_activity_id"`
	GroupName       string    `json:"group_name"`
	ActivityName    string    `json:"activity_name"`
}

// GetLastStudySession returns information about the most recent study session
func (s *DashboardService) GetLastStudySession() (*LastStudySession, error) {
	session, err := s.studyRepo.GetLastStudySession()
	if err != nil {
		if err == repository.ErrNotFound {
			return &LastStudySession{}, nil
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch last study session", err)
	}

	return &LastStudySession{
		ID:              session.ID,
		GroupID:         session.GroupID,
		CreatedAt:       session.CreatedAt,
		StudyActivityID: session.StudyActivityID,
		GroupName:       session.Group.Name,
		ActivityName:    session.Activity.Name,
	}, nil
}

// StudyProgress represents study progress statistics
type StudyProgress struct {
	TotalWordsStudied   int64 `json:"total_words_studied"`
	TotalAvailableWords int64 `json:"total_available_words"`
}

// GetStudyProgress returns study progress statistics
func (s *DashboardService) GetStudyProgress() (*StudyProgress, error) {
	// Get total available words
	totalWords, err := s.wordRepo.GetTotalWordCount()
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get total word count", err)
	}

	// Get total studied words
	studiedWords, err := s.wordRepo.GetStudiedWordCount()
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get studied word count", err)
	}

	return &StudyProgress{
		TotalWordsStudied:   studiedWords,
		TotalAvailableWords: totalWords,
	}, nil
}

// QuickStats represents quick overview statistics
type QuickStats struct {
	SuccessRate        int   `json:"success_rate"`
	TotalStudySessions int64 `json:"total_study_sessions"`
	TotalActiveGroups  int64 `json:"total_active_groups"`
	StudyStreakDays    int   `json:"study_streak_days"`
}

// GetQuickStats returns quick overview statistics
func (s *DashboardService) GetQuickStats() (*QuickStats, error) {
	// Get total study sessions
	totalSessions, totalReviews, correctReviews, err := s.studyRepo.GetStudyStats()
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get study statistics", err)
	}

	// Get total active groups
	activeGroups, err := s.studyRepo.GetActiveGroups()
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get active groups count", err)
	}

	// Calculate success rate
	successRate := 0
	if totalReviews > 0 {
		successRate = int((float64(correctReviews) / float64(totalReviews)) * 100)
	}

	// Get study streak
	streak, err := s.studyRepo.GetStudyStreak()
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to calculate study streak", err)
	}

	return &QuickStats{
		SuccessRate:        successRate,
		TotalStudySessions: totalSessions,
		TotalActiveGroups:  activeGroups,
		StudyStreakDays:    streak,
	}, nil
}

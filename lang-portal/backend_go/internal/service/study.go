package service

import (
	"time"

	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/repository"
)

// StudyService handles study-related business logic
type StudyService struct {
	*BaseService
}

// NewStudyService creates a new study service
func NewStudyService(base *BaseService) *StudyService {
	return &StudyService{BaseService: base}
}

// StudyActivity represents a study activity
type StudyActivity struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// StudySession represents a study session
type StudySession struct {
	ID              uint      `json:"id"`
	GroupID         uint      `json:"group_id"`
	StudyActivityID uint      `json:"study_activity_id"`
	GroupName       string    `json:"group_name"`
	ActivityName    string    `json:"activity_name"`
	CreatedAt       time.Time `json:"created_at"`
	WordCount       int       `json:"word_count"`
}

// WordReview represents a word review
type WordReview struct {
	ID        uint      `json:"id"`
	WordID    uint      `json:"word_id"`
	Japanese  string    `json:"japanese"`
	Romaji    string    `json:"romaji"`
	English   string    `json:"english"`
	Correct   bool      `json:"correct"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStudyActivity creates a new study activity
func (s *StudyService) CreateStudyActivity(activity *models.StudyActivity) error {
	if err := s.studyRepo.CreateStudyActivity(activity); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to create study activity", err)
	}
	return nil
}

// GetStudyActivity retrieves a study activity by ID
func (s *StudyService) GetStudyActivity(id uint) (*StudyActivity, error) {
	activity, err := s.studyRepo.GetStudyActivityByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Study activity not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch study activity", err)
	}

	return &StudyActivity{
		ID:   activity.ID,
		Name: activity.Name,
	}, nil
}

// ListStudyActivities retrieves a paginated list of study activities
func (s *StudyService) ListStudyActivities(params PaginationParams) (*PaginatedResult[StudyActivity], error) {
	result, err := s.studyRepo.ListStudyActivities(repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to list study activities", err)
	}

	// Transform activities
	activities := make([]StudyActivity, len(result.Items))
	for i, a := range result.Items {
		activities[i] = StudyActivity{
			ID:   a.ID,
			Name: a.Name,
		}
	}

	return NewPaginatedResult(activities, result.TotalItems, params.Page, params.PageSize), nil
}

// CreateStudySession creates a new study session
func (s *StudyService) CreateStudySession(session *models.StudySession) error {
	// Verify group exists
	if _, err := s.groupRepo.GetByID(session.GroupID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch group", err)
	}

	// Verify activity exists
	if _, err := s.studyRepo.GetStudyActivityByID(session.StudyActivityID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Study activity not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch study activity", err)
	}

	if err := s.studyRepo.CreateStudySession(session); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to create study session", err)
	}
	return nil
}

// GetStudySession retrieves a study session by ID
func (s *StudyService) GetStudySession(id uint) (*StudySession, error) {
	session, err := s.studyRepo.GetStudySessionByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Study session not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch study session", err)
	}

	return &StudySession{
		ID:              session.ID,
		GroupID:         session.GroupID,
		StudyActivityID: session.StudyActivityID,
		GroupName:       session.Group.Name,
		ActivityName:    session.Activity.Name,
		CreatedAt:       session.CreatedAt,
		WordCount:       len(session.Reviews),
	}, nil
}

// ListStudySessions retrieves a paginated list of study sessions
func (s *StudyService) ListStudySessions(params PaginationParams) (*PaginatedResult[StudySession], error) {
	result, err := s.studyRepo.ListStudySessions(repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to list study sessions", err)
	}

	// Transform sessions
	sessions := make([]StudySession, len(result.Items))
	for i, s := range result.Items {
		sessions[i] = StudySession{
			ID:              s.ID,
			GroupID:         s.GroupID,
			StudyActivityID: s.StudyActivityID,
			GroupName:       s.Group.Name,
			ActivityName:    s.Activity.Name,
			CreatedAt:       s.CreatedAt,
			WordCount:       len(s.Reviews),
		}
	}

	return NewPaginatedResult(sessions, result.TotalItems, params.Page, params.PageSize), nil
}

// GetStudySessionsByGroup retrieves study sessions for a specific group
func (s *StudyService) GetStudySessionsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[StudySession], error) {
	result, err := s.studyRepo.GetStudySessionsByGroup(groupID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get group study sessions", err)
	}

	// Transform sessions
	sessions := make([]StudySession, len(result.Items))
	for i, s := range result.Items {
		sessions[i] = StudySession{
			ID:              s.ID,
			GroupID:         s.GroupID,
			StudyActivityID: s.StudyActivityID,
			GroupName:       s.Group.Name,
			ActivityName:    s.Activity.Name,
			CreatedAt:       s.CreatedAt,
			WordCount:       len(s.Reviews),
		}
	}

	return NewPaginatedResult(sessions, result.TotalItems, params.Page, params.PageSize), nil
}

// GetStudySessionsByActivity retrieves study sessions for a specific activity
func (s *StudyService) GetStudySessionsByActivity(activityID uint, params PaginationParams) (*PaginatedResult[StudySession], error) {
	result, err := s.studyRepo.GetStudySessionsByActivity(activityID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get activity study sessions", err)
	}

	// Transform sessions
	sessions := make([]StudySession, len(result.Items))
	for i, s := range result.Items {
		sessions[i] = StudySession{
			ID:              s.ID,
			GroupID:         s.GroupID,
			StudyActivityID: s.StudyActivityID,
			GroupName:       s.Group.Name,
			ActivityName:    s.Activity.Name,
			CreatedAt:       s.CreatedAt,
			WordCount:       len(s.Reviews),
		}
	}

	return NewPaginatedResult(sessions, result.TotalItems, params.Page, params.PageSize), nil
}

// AddWordReview adds a word review to a study session
func (s *StudyService) AddWordReview(sessionID uint, review *models.WordReview) error {
	// Verify session exists
	if _, err := s.studyRepo.GetStudySessionByID(sessionID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Study session not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch study session", err)
	}

	// Verify word exists
	if _, err := s.wordRepo.GetByID(review.WordID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch word", err)
	}

	// Set the session ID
	review.StudySessionID = sessionID

	if err := s.studyRepo.AddWordReview(review); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to add word review", err)
	}
	return nil
}

// GetWordReviewsBySession retrieves word reviews for a specific study session
func (s *StudyService) GetWordReviewsBySession(sessionID uint, params PaginationParams) (*PaginatedResult[WordReview], error) {
	// Verify session exists
	if _, err := s.studyRepo.GetStudySessionByID(sessionID); err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Study session not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch study session", err)
	}

	result, err := s.studyRepo.GetWordReviewsBySession(sessionID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get word reviews", err)
	}

	// Transform reviews
	reviews := make([]WordReview, len(result.Items))
	for i, r := range result.Items {
		reviews[i] = WordReview{
			ID:        r.ID,
			WordID:    r.WordID,
			Japanese:  r.Word.Japanese,
			Romaji:    r.Word.Romaji,
			English:   r.Word.English,
			Correct:   r.Correct,
			CreatedAt: r.CreatedAt,
		}
	}

	return NewPaginatedResult(reviews, result.TotalItems, params.Page, params.PageSize), nil
}

// GetLastStudySession retrieves the most recent study session
func (s *StudyService) GetLastStudySession() (*StudySession, error) {
	session, err := s.studyRepo.GetLastStudySession()
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, nil // Return nil instead of error for no sessions
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to get last study session", err)
	}

	return &StudySession{
		ID:              session.ID,
		GroupID:         session.GroupID,
		StudyActivityID: session.StudyActivityID,
		GroupName:       session.Group.Name,
		ActivityName:    session.Activity.Name,
		CreatedAt:       session.CreatedAt,
		WordCount:       len(session.Reviews),
	}, nil
}

// GetStudyStats retrieves overall study statistics
func (s *StudyService) GetStudyStats() (totalSessions, totalReviews, correctReviews int64, err error) {
	totalSessions, totalReviews, correctReviews, err = s.studyRepo.GetStudyStats()
	if err != nil {
		return 0, 0, 0, NewServiceError(ErrCodeInternal, "Failed to get study statistics", err)
	}
	return
}

// GetStudyStreak retrieves the current study streak in days
func (s *StudyService) GetStudyStreak() (int, error) {
	streak, err := s.studyRepo.GetStudyStreak()
	if err != nil {
		return 0, NewServiceError(ErrCodeInternal, "Failed to get study streak", err)
	}
	return streak, nil
}

// GetActiveGroups retrieves the number of groups that have been studied
func (s *StudyService) GetActiveGroups() (int64, error) {
	count, err := s.studyRepo.GetActiveGroups()
	if err != nil {
		return 0, NewServiceError(ErrCodeInternal, "Failed to get active groups", err)
	}
	return count, nil
}

// ResetStudyHistory resets all study-related data
func (s *StudyService) ResetStudyHistory() error {
	if err := s.studyRepo.ResetStudyHistory(); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to reset study history", err)
	}
	return nil
}

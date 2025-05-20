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

// StudySessionInfo is the DTO for study session details and list items.
// It matches the structure specified for GET /api/study/sessions and GET /api/study/sessions/:id.
type StudySessionInfo struct {
	ID               uint      `json:"id"`
	ActivityName     string    `json:"activity_name"`
	GroupName        string    `json:"group_name"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"` // Placeholder: using CreatedAt from model
	ReviewItemsCount int       `json:"review_items_count"`
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

	// Reload the session to populate associations for the response for POST
	loadedSession, err := s.studyRepo.GetStudySessionByID(session.ID)
	if err != nil {
		// Log this error or handle as per requirements.
		// For POST, we need the full model, so if this fails, it's an issue.
		// However, the primary record was created. Consider returning partial or error.
		// For now, this logic is for POST. The handler will return session (models.StudySession).
	} else {
		session.Group = loadedSession.Group
		session.Activity = loadedSession.Activity
		session.Reviews = loadedSession.Reviews // Ensure reviews are also copied if needed by model users
	}

	return nil
}

// GetStudySession retrieves a study session by ID and transforms it to StudySessionInfo DTO.
func (s *StudyService) GetStudySession(id uint) (*StudySessionInfo, error) {
	modelSession, err := s.studyRepo.GetStudySessionByID(id) // This fetches models.StudySession
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Study session not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch study session", err)
	}

	return &StudySessionInfo{
		ID:               modelSession.ID,
		ActivityName:     modelSession.Activity.Name,
		GroupName:        modelSession.Group.Name,
		StartTime:        modelSession.CreatedAt,    // Map CreatedAt to StartTime
		EndTime:          modelSession.CreatedAt,    // Placeholder: Map CreatedAt to EndTime
		ReviewItemsCount: len(modelSession.Reviews), // Map Reviews length to ReviewItemsCount
	}, nil
}

// ListStudySessions retrieves a paginated list of study sessions
func (s *StudyService) ListStudySessions(params PaginationParams) (*PaginatedResult[StudySessionInfo], error) {
	result, err := s.studyRepo.ListStudySessions(repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to list study sessions", err)
	}

	// Transform sessions
	infos := make([]StudySessionInfo, len(result.Items))
	for i, modelSession := range result.Items {
		infos[i] = StudySessionInfo{
			ID:               modelSession.ID,
			ActivityName:     modelSession.Activity.Name,
			GroupName:        modelSession.Group.Name,
			StartTime:        modelSession.CreatedAt,
			EndTime:          modelSession.CreatedAt, // Placeholder
			ReviewItemsCount: len(modelSession.Reviews),
		}
	}

	return NewPaginatedResult(infos, result.TotalItems, params.Page, params.PageSize), nil
}

// GetStudySessionsByGroup retrieves study sessions for a specific group
func (s *StudyService) GetStudySessionsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[StudySessionInfo], error) {
	result, err := s.studyRepo.GetStudySessionsByGroup(groupID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get group study sessions", err)
	}

	// Transform sessions
	infos := make([]StudySessionInfo, len(result.Items))
	for i, modelSession := range result.Items {
		infos[i] = StudySessionInfo{
			ID:               modelSession.ID,
			ActivityName:     modelSession.Activity.Name,
			GroupName:        modelSession.Group.Name,
			StartTime:        modelSession.CreatedAt,
			EndTime:          modelSession.CreatedAt, // Placeholder
			ReviewItemsCount: len(modelSession.Reviews),
		}
	}

	return NewPaginatedResult(infos, result.TotalItems, params.Page, params.PageSize), nil
}

// GetStudySessionsByActivity retrieves study sessions for a specific activity
func (s *StudyService) GetStudySessionsByActivity(activityID uint, params PaginationParams) (*PaginatedResult[StudySessionInfo], error) {
	result, err := s.studyRepo.GetStudySessionsByActivity(activityID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get activity study sessions", err)
	}

	// Transform sessions
	infos := make([]StudySessionInfo, len(result.Items))
	for i, modelSession := range result.Items {
		infos[i] = StudySessionInfo{
			ID:               modelSession.ID,
			ActivityName:     modelSession.Activity.Name,
			GroupName:        modelSession.Group.Name,
			StartTime:        modelSession.CreatedAt,
			EndTime:          modelSession.CreatedAt, // Placeholder
			ReviewItemsCount: len(modelSession.Reviews),
		}
	}

	return NewPaginatedResult(infos, result.TotalItems, params.Page, params.PageSize), nil
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
func (s *StudyService) GetLastStudySession() (*StudySessionInfo, error) {
	modelSession, err := s.studyRepo.GetLastStudySession()
	if err != nil {
		if err == repository.ErrNotFound {
			// For GetLastStudySession, returning nil (which becomes null in JSON) is fine if no session exists.
			return nil, nil
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to get last study session", err)
	}

	// If a session is found, transform it to StudySessionInfo
	return &StudySessionInfo{
		ID:               modelSession.ID,
		ActivityName:     modelSession.Activity.Name,
		GroupName:        modelSession.Group.Name,
		StartTime:        modelSession.CreatedAt,
		EndTime:          modelSession.CreatedAt, // Placeholder
		ReviewItemsCount: len(modelSession.Reviews),
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

package repository

import (
	"lang-portal/backend_go/internal/models"
	"time"

	"gorm.io/gorm"
)

// StudyRepository handles database operations for study-related entities
type StudyRepository struct {
	*BaseRepository
}

// NewStudyRepository creates a new study repository
func NewStudyRepository(db *gorm.DB) *StudyRepository {
	return &StudyRepository{BaseRepository: NewBaseRepository(db)}
}

// CreateStudyActivity creates a new study activity
func (r *StudyRepository) CreateStudyActivity(activity *models.StudyActivity) error {
	if err := activity.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Create(activity).Error
}

// GetStudyActivityByID retrieves a study activity by ID
func (r *StudyRepository) GetStudyActivityByID(id uint) (*models.StudyActivity, error) {
	var activity models.StudyActivity
	if err := r.db.Preload("Sessions").First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// ListStudyActivities retrieves a paginated list of study activities
func (r *StudyRepository) ListStudyActivities(params PaginationParams) (*PaginatedResult[models.StudyActivity], error) {
	var activities []models.StudyActivity
	var total int64

	query := r.db.Model(&models.StudyActivity{})
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Find(&activities).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.StudyActivity]{
		Items:      activities,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CreateStudySession creates a new study session
func (r *StudyRepository) CreateStudySession(session *models.StudySession) error {
	if err := session.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Create(session).Error
}

// GetStudySessionByID retrieves a study session by ID
func (r *StudyRepository) GetStudySessionByID(id uint) (*models.StudySession, error) {
	var session models.StudySession
	if err := r.db.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Preload("Reviews.Word").
		First(&session, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

// ListStudySessions retrieves a paginated list of study sessions
func (r *StudyRepository) ListStudySessions(params PaginationParams) (*PaginatedResult[models.StudySession], error) {
	var sessions []models.StudySession
	var total int64

	query := r.db.Model(&models.StudySession{})
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.StudySession]{
		Items:      sessions,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStudySessionsByGroup retrieves study sessions for a specific group
func (r *StudyRepository) GetStudySessionsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[models.StudySession], error) {
	var sessions []models.StudySession
	var total int64

	query := r.db.Model(&models.StudySession{}).Where("group_id = ?", groupID)
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.StudySession]{
		Items:      sessions,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetStudySessionsByActivity retrieves study sessions for a specific activity
func (r *StudyRepository) GetStudySessionsByActivity(activityID uint, params PaginationParams) (*PaginatedResult[models.StudySession], error) {
	var sessions []models.StudySession
	var total int64

	query := r.db.Model(&models.StudySession{}).Where("study_activity_id = ?", activityID)
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Activity").
		Preload("Group").
		Preload("Reviews").
		Order("created_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.StudySession]{
		Items:      sessions,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// AddWordReview adds a word review to a study session
func (r *StudyRepository) AddWordReview(review *models.WordReview) error {
	if err := review.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Create(review).Error
}

// GetWordReviewsBySession retrieves word reviews for a specific study session
func (r *StudyRepository) GetWordReviewsBySession(sessionID uint, params PaginationParams) (*PaginatedResult[models.WordReview], error) {
	var reviews []models.WordReview
	var total int64

	query := r.db.Model(&models.WordReview{}).Where("study_session_id = ?", sessionID)
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Word").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.WordReview]{
		Items:      reviews,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetLastStudySession retrieves the most recent study session
func (r *StudyRepository) GetLastStudySession() (*models.StudySession, error) {
	var session models.StudySession
	if err := r.db.Preload("Activity").
		Preload("Group").
		Order("created_at DESC").
		First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

// GetStudyStats retrieves overall study statistics
func (r *StudyRepository) GetStudyStats() (totalSessions, totalReviews, correctReviews int64, err error) {
	if err := r.db.Model(&models.StudySession{}).Count(&totalSessions).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := r.db.Model(&models.WordReview{}).Count(&totalReviews).Error; err != nil {
		return 0, 0, 0, err
	}

	if err := r.db.Model(&models.WordReview{}).Where("correct = ?", true).Count(&correctReviews).Error; err != nil {
		return 0, 0, 0, err
	}

	return
}

// GetStudyStreak retrieves the current study streak in days
func (r *StudyRepository) GetStudyStreak() (int, error) {
	var sessions []models.StudySession
	if err := r.db.Order("created_at DESC").Find(&sessions).Error; err != nil {
		return 0, err
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	lastStudyDate := sessions[0].CreatedAt
	currentStreak := 1

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

	return currentStreak, nil
}

// GetActiveGroups retrieves the number of groups that have been studied
func (r *StudyRepository) GetActiveGroups() (int64, error) {
	var count int64
	err := r.db.Model(&models.StudySession{}).
		Distinct("group_id").
		Count(&count).Error
	return count, err
}

// ResetStudyHistory resets all study-related data
func (r *StudyRepository) ResetStudyHistory() error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		// Delete word reviews
		if err := tx.Where("1=1").Delete(&models.WordReview{}).Error; err != nil {
			return err
		}
		// Delete study sessions
		if err := tx.Where("1=1").Delete(&models.StudySession{}).Error; err != nil {
			return err
		}
		return nil
	})
}

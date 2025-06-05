package repository

import "lang-portal/backend_go/internal/models"

// WordRepositoryInterface defines the interface for word repository operations.
type WordRepositoryInterface interface {
	Create(word *models.Word) error
	GetByID(id uint) (*models.Word, error)
	List(params PaginationParams) (*PaginatedResult[models.Word], error)
	Update(word *models.Word) error
	Delete(id uint) error
	GetStudyStats(wordID uint) (correctCount int64, wrongCount int64, err error)
	GetWordsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[models.Word], error)
	GetWordsByGroupRaw(groupID uint) ([]models.Word, error)
	GetTotalWordCount() (int64, error)
	GetStudiedWordCount() (int64, error)
	GetByJapanese(japanese string) (*models.Word, error)
}

// GroupRepositoryInterface defines the interface for group repository operations.
// Add methods as they are identified from GroupService usage.
type GroupRepositoryInterface interface {
	Create(group *models.Group) error
	GetByID(id uint) (*models.Group, error)
	GetByName(name string) (*models.Group, error)
	List(params PaginationParams) (*PaginatedResult[models.Group], error)
	Update(group *models.Group) error
	Delete(id uint) error
	AddWord(groupID, wordID uint) error
	RemoveWord(groupID, wordID uint) error
	GetStudyStats(id uint) (totalSessions, totalReviews, correctReviews int, err error) // Matches method in actual repo
	GetGroupsByWord(wordID uint, params PaginationParams) (*PaginatedResult[models.Group], error)
	GetTotalGroupCount() (int64, error)
	GetActiveGroupCount() (int64, error) // Added from GroupRepository
}

// StudyRepositoryInterface defines the interface for study repository operations.
// Add methods as they are identified from StudyService usage.
type StudyRepositoryInterface interface {
	CreateStudyActivity(activity *models.StudyActivity) error
	GetStudyActivityByID(id uint) (*models.StudyActivity, error)
	ListStudyActivities(params PaginationParams) (*PaginatedResult[models.StudyActivity], error)

	CreateStudySession(session *models.StudySession) error
	GetStudySessionByID(id uint) (*models.StudySession, error)
	ListStudySessions(params PaginationParams) (*PaginatedResult[models.StudySession], error)
	GetStudySessionsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[models.StudySession], error)
	GetStudySessionsByActivity(activityID uint, params PaginationParams) (*PaginatedResult[models.StudySession], error)

	AddWordReview(review *models.WordReview) error
	GetWordReviewsBySession(sessionID uint, params PaginationParams) (*PaginatedResult[models.WordReview], error)

	GetLastStudySession() (*models.StudySession, error)
	GetStudyStats() (totalSessions, totalReviews, correctReviews int64, err error)
	GetStudyStreak() (int, error)
	GetActiveGroups() (int64, error)
	ResetStudyHistory() error
}

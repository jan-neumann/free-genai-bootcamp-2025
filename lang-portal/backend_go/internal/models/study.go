package models

import (
	"time"
)

// StudyActivity represents a specific study activity type
type StudyActivity struct {
	ID           uint           `gorm:"primarykey" json:"id" validate:"required"`
	Name         string         `gorm:"not null;uniqueIndex" json:"name" validate:"required,min=1"`
	Description  string         `gorm:"not null" json:"description" validate:"required,min=1"`
	ThumbnailURL string         `gorm:"not null" json:"thumbnail_url" validate:"required,url"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Sessions     []StudySession `gorm:"foreignKey:StudyActivityID" json:"sessions,omitempty"`
}

// TableName specifies the table name for the StudyActivity model
func (StudyActivity) TableName() string {
	return "study_activities"
}

// Validate validates the StudyActivity model
func (a *StudyActivity) Validate() error {
	return validate.Struct(a)
}

// StudySession represents a study session
type StudySession struct {
	ID              uint          `gorm:"primarykey" json:"id" validate:"required"`
	GroupID         uint          `gorm:"not null;index" json:"group_id" validate:"required"`
	StudyActivityID uint          `gorm:"not null;index" json:"study_activity_id" validate:"required"`
	CreatedAt       time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Group           Group         `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Activity        StudyActivity `gorm:"foreignKey:StudyActivityID" json:"activity,omitempty"`
	Reviews         []WordReview  `gorm:"foreignKey:StudySessionID" json:"reviews,omitempty"`
}

// TableName specifies the table name for the StudySession model
func (StudySession) TableName() string {
	return "study_sessions"
}

// Validate validates the StudySession model
func (s *StudySession) Validate() error {
	return validate.Struct(s)
}

// GetStudyStats returns the study statistics for the session
func (s *StudySession) GetStudyStats() (totalReviews, correctReviews int) {
	totalReviews = len(s.Reviews)
	for _, review := range s.Reviews {
		if review.Correct {
			correctReviews++
		}
	}
	return
}

// GetSuccessRate returns the success rate for the session
func (s *StudySession) GetSuccessRate() float64 {
	totalReviews, correctReviews := s.GetStudyStats()
	if totalReviews == 0 {
		return 0
	}
	return float64(correctReviews) / float64(totalReviews) * 100
}

// WordReview represents a word review in a study session
type WordReview struct {
	ID             uint         `gorm:"primarykey" json:"id" validate:"required"`
	WordID         uint         `gorm:"not null;index" json:"word_id" validate:"required"`
	StudySessionID uint         `gorm:"not null;index" json:"study_session_id" validate:"required"`
	Correct        bool         `gorm:"not null" json:"correct" validate:"required"`
	CreatedAt      time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Word           Word         `gorm:"foreignKey:WordID" json:"word,omitempty"`
	StudySession   StudySession `gorm:"foreignKey:StudySessionID" json:"study_session,omitempty"`
}

// TableName specifies the table name for the WordReview model
func (WordReview) TableName() string {
	return "word_review_items"
}

// Validate validates the WordReview model
func (r *WordReview) Validate() error {
	return validate.Struct(r)
}

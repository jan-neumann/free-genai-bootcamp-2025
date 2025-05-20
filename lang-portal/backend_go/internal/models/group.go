package models

import (
	"time"
)

// Group represents a thematic group of words
type Group struct {
	ID        uint           `gorm:"primarykey" json:"id" validate:"required"`
	Name      string         `gorm:"not null;uniqueIndex" json:"name" validate:"required,min=1"`
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Words     []Word         `gorm:"many2many:word_groups;" json:"words,omitempty"`
	Sessions  []StudySession `gorm:"foreignKey:GroupID" json:"sessions,omitempty"`
}

// TableName specifies the table name for the Group model
func (Group) TableName() string {
	return "groups"
}

// Validate validates the Group model
func (g *Group) Validate() error {
	return validate.Struct(g)
}

// GetWordCount returns the number of words in the group
func (g *Group) GetWordCount() int {
	return len(g.Words)
}

// GetStudyStats returns the study statistics for the group
func (g *Group) GetStudyStats() (totalSessions, totalReviews, correctReviews int) {
	totalSessions = len(g.Sessions)
	for _, session := range g.Sessions {
		for _, review := range session.Reviews {
			totalReviews++
			if review.Correct {
				correctReviews++
			}
		}
	}
	return
}

// GetSuccessRate returns the success rate for the group
func (g *Group) GetSuccessRate() float64 {
	_, totalReviews, correctReviews := g.GetStudyStats()
	if totalReviews == 0 {
		return 0
	}
	return float64(correctReviews) / float64(totalReviews) * 100
}

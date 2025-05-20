package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
)

// StringSlice is a custom type for JSON array storage
type StringSlice []string

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	return json.Unmarshal(value.([]byte), s)
}

// Word represents a vocabulary word
type Word struct {
	ID        uint         `gorm:"primarykey" json:"id" validate:"required"`
	Japanese  string       `gorm:"not null;index" json:"japanese" validate:"required,min=1"`
	Romaji    string       `gorm:"not null" json:"romaji" validate:"required,min=1"`
	English   string       `gorm:"not null" json:"english" validate:"required,min=1"`
	Parts     StringSlice  `gorm:"type:json;not null" json:"parts" validate:"required,min=1"`
	CreatedAt time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	Groups    []Group      `gorm:"many2many:word_groups;" json:"groups,omitempty"`
	Reviews   []WordReview `gorm:"foreignKey:WordID" json:"reviews,omitempty"`
}

// TableName specifies the table name for the Word model
func (Word) TableName() string {
	return "words"
}

var validate = validator.New()

// Validate validates the Word model
func (w *Word) Validate() error {
	return validate.Struct(w)
}

// GetStudyStats returns the study statistics for the word
func (w *Word) GetStudyStats() (correctCount, wrongCount int) {
	for _, review := range w.Reviews {
		if review.Correct {
			correctCount++
		} else {
			wrongCount++
		}
	}
	return
}

// GetSuccessRate returns the success rate for the word
func (w *Word) GetSuccessRate() float64 {
	correctCount, wrongCount := w.GetStudyStats()
	total := correctCount + wrongCount
	if total == 0 {
		return 0
	}
	return float64(correctCount) / float64(total) * 100
}

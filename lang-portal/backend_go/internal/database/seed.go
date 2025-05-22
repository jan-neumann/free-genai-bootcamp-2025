package database

import (
	"lang-portal/backend_go/internal/models"

	"gorm.io/gorm"
)

// Seed populates the database with initial data
func Seed(db *gorm.DB) error {
	// Create study activities
	activities := []models.StudyActivity{
		{
			Name: "Reading Practice",
			Description: "Practice reading Japanese text",
			ThumbnailURL: "/images/reading.png",
		},
		{
			Name: "Writing Practice",
			Description: "Practice writing Japanese characters",
			ThumbnailURL: "/images/writing.png",
		},
		{
			Name: "Listening Practice",
			Description: "Practice listening to Japanese",
			ThumbnailURL: "/images/listening.png",
		},
		{
			Name: "Speaking Practice",
			Description: "Practice speaking Japanese",
			ThumbnailURL: "/images/speaking.png",
		},
	}
	if err := db.Create(&activities).Error; err != nil {
		return err
	}

	// Create groups
	groups := []models.Group{
		{Name: "Basic Greetings"},
		{Name: "Numbers"},
		{Name: "Days of the Week"},
		{Name: "Common Verbs"},
	}
	if err := db.Create(&groups).Error; err != nil {
		return err
	}

	// Create words
	words := []models.Word{
		{
			Japanese: "こんにちは",
			Romaji:   "konnichiwa",
			English:  "Hello",
			Groups:   []models.Group{groups[0]},
		},
		{
			Japanese: "さようなら",
			Romaji:   "sayounara",
			English:  "Goodbye",
			Groups:   []models.Group{groups[0]},
		},
		{
			Japanese: "一",
			Romaji:   "ichi",
			English:  "One",
			Groups:   []models.Group{groups[1]},
		},
		{
			Japanese: "二",
			Romaji:   "ni",
			English:  "Two",
			Groups:   []models.Group{groups[1]},
		},
		{
			Japanese: "月曜日",
			Romaji:   "getsuyoubi",
			English:  "Monday",
			Groups:   []models.Group{groups[2]},
		},
		{
			Japanese: "火曜日",
			Romaji:   "kayoubi",
			English:  "Tuesday",
			Groups:   []models.Group{groups[2]},
		},
		{
			Japanese: "食べる",
			Romaji:   "taberu",
			English:  "To eat",
			Groups:   []models.Group{groups[3]},
		},
		{
			Japanese: "飲む",
			Romaji:   "nomu",
			English:  "To drink",
			Groups:   []models.Group{groups[3]},
		},
	}
	if err := db.Create(&words).Error; err != nil {
		return err
	}

	return nil
}

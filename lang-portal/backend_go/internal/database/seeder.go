package database

import (
	"log"

	"gorm.io/gorm"

	"lang-portal/backend_go/internal/models"
)

// SeedData populates the database with initial data
func SeedData(db *gorm.DB) error {
	// Seed study activities
	activities := []models.StudyActivity{
		{
			Name:         "Flashcards",
			Description:  "Practice vocabulary with flashcards",
			ThumbnailURL: "/images/activities/flashcards.jpg",
		},
		{
			Name:         "Multiple Choice",
			Description:  "Test your knowledge with multiple choice questions",
			ThumbnailURL: "/images/activities/multiple-choice.jpg",
		},
		{
			Name:         "Word Matching",
			Description:  "Match words with their meanings",
			ThumbnailURL: "/images/activities/word-matching.jpg",
		},
		{
			Name:         "Sentence Completion",
			Description:  "Complete sentences with the correct words",
			ThumbnailURL: "/images/activities/sentence-completion.jpg",
		},
		{
			Name:         "Listening Practice",
			Description:  "Practice listening and understanding spoken words",
			ThumbnailURL: "/images/activities/listening.jpg",
		},
	}

	// Seed basic verb groups
	groups := []models.Group{
		{
			Name: "Basic Verbs - Present Tense",
		},
		{
			Name: "Basic Verbs - Past Tense",
		},
		{
			Name: "Basic Verbs - Future Tense",
		},
	}

	// Seed some basic words
	words := []models.Word{
		{
			Japanese: "食べる",
			Romaji:   "taberu",
			English:  "to eat",
			Parts:    []string{"verb", "ichidan", "present"},
		},
		{
			Japanese: "飲む",
			Romaji:   "nomu",
			English:  "to drink",
			Parts:    []string{"verb", "godan", "present"},
		},
		{
			Japanese: "行く",
			Romaji:   "iku",
			English:  "to go",
			Parts:    []string{"verb", "godan", "present"},
		},
		{
			Japanese: "来る",
			Romaji:   "kuru",
			English:  "to come",
			Parts:    []string{"verb", "irregular", "present"},
		},
		{
			Japanese: "する",
			Romaji:   "suru",
			English:  "to do",
			Parts:    []string{"verb", "irregular", "present"},
		},
	}

	// Begin transaction
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Seed activities
	for _, activity := range activities {
		if err := tx.FirstOrCreate(&activity, models.StudyActivity{Name: activity.Name}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Seed groups
	for _, group := range groups {
		if err := tx.FirstOrCreate(&group, models.Group{Name: group.Name}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Seed words and associate them with groups
	for _, word := range words {
		if err := tx.FirstOrCreate(&word, models.Word{Japanese: word.Japanese}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Associate words with the "Basic Verbs - Present Tense" group
		var presentGroup models.Group
		if err := tx.Where("name = ?", "Basic Verbs - Present Tense").First(&presentGroup).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Model(&presentGroup).Association("Words").Append(&word); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

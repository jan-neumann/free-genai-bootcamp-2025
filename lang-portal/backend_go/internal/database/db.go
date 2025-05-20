package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"lang-portal/backend_go/internal/models"
)

// InitDB initializes the database connection, runs migrations, and seeds initial data
func InitDB() (*gorm.DB, error) {
	// Open SQLite database
	db, err := gorm.Open(sqlite.Open("lang_portal.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Word{},
		&models.Group{},
		&models.StudyActivity{},
		&models.StudySession{},
		&models.WordReview{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database migrations completed successfully")

	// Seed initial data
	if err := SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
		// Don't return error here, as seeding is not critical for the application to run
	}

	return db, nil
}

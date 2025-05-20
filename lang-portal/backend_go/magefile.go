//go:build mage

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/magefile/mage/mg"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"lang-portal/backend_go/internal/models"
)

// Database tasks
type DB mg.Namespace

// Initialize creates a new SQLite database
func (DB) Initialize() error {
	fmt.Println("Initializing database...")
	dbPath := "words.db"

	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf("database already exists at %s", dbPath)
	}

	// Create empty database file
	file, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}
	file.Close()

	// Initialize database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		os.Remove(dbPath) // Clean up the file if we can't open it
		return fmt.Errorf("failed to open database: %v", err)
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
		os.Remove(dbPath) // Clean up the file if migration fails
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	fmt.Println("Database initialized successfully at", dbPath)
	return nil
}

// Migrate runs all database migrations
func (DB) Migrate() error {
	fmt.Println("Running database migrations...")
	dbPath := "words.db"

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database does not exist at %s, run 'mage db:initialize' first", dbPath)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Get all migration files
	migrationsDir := "db/migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}

	// Sort files by name (they should be prefixed with numbers)
	sort.Strings(files)

	// Create migrations table if it doesn't exist
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Run each migration in a transaction
	for _, file := range files {
		fileName := filepath.Base(file)
		fmt.Printf("Checking migration: %s\n", fileName)

		// Check if migration has already been applied
		var count int64
		if err := db.Model(&struct{}{}).Table("migrations").Where("name = ?", fileName).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}
		if count > 0 {
			fmt.Printf("Migration %s already applied, skipping\n", fileName)
			continue
		}

		// Read migration file
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", fileName, err)
		}

		// Begin transaction
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction: %v", tx.Error)
		}

		// Execute migration
		if err := tx.Exec(string(content)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %v", fileName, err)
		}

		// Record migration
		if err := tx.Exec("INSERT INTO migrations (name) VALUES (?)", fileName).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %v", fileName, err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit migration %s: %v", fileName, err)
		}

		fmt.Printf("Successfully applied migration: %s\n", fileName)
	}

	fmt.Println("All migrations completed successfully")
	return nil
}

// Seed imports seed data into the database
func (DB) Seed() error {
	fmt.Println("Seeding database...")
	dbPath := "words.db"

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("database does not exist at %s, run 'mage db:initialize' first", dbPath)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Reduced default GORM logging for seeds
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Get all seed files (JSON and SQL)
	seedsDir := "db/seeds"
	jsonFiles, err := filepath.Glob(filepath.Join(seedsDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to find JSON seed files: %v", err)
	}
	sqlFiles, err := filepath.Glob(filepath.Join(seedsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to find SQL seed files: %v", err)
	}

	files := append(jsonFiles, sqlFiles...)
	if len(files) == 0 {
		fmt.Println("No seed files found in", seedsDir) // Changed to info, not error if no seeds
		return nil
	}

	// Sort files by name to ensure consistent order
	sort.Strings(files)

	// Process each seed file
	for _, file := range files {
		fileName := filepath.Base(file)
		fmt.Printf("Processing seed file: %s\n", fileName)

		// Read seed file content
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read seed file %s: %v", fileName, err)
		}

		// Begin transaction for this file
		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("failed to begin transaction for %s: %v", fileName, tx.Error)
		}

		if strings.HasSuffix(fileName, ".json") {
			groupName := strings.TrimSuffix(fileName, ".json")
			// Parse JSON data
			var words []struct {
				Japanese string   `json:"japanese"`
				Romaji   string   `json:"romaji"`
				English  string   `json:"english"`
				Parts    []string `json:"parts"`
			}
			if err := json.Unmarshal(content, &words); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to parse seed file %s: %v", fileName, err)
			}

			// Create or get group
			group := models.Group{Name: groupName}
			if err := tx.FirstOrCreate(&group, models.Group{Name: groupName}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create/get group %s: %v", groupName, err)
			}

			// Create words and associate with group
			for _, wordData := range words {
				word := models.Word{
					Japanese: wordData.Japanese,
					Romaji:   wordData.Romaji,
					English:  wordData.English,
					Parts:    wordData.Parts,
				}

				// Create or get word
				if err := tx.FirstOrCreate(&word, models.Word{Japanese: wordData.Japanese}).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create/get word %s: %v", wordData.Japanese, err)
				}

				// Associate word with group
				if err := tx.Model(&group).Association("Words").Append(&word); err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to associate word %s with group %s: %v", wordData.Japanese, groupName, err)
				}
			}
			fmt.Printf("Successfully seeded group '%s' from %s\n", groupName, fileName)
		} else if strings.HasSuffix(fileName, ".sql") {
			// Execute SQL file content
			if err := tx.Exec(string(content)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute SQL seed file %s: %v", fileName, err)
			}
			fmt.Printf("Successfully executed SQL seed file: %s\n", fileName)
		} else {
			tx.Rollback() // Should not happen with current glob, but good practice
			return fmt.Errorf("unsupported seed file type: %s", fileName)
		}

		// Commit transaction for this file
		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("failed to commit seed data for %s: %v", fileName, err)
		}
	}

	fmt.Println("All seed data processed successfully")
	return nil
}

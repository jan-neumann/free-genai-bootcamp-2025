//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/magefile/mage/mg"
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

	fmt.Println("Database initialized successfully at", dbPath)
	return nil
}

// Migrate runs all database migrations
func (DB) Migrate() error {
	fmt.Println("Running database migrations...")

	// Get all migration files
	files, err := filepath.Glob("db/migrations/*.sql")
	if err != nil {
		return fmt.Errorf("failed to find migration files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found in db/migrations")
	}

	// Sort files by name (they should be prefixed with numbers)
	sort.Strings(files)

	// TODO: Implement actual migration logic
	// For now, just print the files that would be run
	for _, file := range files {
		fmt.Printf("Would run migration: %s\n", file)
	}

	return nil
}

// Seed imports seed data into the database
func (DB) Seed() error {
	fmt.Println("Seeding database...")

	// Get all seed files
	files, err := filepath.Glob("db/seeds/*.json")
	if err != nil {
		return fmt.Errorf("failed to find seed files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no seed files found in db/seeds")
	}

	// TODO: Implement actual seeding logic
	// For now, just print the files that would be processed
	for _, file := range files {
		groupName := strings.TrimSuffix(filepath.Base(file), ".json")
		fmt.Printf("Would seed group '%s' from %s\n", groupName, file)
	}

	return nil
}

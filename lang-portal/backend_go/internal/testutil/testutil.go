package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"lang-portal/backend_go/internal/models"
)

// SetupTestDB creates a new in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.Word{},
		&models.Group{},
		&models.StudyActivity{},
		&models.StudySession{},
		&models.WordReview{},
	)
	require.NoError(t, err)

	return db
}

// CreateTestWord creates a test word with default values
func CreateTestWord(t *testing.T, db *gorm.DB) *models.Word {
	word := &models.Word{
		Japanese: "テスト",
		Romaji:   "tesuto",
		English:  "test",
		Parts:    models.StringSlice{"noun"},
	}
	err := db.Create(word).Error
	require.NoError(t, err)
	return word
}

// CreateTestGroup creates a test group with default values
func CreateTestGroup(t *testing.T, db *gorm.DB) *models.Group {
	group := &models.Group{
		Name: "Test Group",
	}
	err := db.Create(group).Error
	require.NoError(t, err)
	return group
}

// CreateTestStudyActivity creates a test study activity with default values
func CreateTestStudyActivity(t *testing.T, db *gorm.DB) *models.StudyActivity {
	activity := &models.StudyActivity{
		Name:        "Test Activity",
		Description: "Test Description",
	}
	err := db.Create(activity).Error
	require.NoError(t, err)
	return activity
}

// CreateTestStudySession creates a test study session with default values
func CreateTestStudySession(t *testing.T, db *gorm.DB, groupID, activityID uint) *models.StudySession {
	session := &models.StudySession{
		GroupID:         groupID,
		StudyActivityID: activityID,
		CreatedAt:       time.Now(),
	}
	err := db.Create(session).Error
	require.NoError(t, err)
	return session
}

// CreateTestWordReview creates a test word review with default values
func CreateTestWordReview(t *testing.T, db *gorm.DB, wordID, sessionID uint) *models.WordReview {
	review := &models.WordReview{
		WordID:         wordID,
		StudySessionID: sessionID,
		Correct:        true,
		CreatedAt:      time.Now(),
	}
	err := db.Create(review).Error
	require.NoError(t, err)
	return review
}

// AssertTimeApprox asserts that two times are approximately equal (within 1 second)
func AssertTimeApprox(t *testing.T, expected, actual time.Time) {
	assert.True(t, expected.Sub(actual) < time.Second || actual.Sub(expected) < time.Second,
		"expected time %v to be approximately equal to %v", expected, actual)
}

// AssertError asserts that an error matches the expected error
func AssertError(t *testing.T, expected, actual error) {
	if expected == nil {
		assert.NoError(t, actual)
		return
	}
	assert.EqualError(t, actual, expected.Error())
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	err := db.Migrator().DropTable(
		&models.WordReview{},
		&models.StudySession{},
		&models.StudyActivity{},
		&models.Group{},
		&models.Word{},
	)
	require.NoError(t, err)
}

package repository

import (
	"testing"

	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWordRepo(t *testing.T) (*WordRepository, func()) {
	db := testutil.SetupTestDB(t)
	repo := NewWordRepository(db)
	cleanup := func() { testutil.CleanupTestDB(t, db) }
	return repo, cleanup
}

func TestWordRepository_CreateAndGet(t *testing.T) {
	repo, cleanup := setupWordRepo(t)
	defer cleanup()

	word := &models.Word{
		Japanese: "こんにちは",
		Romaji:   "konnichiwa",
		English:  "hello",
		Parts:    models.StringSlice{"greeting"},
	}
	err := repo.Create(word)
	require.NoError(t, err)
	assert.NotZero(t, word.ID)

	fetched, err := repo.GetByID(word.ID)
	require.NoError(t, err)
	assert.Equal(t, word.Japanese, fetched.Japanese)
	assert.Equal(t, word.English, fetched.English)
}

func TestWordRepository_Update(t *testing.T) {
	repo, cleanup := setupWordRepo(t)
	defer cleanup()
	word := &models.Word{
		Japanese: "ありがとう",
		Romaji:   "arigatou",
		English:  "thanks",
		Parts:    models.StringSlice{"greeting"},
	}
	err := repo.Create(word)
	require.NoError(t, err)
	word.English = "thank you"
	err = repo.Update(word)
	require.NoError(t, err)
	assert.Equal(t, "thank you", word.English)
}

func TestWordRepository_Delete(t *testing.T) {
	repo, cleanup := setupWordRepo(t)
	defer cleanup()
	word := &models.Word{
		Japanese: "さようなら",
		Romaji:   "sayounara",
		English:  "goodbye",
		Parts:    models.StringSlice{"greeting"},
	}
	err := repo.Create(word)
	require.NoError(t, err)
	err = repo.Delete(word.ID)
	require.NoError(t, err)
	_, err = repo.GetByID(word.ID)
	assert.Error(t, err)
}

func TestWordRepository_List(t *testing.T) {
	repo, cleanup := setupWordRepo(t)
	defer cleanup()
	for i := 0; i < 15; i++ {
		word := &models.Word{
			Japanese: "単語" + string(rune('A'+i)),
			Romaji:   "tango" + string(rune('A'+i)),
			English:  "word" + string(rune('A'+i)),
			Parts:    models.StringSlice{"noun"},
		}
		require.NoError(t, repo.Create(word))
	}
	params := PaginationParams{Page: 1, PageSize: 10}
	result, err := repo.List(params)
	require.NoError(t, err)
	assert.Equal(t, 10, len(result.Items))
	assert.Equal(t, int64(15), result.TotalItems)
	assert.Equal(t, 2, result.TotalPages)
}

func TestWordRepository_Stats(t *testing.T) {
	repo, cleanup := setupWordRepo(t)
	defer cleanup()
	word := &models.Word{
		Japanese: "テスト",
		Romaji:   "tesuto",
		English:  "test",
		Parts:    models.StringSlice{"noun"},
	}
	require.NoError(t, repo.Create(word))
	// Add reviews
	db := repo.db
	db.Create(&models.WordReview{WordID: word.ID, Correct: true})
	db.Create(&models.WordReview{WordID: word.ID, Correct: false})
	correct, wrong, err := repo.GetStudyStats(word.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), correct)
	assert.Equal(t, int64(1), wrong)
}

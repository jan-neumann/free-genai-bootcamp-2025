package service

import (
	"errors"
	"testing"

	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockWordRepository is a mock implementation of WordRepositoryInterface
type mockWordRepository struct {
	mock.Mock
}

func (m *mockWordRepository) Create(word *models.Word) error {
	args := m.Called(word)
	return args.Error(0)
}

func (m *mockWordRepository) GetByID(id uint) (*models.Word, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Word), args.Error(1)
}

func (m *mockWordRepository) List(params repository.PaginationParams) (*repository.PaginatedResult[models.Word], error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[models.Word]), args.Error(1)
}

func (m *mockWordRepository) Update(word *models.Word) error {
	args := m.Called(word)
	return args.Error(0)
}

func (m *mockWordRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockWordRepository) GetStudyStats(wordID uint) (correctCount int64, wrongCount int64, err error) {
	args := m.Called(wordID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *mockWordRepository) GetWordsByGroup(groupID uint, params repository.PaginationParams) (*repository.PaginatedResult[models.Word], error) {
	args := m.Called(groupID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PaginatedResult[models.Word]), args.Error(1)
}

func (m *mockWordRepository) GetTotalWordCount() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockWordRepository) GetStudiedWordCount() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockWordRepository) GetByJapanese(japanese string) (*models.Word, error) {
	args := m.Called(japanese)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Word), args.Error(1)
}

func (m *mockWordRepository) GetWordsByGroupRaw(groupID uint) ([]models.Word, error) {
	args := m.Called(groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Word), args.Error(1)
}

func TestWordService_GetWord(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil) // Other repos are nil as they are not used by WordService's GetWord
	wordService := NewWordService(baseService)

	testWordID := uint(1)
	expectedWord := &models.Word{
		ID:       testWordID,
		Japanese: "こんにちは",
		Romaji:   "Konnichiwa",
		English:  "Hello",
		Groups:   []models.Group{{ID: 1, Name: "Greetings"}},
	}

	// Setup mock expectations
	mockRepo.On("GetByID", testWordID).Return(expectedWord, nil)
	mockRepo.On("GetStudyStats", testWordID).Return(int64(10), int64(2), nil)

	// Call the service method
	wordDetail, err := wordService.GetWord(testWordID)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, wordDetail)
	assert.Equal(t, expectedWord.ID, wordDetail.ID)
	assert.Equal(t, expectedWord.Japanese, wordDetail.Japanese)
	assert.Equal(t, int64(10), wordDetail.StudyStats.CorrectCount)
	assert.Equal(t, int64(2), wordDetail.StudyStats.WrongCount)
	assert.Len(t, wordDetail.Groups, 1)
	assert.Equal(t, expectedWord.Groups[0].Name, wordDetail.Groups[0].Name)

	// Verify that all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestWordService_GetWord_NotFound(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(2)

	// Setup mock expectations for repository.ErrNotFound
	mockRepo.On("GetByID", testWordID).Return(nil, repository.ErrNotFound)

	// Call the service method
	wordDetail, err := wordService.GetWord(testWordID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, wordDetail)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, serviceErr.Code)

	// Verify that all expectations were met
	mockRepo.AssertExpectations(t)
}

func TestWordService_CreateWord(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	newWord := &models.Word{
		Japanese: "新しい単語",
		Romaji:   "Atarashii tango",
		English:  "New word",
		Parts:    []string{"noun"},
	}

	// Setup mock expectation
	mockRepo.On("Create", newWord).Return(nil)

	// Call the service method
	err := wordService.CreateWord(newWord)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWordService_CreateWord_Error(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	newWord := &models.Word{
		Japanese: "テスト",
	}
	expectedError := errors.New("create failed")

	// Setup mock expectation
	mockRepo.On("Create", newWord).Return(expectedError)

	// Call the service method
	err := wordService.CreateWord(newWord)

	// Assertions
	assert.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	assert.Contains(t, serviceErr.Error(), expectedError.Error())
	mockRepo.AssertExpectations(t)
}

func TestWordService_ListWords(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	params := PaginationParams{Page: 1, PageSize: 10}
	repoParams := repository.PaginationParams{Page: 1, PageSize: 10}

	expectedRepoResult := &repository.PaginatedResult[models.Word]{
		Items: []models.Word{
			{ID: 1, Japanese: "こんにちは", Romaji: "Konnichiwa", English: "Hello"},
			{ID: 2, Japanese: "ありがとう", Romaji: "Arigato", English: "Thank you"},
		},
		TotalItems: 2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	// Mock expectations for List
	mockRepo.On("List", repoParams).Return(expectedRepoResult, nil)
	// Mock expectations for GetStudyStats for each word
	mockRepo.On("GetStudyStats", uint(1)).Return(int64(5), int64(1), nil)
	mockRepo.On("GetStudyStats", uint(2)).Return(int64(10), int64(0), nil)

	result, err := wordService.ListWords(params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, int64(2), result.TotalItems)
	assert.Equal(t, expectedRepoResult.Items[0].Japanese, result.Items[0].Japanese)
	assert.Equal(t, int64(5), result.Items[0].CorrectCount)
	assert.Equal(t, expectedRepoResult.Items[1].Romaji, result.Items[1].Romaji)
	assert.Equal(t, int64(10), result.Items[1].CorrectCount)
	mockRepo.AssertExpectations(t)
}

func TestWordService_ListWords_RepoError(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	params := PaginationParams{Page: 1, PageSize: 10}
	repoParams := repository.PaginationParams{Page: 1, PageSize: 10}
	expectedError := errors.New("list failed")

	mockRepo.On("List", repoParams).Return(nil, expectedError)

	result, err := wordService.ListWords(params)

	assert.Error(t, err)
	assert.Nil(t, result)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestWordService_ListWords_GetStudyStatsError(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	params := PaginationParams{Page: 1, PageSize: 10}
	repoParams := repository.PaginationParams{Page: 1, PageSize: 10}

	expectedRepoResult := &repository.PaginatedResult[models.Word]{
		Items: []models.Word{
			{ID: 1, Japanese: "こんにちは"},
			{ID: 2, Japanese: "ありがとう"},
		},
		TotalItems: 2,
	}

	statsError := errors.New("failed to get stats")

	mockRepo.On("List", repoParams).Return(expectedRepoResult, nil)
	mockRepo.On("GetStudyStats", uint(1)).Return(int64(5), int64(1), nil)        // First word stats succeed
	mockRepo.On("GetStudyStats", uint(2)).Return(int64(0), int64(0), statsError) // Second word stats fail

	result, err := wordService.ListWords(params)

	assert.Error(t, err)
	assert.Nil(t, result)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	assert.Contains(t, serviceErr.Error(), statsError.Error())
	mockRepo.AssertExpectations(t)
}

func TestWordService_UpdateWord(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(1)
	updateData := &models.Word{
		Japanese: "じゃあね",
		Romaji:   "Jaane",
		English:  "Bye",
	}

	existingWord := &models.Word{ID: testWordID, Japanese: "こんにちは", Romaji: "Konnichiwa", English: "Hello"}

	mockRepo.On("GetByID", testWordID).Return(existingWord, nil)
	mockRepo.On("Update", mock.MatchedBy(func(w *models.Word) bool {
		return w.ID == testWordID && w.Japanese == updateData.Japanese // Check a few fields
	})).Return(nil)

	err := wordService.UpdateWord(testWordID, updateData)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWordService_UpdateWord_RepoUpdateError(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(1)
	updateData := &models.Word{
		Japanese: "じゃあね",
		Romaji:   "Jaane",
		English:  "Bye",
	}

	existingWord := &models.Word{ID: testWordID, Japanese: "こんにちは", Romaji: "Konnichiwa", English: "Hello"}
	updateError := errors.New("failed to update in repo")

	mockRepo.On("GetByID", testWordID).Return(existingWord, nil)
	mockRepo.On("Update", mock.MatchedBy(func(w *models.Word) bool {
		return w.ID == testWordID && w.Japanese == updateData.Japanese
	})).Return(updateError)

	err := wordService.UpdateWord(testWordID, updateData)

	assert.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	assert.Contains(t, serviceErr.Error(), updateError.Error())
	mockRepo.AssertExpectations(t)
}

func TestWordService_UpdateWord_NotFound(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(99)
	updateData := &models.Word{Japanese: "Test"}

	mockRepo.On("GetByID", testWordID).Return(nil, repository.ErrNotFound)

	err := wordService.UpdateWord(testWordID, updateData)

	assert.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, serviceErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestWordService_DeleteWord(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(1)

	mockRepo.On("Delete", testWordID).Return(nil)

	err := wordService.DeleteWord(testWordID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWordService_DeleteWord_Error(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(1)
	expectedError := errors.New("delete failed")

	mockRepo.On("Delete", testWordID).Return(expectedError)

	err := wordService.DeleteWord(testWordID)

	assert.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestWordService_DeleteWord_NotFound(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testWordID := uint(99)
	mockRepo.On("Delete", testWordID).Return(repository.ErrNotFound)

	err := wordService.DeleteWord(testWordID)

	assert.Error(t, err)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeNotFound, serviceErr.Code)
	mockRepo.AssertExpectations(t)
}

func TestWordService_GetWordsByGroup(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testGroupID := uint(1)
	params := PaginationParams{Page: 1, PageSize: 5}
	repoParams := repository.PaginationParams{Page: params.Page, PageSize: params.PageSize}

	expectedRepoResult := &repository.PaginatedResult[models.Word]{
		Items: []models.Word{
			{ID: 1, Japanese: "犬", Romaji: "Inu", English: "Dog"},
		},
		TotalItems: 1,
	}

	mockRepo.On("GetWordsByGroup", testGroupID, repoParams).Return(expectedRepoResult, nil)
	mockRepo.On("GetStudyStats", uint(1)).Return(int64(3), int64(0), nil)

	result, err := wordService.GetWordsByGroup(testGroupID, params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.TotalItems)
	assert.Equal(t, int64(3), result.Items[0].CorrectCount)
	mockRepo.AssertExpectations(t)
}

func TestWordService_GetWordsByGroup_GetStudyStatsError(t *testing.T) {
	mockRepo := new(mockWordRepository)
	baseService := NewBaseService(mockRepo, nil, nil)
	wordService := NewWordService(baseService)

	testGroupID := uint(1)
	params := PaginationParams{Page: 1, PageSize: 5}
	repoParams := repository.PaginationParams{Page: params.Page, PageSize: params.PageSize}

	expectedRepoResult := &repository.PaginatedResult[models.Word]{
		Items: []models.Word{
			{ID: 1, Japanese: "犬"},
			{ID: 2, Japanese: "猫"},
		},
		TotalItems: 2,
	}
	statsError := errors.New("failed to get stats for group word")

	mockRepo.On("GetWordsByGroup", testGroupID, repoParams).Return(expectedRepoResult, nil)
	mockRepo.On("GetStudyStats", uint(1)).Return(int64(3), int64(0), nil)        // First word stats succeed
	mockRepo.On("GetStudyStats", uint(2)).Return(int64(0), int64(0), statsError) // Second word stats fail

	result, err := wordService.GetWordsByGroup(testGroupID, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	serviceErr, ok := err.(*ServiceError)
	assert.True(t, ok)
	assert.Equal(t, ErrCodeInternal, serviceErr.Code)
	assert.Contains(t, serviceErr.Error(), statsError.Error())
	mockRepo.AssertExpectations(t)
}

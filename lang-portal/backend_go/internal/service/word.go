package service

import (
	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/repository"
)

// WordService handles word-related business logic
type WordService struct {
	*BaseService
}

// NewWordService creates a new word service
func NewWordService(base *BaseService) *WordService {
	return &WordService{BaseService: base}
}

// Word represents a word with its study statistics
type Word struct {
	ID           uint   `json:"id"`
	Japanese     string `json:"japanese"`
	Romaji       string `json:"romaji"`
	English      string `json:"english"`
	CorrectCount int64  `json:"correct_count"`
	WrongCount   int64  `json:"wrong_count"`
}

// WordDetail represents detailed word information
type WordDetail struct {
	ID         uint   `json:"id"`
	Japanese   string `json:"japanese"`
	Romaji     string `json:"romaji"`
	English    string `json:"english"`
	StudyStats struct {
		CorrectCount int64 `json:"correct_count"`
		WrongCount   int64 `json:"wrong_count"`
	} `json:"study_stats"`
	Groups []GroupInfo `json:"groups"`
}

// GroupInfo represents basic group information
type GroupInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CreateWord creates a new word
func (s *WordService) CreateWord(word *models.Word) error {
	if err := s.wordRepo.Create(word); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to create word", err)
	}
	return nil
}

// GetWord retrieves a word by ID
func (s *WordService) GetWord(id uint) (*WordDetail, error) {
	word, err := s.wordRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch word", err)
	}

	// Calculate study statistics
	correctCount, wrongCount, err := s.wordRepo.GetStudyStats(id)
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get word statistics", err)
	}

	// Transform groups
	groups := make([]GroupInfo, len(word.Groups))
	for i, group := range word.Groups {
		groups[i] = GroupInfo{
			ID:   group.ID,
			Name: group.Name,
		}
	}

	return &WordDetail{
		ID:       word.ID,
		Japanese: word.Japanese,
		Romaji:   word.Romaji,
		English:  word.English,
		StudyStats: struct {
			CorrectCount int64 `json:"correct_count"`
			WrongCount   int64 `json:"wrong_count"`
		}{
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
		},
		Groups: groups,
	}, nil
}

// ListWords retrieves a paginated list of words
func (s *WordService) ListWords(params PaginationParams) (*PaginatedResult[Word], error) {
	result, err := s.wordRepo.List(repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to list words", err)
	}

	// Transform words
	words := make([]Word, len(result.Items))
	for i, w := range result.Items {
		correctCount, wrongCount, err := s.wordRepo.GetStudyStats(w.ID)
		if err != nil {
			return nil, NewServiceError(ErrCodeInternal, "Failed to get word statistics", err)
		}

		words[i] = Word{
			ID:           w.ID,
			Japanese:     w.Japanese,
			Romaji:       w.Romaji,
			English:      w.English,
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
		}
	}

	return NewPaginatedResult(words, result.TotalItems, params.Page, params.PageSize), nil
}

// UpdateWord updates an existing word
func (s *WordService) UpdateWord(id uint, word *models.Word) error {
	// Verify word exists
	existing, err := s.wordRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch word", err)
	}

	// Update fields
	existing.Japanese = word.Japanese
	existing.Romaji = word.Romaji
	existing.English = word.English

	if err := s.wordRepo.Update(existing); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to update word", err)
	}
	return nil
}

// DeleteWord deletes a word
func (s *WordService) DeleteWord(id uint) error {
	if err := s.wordRepo.Delete(id); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to delete word", err)
	}
	return nil
}

// GetWordsByGroup retrieves words belonging to a group
func (s *WordService) GetWordsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[Word], error) {
	result, err := s.wordRepo.GetWordsByGroup(groupID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get group words", err)
	}

	// Transform words
	words := make([]Word, len(result.Items))
	for i, w := range result.Items {
		correctCount, wrongCount, err := s.wordRepo.GetStudyStats(w.ID)
		if err != nil {
			return nil, NewServiceError(ErrCodeInternal, "Failed to get word statistics", err)
		}

		words[i] = Word{
			ID:           w.ID,
			Japanese:     w.Japanese,
			Romaji:       w.Romaji,
			English:      w.English,
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
		}
	}

	return NewPaginatedResult(words, result.TotalItems, params.Page, params.PageSize), nil
}

package service

import (
	"lang-portal/backend_go/internal/models"
	"lang-portal/backend_go/internal/repository"
)

// GroupService handles group-related business logic
type GroupService struct {
	*BaseService
}

// NewGroupService creates a new group service
func NewGroupService(base *BaseService) *GroupService {
	return &GroupService{BaseService: base}
}

// Group represents a word group with its word count
type Group struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	WordCount int    `json:"word_count"`
}

// GroupDetail represents detailed group information
type GroupDetail struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	WordCount int    `json:"word_count"`
}

// GroupWordRaw represents a simplified word in a group (for raw endpoint)
type GroupWordRaw struct {
	ID       uint   `json:"id"`
	Japanese string `json:"japanese"`
	Romaji   string `json:"romaji"`
	English  string `json:"english"`
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(group *models.Group) error {
	// Check if group with same name exists
	existing, err := s.groupRepo.GetByName(group.Name)
	if err != nil && err != repository.ErrNotFound {
		return NewServiceError(ErrCodeInternal, "Failed to check for existing group", err)
	}
	if existing != nil {
		return NewServiceError(ErrCodeInvalidInput, "A group with this name already exists", nil)
	}

	if err := s.groupRepo.Create(group); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to create group", err)
	}
	return nil
}

// GetGroup retrieves a group by ID
func (s *GroupService) GetGroup(id uint) (*GroupDetail, error) {
	group, err := s.groupRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return nil, NewServiceError(ErrCodeInternal, "Failed to fetch group", err)
	}

	return &GroupDetail{
		ID:        group.ID,
		Name:      group.Name,
		WordCount: len(group.Words),
	}, nil
}

// ListGroups retrieves a paginated list of groups
func (s *GroupService) ListGroups(params PaginationParams) (*PaginatedResult[Group], error) {
	result, err := s.groupRepo.List(repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to list groups", err)
	}

	// Transform groups
	groups := make([]Group, len(result.Items))
	for i, g := range result.Items {
		groups[i] = Group{
			ID:        g.ID,
			Name:      g.Name,
			WordCount: len(g.Words),
		}
	}

	return NewPaginatedResult(groups, result.TotalItems, params.Page, params.PageSize), nil
}

// UpdateGroup updates an existing group
func (s *GroupService) UpdateGroup(id uint, group *models.Group) error {
	// Verify group exists
	existing, err := s.groupRepo.GetByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch group", err)
	}

	// Check if new name conflicts with existing group
	if existing.Name != group.Name {
		conflicting, err := s.groupRepo.GetByName(group.Name)
		if err != nil && err != repository.ErrNotFound {
			return NewServiceError(ErrCodeInternal, "Failed to check for existing group", err)
		}
		if conflicting != nil {
			return NewServiceError(ErrCodeInvalidInput, "A group with this name already exists", nil)
		}
	}

	// Update fields
	existing.Name = group.Name

	if err := s.groupRepo.Update(existing); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to update group", err)
	}
	return nil
}

// DeleteGroup deletes a group
func (s *GroupService) DeleteGroup(id uint) error {
	if err := s.groupRepo.Delete(id); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to delete group", err)
	}
	return nil
}

// AddWordToGroup adds a word to a group
func (s *GroupService) AddWordToGroup(groupID, wordID uint) error {
	// Verify group exists
	if _, err := s.groupRepo.GetByID(groupID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch group", err)
	}

	// Verify word exists
	if _, err := s.wordRepo.GetByID(wordID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch word", err)
	}

	if err := s.groupRepo.AddWord(groupID, wordID); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to add word to group", err)
	}
	return nil
}

// RemoveWordFromGroup removes a word from a group
func (s *GroupService) RemoveWordFromGroup(groupID, wordID uint) error {
	// Verify group exists
	if _, err := s.groupRepo.GetByID(groupID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch group", err)
	}

	// Verify word exists
	if _, err := s.wordRepo.GetByID(wordID); err != nil {
		if err == repository.ErrNotFound {
			return NewServiceError(ErrCodeNotFound, "Word not found", err)
		}
		return NewServiceError(ErrCodeInternal, "Failed to fetch word", err)
	}

	if err := s.groupRepo.RemoveWord(groupID, wordID); err != nil {
		return NewServiceError(ErrCodeInternal, "Failed to remove word from group", err)
	}
	return nil
}

// GetGroupStudyStats retrieves study statistics for a group
func (s *GroupService) GetGroupStudyStats(id uint) (totalSessions, totalReviews, correctReviews int, err error) {
	totalSessions, totalReviews, correctReviews, err = s.groupRepo.GetStudyStats(id)
	if err != nil {
		if err == repository.ErrNotFound {
			return 0, 0, 0, NewServiceError(ErrCodeNotFound, "Group not found", err)
		}
		return 0, 0, 0, NewServiceError(ErrCodeInternal, "Failed to get group statistics", err)
	}
	return
}

// GetGroupsByWord retrieves groups containing a specific word
func (s *GroupService) GetGroupsByWord(wordID uint, params PaginationParams) (*PaginatedResult[Group], error) {
	result, err := s.groupRepo.GetGroupsByWord(wordID, repository.PaginationParams{
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get groups by word", err)
	}

	// Transform groups
	groups := make([]Group, len(result.Items))
	for i, g := range result.Items {
		wordCount := int64(len(g.Words)) // Calculate word count from the Words relationship
		groups[i] = Group{
			ID:        g.ID,
			Name:      g.Name,
			WordCount: int(wordCount),
		}
	}

	return NewPaginatedResult(groups, result.TotalItems, params.Page, params.PageSize), nil
}

// GetWordsRaw retrieves a simplified list of words in a group (id, japanese, romaji, english only)
func (s *GroupService) GetWordsRaw(groupID uint) ([]GroupWordRaw, error) {
	words, err := s.wordRepo.GetWordsByGroupRaw(groupID)
	if err != nil {
		return nil, NewServiceError(ErrCodeInternal, "Failed to get group words", err)
	}

	// Transform to raw word format
	rawWords := make([]GroupWordRaw, len(words))
	for i, w := range words {
		rawWords[i] = GroupWordRaw{
			ID:       w.ID,
			Japanese: w.Japanese,
			Romaji:   w.Romaji,
			English:  w.English,
		}
	}

	return rawWords, nil
}

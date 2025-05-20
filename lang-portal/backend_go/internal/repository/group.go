package repository

import (
	"lang-portal/backend_go/internal/models"

	"gorm.io/gorm"
)

// WordGroup represents the many-to-many relationship between words and groups
type WordGroup struct {
	GroupID uint `gorm:"primaryKey"`
	WordID  uint `gorm:"primaryKey"`
}

// GroupRepository handles database operations for groups
type GroupRepository struct {
	*BaseRepository
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{BaseRepository: NewBaseRepository(db)}
}

// Create creates a new group
func (r *GroupRepository) Create(group *models.Group) error {
	if err := group.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Create(group).Error
}

// GetByID retrieves a group by ID
func (r *GroupRepository) GetByID(id uint) (*models.Group, error) {
	var group models.Group
	if err := r.db.Preload("Words").Preload("Words.Reviews").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

// GetByName retrieves a group by name
func (r *GroupRepository) GetByName(name string) (*models.Group, error) {
	var group models.Group
	if err := r.db.Where("name = ?", name).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &group, nil
}

// List retrieves a paginated list of groups
func (r *GroupRepository) List(params PaginationParams) (*PaginatedResult[models.Group], error) {
	var groups []models.Group
	var total int64

	query := r.db.Model(&models.Group{})
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Words").Preload("Words.Reviews").Find(&groups).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.Group]{
		Items:      groups,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates a group
func (r *GroupRepository) Update(group *models.Group) error {
	if err := group.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Save(group).Error
}

// Delete deletes a group and its associations
func (r *GroupRepository) Delete(id uint) error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		// Delete word-group associations
		if err := tx.Where("group_id = ?", id).Delete(&WordGroup{}).Error; err != nil {
			return err
		}
		// Delete the group
		return tx.Delete(&models.Group{}, id).Error
	})
}

// AddWord adds a word to a group
func (r *GroupRepository) AddWord(groupID, wordID uint) error {
	wordGroup := WordGroup{
		GroupID: groupID,
		WordID:  wordID,
	}
	return r.db.Create(&wordGroup).Error
}

// RemoveWord removes a word from a group
func (r *GroupRepository) RemoveWord(groupID, wordID uint) error {
	return r.db.Where("group_id = ? AND word_id = ?", groupID, wordID).Delete(&WordGroup{}).Error
}

// GetStudyStats retrieves study statistics for a group
func (r *GroupRepository) GetStudyStats(id uint) (totalSessions, totalReviews, correctReviews int, err error) {
	var group models.Group
	if err := r.db.Preload("Words").Preload("Words.Reviews").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, 0, 0, ErrNotFound
		}
		return 0, 0, 0, err
	}
	totalSessions, totalReviews, correctReviews = group.GetStudyStats()
	return
}

// GetGroupsByWord retrieves groups containing a specific word
func (r *GroupRepository) GetGroupsByWord(wordID uint, params PaginationParams) (*PaginatedResult[models.Group], error) {
	var groups []models.Group
	var total int64

	query := r.db.Model(&models.Group{}).
		Joins("JOIN word_groups ON word_groups.group_id = groups.id").
		Where("word_groups.word_id = ?", wordID)

	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Words").Preload("Words.Reviews").Find(&groups).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.Group]{
		Items:      groups,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTotalGroupCount returns the total number of groups
func (r *GroupRepository) GetTotalGroupCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Group{}).Count(&count).Error
	return count, err
}

// GetActiveGroupCount returns the number of groups that have been studied
func (r *GroupRepository) GetActiveGroupCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Group{}).
		Joins("JOIN word_groups ON word_groups.group_id = groups.id").
		Joins("JOIN word_review_items ON word_review_items.word_id = word_groups.word_id").
		Distinct().
		Count(&count).Error
	return count, err
}

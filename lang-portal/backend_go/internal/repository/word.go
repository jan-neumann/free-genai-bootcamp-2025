package repository

import (
	"lang-portal/backend_go/internal/models"

	"gorm.io/gorm"
)

// WordRepository handles database operations for words
type WordRepository struct {
	*BaseRepository
}

// NewWordRepository creates a new word repository
func NewWordRepository(db *gorm.DB) *WordRepository {
	return &WordRepository{BaseRepository: NewBaseRepository(db)}
}

// Create creates a new word
func (r *WordRepository) Create(word *models.Word) error {
	if err := word.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Create(word).Error
}

// GetByID retrieves a word by ID
func (r *WordRepository) GetByID(id uint) (*models.Word, error) {
	var word models.Word
	if err := r.db.Preload("Groups").Preload("Reviews").First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &word, nil
}

// GetByJapanese retrieves a word by Japanese text
func (r *WordRepository) GetByJapanese(japanese string) (*models.Word, error) {
	var word models.Word
	if err := r.db.Where("japanese = ?", japanese).First(&word).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &word, nil
}

// List retrieves a paginated list of words
func (r *WordRepository) List(params PaginationParams) (*PaginatedResult[models.Word], error) {
	var words []models.Word
	var total int64

	query := r.db.Model(&models.Word{})
	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Groups").Preload("Reviews").Find(&words).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.Word]{
		Items:      words,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Update updates a word
func (r *WordRepository) Update(word *models.Word) error {
	if err := word.Validate(); err != nil {
		return ErrInvalidInput
	}
	return r.db.Save(word).Error
}

// Delete deletes a word
func (r *WordRepository) Delete(id uint) error {
	return r.WithTransaction(func(tx *gorm.DB) error {
		// Delete word-group associations
		if err := tx.Where("word_id = ?", id).Delete(&models.Word{}).Error; err != nil {
			return err
		}
		// Delete word reviews
		if err := tx.Where("word_id = ?", id).Delete(&models.WordReview{}).Error; err != nil {
			return err
		}
		// Delete the word
		return tx.Delete(&models.Word{}, id).Error
	})
}

// GetStudyStats retrieves study statistics for a word
func (r *WordRepository) GetStudyStats(id uint) (correctCount, wrongCount int, err error) {
	var word models.Word
	if err := r.db.Preload("Reviews").First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, 0, ErrNotFound
		}
		return 0, 0, err
	}
	correctCount, wrongCount = word.GetStudyStats()
	return
}

// GetWordsByGroup retrieves words belonging to a group
func (r *WordRepository) GetWordsByGroup(groupID uint, params PaginationParams) (*PaginatedResult[models.Word], error) {
	var words []models.Word
	var total int64

	query := r.db.Model(&models.Word{}).
		Joins("JOIN word_groups ON word_groups.word_id = words.id").
		Where("word_groups.group_id = ?", groupID)

	paginatedQuery, err := r.Paginate(query, params)
	if err != nil {
		return nil, err
	}

	if err := paginatedQuery.Preload("Groups").Preload("Reviews").Find(&words).Error; err != nil {
		return nil, err
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := (int(total) + params.PageSize - 1) / params.PageSize
	return &PaginatedResult[models.Word]{
		Items:      words,
		TotalItems: total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTotalWordCount returns the total number of words
func (r *WordRepository) GetTotalWordCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Word{}).Count(&count).Error
	return count, err
}

// GetStudiedWordCount returns the number of words that have been studied
func (r *WordRepository) GetStudiedWordCount() (int64, error) {
	var count int64
	err := r.db.Model(&models.Word{}).
		Joins("JOIN word_review_items ON word_review_items.word_id = words.id").
		Distinct().
		Count(&count).Error
	return count, err
}

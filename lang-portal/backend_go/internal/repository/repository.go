package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Common errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("record already exists")
)

// PaginationParams represents common pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginatedResult represents a paginated result set
type PaginatedResult[T any] struct {
	Items      []T
	TotalItems int64
	Page       int
	PageSize   int
	TotalPages int
}

// BaseRepository provides common repository functionality
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// WithTransaction executes the given function within a transaction
func (r *BaseRepository) WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Paginate applies pagination to a query
func (r *BaseRepository) Paginate(query *gorm.DB, params PaginationParams) (*gorm.DB, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	return query.Offset(offset).Limit(params.PageSize), nil
}

// TimeRange represents a time range for filtering
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// WithTimeRange applies a time range filter to a query
func (r *BaseRepository) WithTimeRange(query *gorm.DB, field string, tr TimeRange) *gorm.DB {
	if !tr.Start.IsZero() {
		query = query.Where(field+" >= ?", tr.Start)
	}
	if !tr.End.IsZero() {
		query = query.Where(field+" <= ?", tr.End)
	}
	return query
}

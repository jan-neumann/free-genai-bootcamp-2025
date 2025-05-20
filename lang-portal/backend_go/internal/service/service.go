package service

import (
	"lang-portal/backend_go/internal/repository"
)

// BaseService provides common service functionality
type BaseService struct {
	wordRepo  *repository.WordRepository
	groupRepo *repository.GroupRepository
	studyRepo *repository.StudyRepository
}

// NewBaseService creates a new base service
func NewBaseService(wordRepo *repository.WordRepository, groupRepo *repository.GroupRepository, studyRepo *repository.StudyRepository) *BaseService {
	return &BaseService{
		wordRepo:  wordRepo,
		groupRepo: groupRepo,
		studyRepo: studyRepo,
	}
}

// ServiceError represents a service-level error
type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Common error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeInternal     = "INTERNAL_ERROR"
)

// NewServiceError creates a new service error
func NewServiceError(code, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

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

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](items []T, total int64, page, pageSize int) *PaginatedResult[T] {
	totalPages := (int(total) + pageSize - 1) / pageSize
	return &PaginatedResult[T]{
		Items:      items,
		TotalItems: total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

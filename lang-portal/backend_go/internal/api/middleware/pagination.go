package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// PaginationParams represents pagination parameters from the request
type PaginationParams struct {
	Page     int
	PageSize int
}

// PaginationMiddleware extracts pagination parameters from the request
func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get page parameter
		pageStr := c.DefaultQuery("page", strconv.Itoa(defaultPage))
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = defaultPage
		}

		// Get page size parameter
		pageSizeStr := c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize))
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			pageSize = defaultPageSize
		}
		if pageSize > maxPageSize {
			pageSize = maxPageSize
		}

		// Store pagination parameters in context
		c.Set("pagination", PaginationParams{
			Page:     page,
			PageSize: pageSize,
		})

		c.Next()
	}
}

// GetPaginationParams retrieves pagination parameters from the context
func GetPaginationParams(c *gin.Context) PaginationParams {
	if params, exists := c.Get("pagination"); exists {
		return params.(PaginationParams)
	}
	return PaginationParams{
		Page:     defaultPage,
		PageSize: defaultPageSize,
	}
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Items      []interface{} `json:"items"`
	Pagination struct {
		CurrentPage  int `json:"current_page"`
		TotalPages   int `json:"total_pages"`
		TotalItems   int `json:"total_items"`
		ItemsPerPage int `json:"items_per_page"`
	} `json:"pagination"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(items []interface{}, totalItems int, params PaginationParams) PaginatedResponse {
	totalPages := (totalItems + params.PageSize - 1) / params.PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginatedResponse{
		Items: items,
		Pagination: struct {
			CurrentPage  int `json:"current_page"`
			TotalPages   int `json:"total_pages"`
			TotalItems   int `json:"total_items"`
			ItemsPerPage int `json:"items_per_page"`
		}{
			CurrentPage:  params.Page,
			TotalPages:   totalPages,
			TotalItems:   totalItems,
			ItemsPerPage: params.PageSize,
		},
	}
}

package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds pagination parameters.
type Params struct {
	Page   int `form:"page"`
	Limit  int `form:"limit"`
}

// Offset computes the DB offset from page and limit.
func (p Params) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// FromQuery extracts pagination params from a Gin context.
func FromQuery(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return Params{Page: page, Limit: limit}
}

// Response wraps paginated results.
type Response[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

// NewResponse builds a paginated response.
func NewResponse[T any](data []T, total int64, params Params) Response[T] {
	totalPages := int(total) / params.Limit
	if int(total)%params.Limit != 0 {
		totalPages++
	}
	return Response[T]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}
}

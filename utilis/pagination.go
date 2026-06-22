package utilis

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage  int64 = 1
	defaultLimit int64 = 10
)

type Pagination struct {
	Page  int64
	Limit int64
}

// Skip returns the number of documents to skip for the current page.
func (p Pagination) Skip() int64 {
	return (p.Page - 1) * p.Limit
}

// GetPagination reads the "page" and "limit" query params, falling back to
// page 1 and 10 records per page when they are missing or invalid.
func GetPagination(c *gin.Context) Pagination {
	page, err := strconv.ParseInt(c.Query("page"), 10, 64)
	if err != nil || page < 1 {
		page = defaultPage
	}

	limit, err := strconv.ParseInt(c.Query("limit"), 10, 64)
	if err != nil || limit < 1 {
		limit = defaultLimit
	}

	return Pagination{Page: page, Limit: limit}
}

// Package pagination normalizes list-endpoint query params and wraps results
// in a consistent paginated envelope.
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

type Query struct {
	Page  int
	Limit int
}

// Offset returns the SQL OFFSET for this page.
func (q Query) Offset() int32 { return int32((q.Page - 1) * q.Limit) }

// LimitInt32 returns Limit as int32 for sqlc LIMIT params.
func (q Query) LimitInt32() int32 { return int32(q.Limit) }

// Parse reads "page" and "limit" query params, clamping to sane bounds.
func Parse(c *gin.Context) Query {
	page := DefaultPage
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	limit := DefaultLimit
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Query{Page: page, Limit: limit}
}

type Paginated struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"totalPages"`
}

// Build wraps items/total into the standard Paginated envelope.
func Build(items interface{}, total int64, q Query) Paginated {
	totalPages := 1
	if q.Limit > 0 {
		tp := int((total + int64(q.Limit) - 1) / int64(q.Limit))
		if tp > 1 {
			totalPages = tp
		}
	}
	return Paginated{Items: items, Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages}
}

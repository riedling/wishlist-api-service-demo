package httputil

import "github.com/gin-gonic/gin"

// Pagination holds parsed pagination query parameters, with sane defaults
// and upper bounds so handlers don't need to re-implement this logic.
type Pagination struct {
	Page    int `form:"page,default=1" binding:"omitempty,min=1"`
	PerPage int `form:"per_page,default=20" binding:"omitempty,min=1,max=100"`
}

// Offset returns the SQL/slice offset implied by the current page.
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// ToMeta builds a Meta struct from the pagination params and a known total
// item count.
func (p Pagination) ToMeta(totalItems int) *Meta {
	totalPages := 0
	if p.PerPage > 0 {
		totalPages = (totalItems + p.PerPage - 1) / p.PerPage
	}
	return &Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// ParsePagination binds pagination query parameters from the request. On
// failure it writes a 400 response and returns false.
func ParsePagination(c *gin.Context) (Pagination, bool) {
	var p Pagination
	if !BindQuery(c, &p) {
		return p, false
	}
	return p, true
}

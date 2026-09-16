package dto

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PaginationQuery is embedded into every list-query DTO (ListTeamsQuery,
// ListPlayersQuery, ...) so page/limit binding, defaults, and the offset
// calculation are defined exactly once.
type PaginationQuery struct {
	Page  int `form:"page" binding:"omitempty,min=1"`
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

// Normalize fills in defaults for whichever fields the client omitted.
// Gin's query binding leaves unset fields at their zero value, so this must
// run once after binding and before the value is used.
func (p *PaginationQuery) Normalize() {
	if p.Page == 0 {
		p.Page = DefaultPage
	}
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
}

// Offset computes the SQL OFFSET for the current page. Call Normalize first.
func (p PaginationQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}

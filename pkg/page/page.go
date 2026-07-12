package page

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Page struct {
	Limit  int
	Offset int
}

func New(limit, offset int) Page {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return Page{Limit: limit, Offset: offset}
}

type Result[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

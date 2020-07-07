package middleware

// PaginateKey is strongly typed key for indexing a request context.
type PaginateKey string

// Pagination wraps the common values required to paginate query results.
type Pagination struct {
	Height uint64 `json:"block_height"`
	Limit  uint64 `json:"limit"`
	Page   uint64 `json:"page"`
}

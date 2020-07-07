// Parse pagination information automatically for each Route.

package middleware

import (
	"context"
	"net/http"
	"strconv"
)

// Export a constant key that names the pagination index in a request context,
// so that consumers of the middleware can access the pagination data.
const (
	Key PaginateKey = PaginateKey("paginate")
)

// defaultNumber attempts to parse a get argument, returns 0 if the argument is
// either not present or unparseable.
func defaultNumber(s string, n uint64) uint64 {
	if n, err := strconv.ParseUint(s, 10, 64); err == nil {
		return n
	}
	return n
}

// min exists because Go sucks and doesn't provide one for anything other than
// floats.
func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// Paginate will look for common pagination get args in a URL and construction
// a Pagination object that can be used by endpoint handlers.
func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), Key, Pagination{
			Height: defaultNumber(r.URL.Query().Get("block_height"), 0),
			Limit:  defaultNumber(r.URL.Query().Get("limit"), 50),
			Page:   defaultNumber(r.URL.Query().Get("page"), 1) - 1,
		})))
	})
}

// GetPagination is a helper that performs the key extraction from the context
// of an http Request.
func GetPagination(r *http.Request) Pagination {
	ctx := r.Context()
	if pagination, ok := ctx.Value(Key).(Pagination); ok {
		return pagination
	}

	// If no pagination existed, or is missing entirely, return a default
	// pagination of 50 entries starting from the first page.
	return Pagination{
		Height: 0,
		Limit:  50,
		Page:   0,
	}
}

// PaginateList will automatically calculate the minimum and maximum for
// pagination bounds for some list length. Variables for this calculation are
// extracted from request parameters.
func PaginateList(r *http.Request, signedLength int) (uint64, uint64) {
	length := uint64(signedLength)
	pagination := GetPagination(r)
	from := min(pagination.Page*pagination.Limit, length)
	to := min(from+pagination.Limit, length)
	return from, to
}

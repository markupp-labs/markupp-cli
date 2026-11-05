package client

import (
	"net/url"
	"strconv"
)

// The server clamps limit and offset to these values instead of refusing them.
const (
	DefaultLimit = 10
	MaxLimit     = 100
	MaxOffset    = 10000
)

// Page is the limit and offset of a paginated listing. A zero field is not sent, so
// the server default applies.
type Page struct {
	Limit  int
	Offset int
}

func (p Page) apply(query url.Values) {
	if p.Limit > 0 {
		query.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		query.Set("offset", strconv.Itoa(p.Offset))
	}
}

// setNonEmpty sets key only when value is not empty, for optional filters.
func setNonEmpty(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

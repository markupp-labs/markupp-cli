package client

import (
	"net/url"
	"strconv"
)

// NoteFilter narrows the notes of a vault listing (GET /vaults/{id}/notes and its
// public twin). A nil or empty field is not sent.
type NoteFilter struct {
	Orphan    *bool
	Community *int64
	Sort      string
	Direction string
}

func (f NoteFilter) apply(query url.Values) {
	setBool(query, "orphan", f.Orphan)
	setInt(query, "community", f.Community)
	setNonEmpty(query, "sort", f.Sort)
	setNonEmpty(query, "direction", f.Direction)
}

// setBool sets key only when value is not nil.
func setBool(query url.Values, key string, value *bool) {
	if value != nil {
		query.Set(key, strconv.FormatBool(*value))
	}
}

// setInt sets key only when value is not nil.
func setInt(query url.Values, key string, value *int64) {
	if value != nil {
		query.Set(key, strconv.FormatInt(*value, 10))
	}
}

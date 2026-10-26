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

// GraphFilter narrows the graph of a vault (GET /vaults/{id}/graph and its public
// twin). A nil field is not sent, so the server default, true for the booleans,
// applies.
type GraphFilter struct {
	Orphans    *bool
	Unresolved *bool
	Inferred   *bool
	Community  *int64
}

func (f GraphFilter) apply(query url.Values) {
	setBool(query, "orphans", f.Orphans)
	setBool(query, "unresolved", f.Unresolved)
	setBool(query, "inferred", f.Inferred)
	setInt(query, "community", f.Community)
}

// LinkFilter narrows the links of a vault listing (GET /vaults/{id}/links and its
// public twin). An empty field is not sent.
type LinkFilter struct {
	Status string
	Type   string
}

func (f LinkFilter) apply(query url.Values) {
	setNonEmpty(query, "status", f.Status)
	setNonEmpty(query, "type", f.Type)
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

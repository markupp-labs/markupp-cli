package client

import "net/url"

// NoteFilter narrows the notes of a vault listing. A nil or empty field is not sent.
type NoteFilter struct {
	Orphan    *bool
	Community *int64
	Sort      string
	Direction string
}

func (NoteFilter) apply(url.Values) {}

// GraphFilter narrows the graph of a vault. A nil field is not sent, so the server
// default (true for the booleans) applies.
type GraphFilter struct {
	Orphans    *bool
	Unresolved *bool
	Inferred   *bool
	Community  *int64
}

func (GraphFilter) apply(url.Values) {}

// LinkFilter narrows the links of a vault listing. An empty field is not sent.
type LinkFilter struct {
	Status string
	Type   string
}

func (LinkFilter) apply(url.Values) {}

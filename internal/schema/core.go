// Package schema mirrors the JSON shapes of the Markupp server API, one file per
// resource. The types decode responses for the text output; --json never goes
// through them.
package schema

// ErrorResponse is the body of every error answer of the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NoteSummary identifies a note wherever another resource refers to one.
type NoteSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// NoteRank is the derived organization of a note. Every field is nil until the second
// stage of the indexer has seen the note.
type NoteRank struct {
	Degree    *int     `json:"degree"`
	PageRank  *float64 `json:"pagerank"`
	Community *int64   `json:"community"`
}

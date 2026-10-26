package schema

// Note is a note with its content and derived organization, the NoteV1Response of
// GET, POST and PUT on a note.
type Note struct {
	ID            string `json:"id"`
	VaultID       string `json:"vault_id"`
	Path          string `json:"path"`
	Content       string `json:"content"`
	ContentSHA256 string `json:"content_sha256"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	NoteRank
}

// NoteListItem is a note in the listing of a vault, without its content.
type NoteListItem struct {
	NoteSummary
	UpdatedAt     string `json:"updated_at"`
	ContentSHA256 string `json:"content_sha256"`
	NoteRank
}

// SearchHit is one result of the lexical search, without the note content.
type SearchHit struct {
	Note    NoteSummary `json:"note"`
	VaultID string      `json:"vault_id"`
	Rank    float64     `json:"rank"`
	Snippet string      `json:"snippet"`
}

// NoteCreateRequest is the body of POST /vaults/{id}/notes.
type NoteCreateRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// NoteUpdateRequest is the body of PUT /notes/{id}. LastModifiedAt is the updated_at
// the client last read; with Force the server overwrites whatever changed since.
type NoteUpdateRequest struct {
	Path           string `json:"path"`
	Content        string `json:"content"`
	LastModifiedAt string `json:"lastModifiedAt"`
	Force          bool   `json:"force"`
}

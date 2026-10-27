package client

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// ListNotes streams the notes of vault that pass filter, without their content, one
// page at a time.
func (c *Client) ListNotes(ctx context.Context, vault string, filter NoteFilter, page Page) (io.ReadCloser, error) {
	query := url.Values{}
	filter.apply(query)
	page.apply(query)
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("vaults", vault, "notes"), query: query})
}

// GetNote reads one note with its content and derived organization.
func (c *Client) GetNote(ctx context.Context, id string) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("notes", id)})
}

// SearchNotes runs a lexical search in vault, or in every vault of the account when
// vault is empty. The results carry a snippet, never the note content.
func (c *Client) SearchNotes(ctx context.Context, text, vault string, page Page) (io.ReadCloser, error) {
	query := url.Values{"q": {text}}
	setNonEmpty(query, "vault", vault)
	page.apply(query)
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("search", "notes"), query: query})
}

// CreateNote creates a note in vault and returns the created note.
func (c *Client) CreateNote(ctx context.Context, vault string, note schema.NoteCreateRequest) (io.ReadCloser, error) {
	body, err := marshalBody(note)
	if err != nil {
		return nil, err
	}
	return c.open(ctx, request{method: http.MethodPost, path: endpoint("vaults", vault, "notes"), body: body})
}

// UpdateNote writes path and content of note id. The server answers 409 when the note
// changed since update.LastModifiedAt and update.Force is false; that conflict is
// returned as is, never retried.
func (c *Client) UpdateNote(ctx context.Context, id string, update schema.NoteUpdateRequest) (io.ReadCloser, error) {
	body, err := marshalBody(update)
	if err != nil {
		return nil, err
	}
	return c.open(ctx, request{method: http.MethodPut, path: endpoint("notes", id), body: body})
}

// DeleteNote deletes note id.
func (c *Client) DeleteNote(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: endpoint("notes", id)})
}

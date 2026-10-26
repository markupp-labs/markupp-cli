package client

import (
	"context"
	"errors"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

var errNotesPending = errors.New("not implemented")

// ListNotes streams the notes of vault, without their content.
func (c *Client) ListNotes(context.Context, string, Page) (io.ReadCloser, error) {
	return nil, errNotesPending
}

// GetNote reads one note with its content.
func (c *Client) GetNote(context.Context, string) (io.ReadCloser, error) {
	return nil, errNotesPending
}

// SearchNotes runs a lexical search.
func (c *Client) SearchNotes(context.Context, string, string, Page) (io.ReadCloser, error) {
	return nil, errNotesPending
}

// CreateNote creates a note in vault.
func (c *Client) CreateNote(context.Context, string, schema.NoteCreateRequest) (io.ReadCloser, error) {
	return nil, errNotesPending
}

// UpdateNote writes a note.
func (c *Client) UpdateNote(context.Context, string, schema.NoteUpdateRequest) (io.ReadCloser, error) {
	return nil, errNotesPending
}

// DeleteNote deletes a note.
func (c *Client) DeleteNote(context.Context, string) error {
	return errNotesPending
}

package tui

import (
	"context"
	"encoding/json"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// payload is every shape the interface decodes, a type set instead of any.
type payload interface {
	schema.Note | schema.NoteListItem | schema.SearchHit | schema.Link | schema.Mention |
		schema.LinkSuggestion | schema.Graph | schema.Community | schema.User |
		schema.MyTenant | vaultRecord | memberRecord | inviteRecord | notificationFeed
}

// decodeRaw decodes one record, naming it when it does not have the expected shape.
func decodeRaw[T payload](raw json.RawMessage) (T, error) { return *new(T), *new(error) }

// decodeBody reads a whole object body and decodes it.
func decodeBody[T payload](body io.ReadCloser) (T, error) { return *new(T), *new(error) }

// fetchBody is a call that returns a body.
type fetchBody func(ctx context.Context) (io.ReadCloser, error)

// loader reads the records of a listing.
type loader func(ctx context.Context) ([]record, error)

// arrayOf reads an array body element by element into records.
func arrayOf(fetch fetchBody, decode func(json.RawMessage) (record, error)) loader {
	return *new(loader)
}

func collect(body io.ReadCloser, decode func(json.RawMessage) (record, error)) ([]record, error) {
	return *new([]record), *new(error)
}

// rowsOf adapts a typed row function to a decoder of raw records.
func rowsOf[T payload](row func(T) record) func(json.RawMessage) (record, error) {
	return *new(func(json.RawMessage) (record, error))
}

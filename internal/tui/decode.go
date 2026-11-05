package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// payload is every shape the interface decodes, a type set instead of any.
type payload interface {
	schema.Note | schema.NoteListItem | schema.SearchHit | schema.Link | schema.Mention |
		schema.LinkSuggestion | schema.Graph | schema.Community | schema.User |
		schema.MyTenant | schema.Vault | schema.Member | schema.PendingInvite | schema.NotificationFeed
}

// decodeRaw decodes one record, naming it when it does not have the expected shape.
func decodeRaw[T payload](raw json.RawMessage) (T, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, failure.Wrap(failure.Internal, fmt.Sprintf("malformed server record %.120s", raw), err)
	}
	return value, nil
}

// decodeBody reads a whole object body and decodes it.
func decodeBody[T payload](body io.ReadCloser) (T, error) {
	defer func() { _ = body.Close() }()
	raw, err := io.ReadAll(body)
	if err != nil {
		var zero T
		return zero, failure.WrapRead(failure.Unavailable, "reading the server response", err)
	}
	return decodeRaw[T](raw)
}

// fetchBody is a call that returns a body.
type fetchBody func(ctx context.Context) (io.ReadCloser, error)

// loader reads the records of a listing.
type loader func(ctx context.Context) ([]record, error)

// arrayOf reads an array body element by element into records.
func arrayOf(fetch fetchBody, decode func(json.RawMessage) (record, error)) loader {
	return func(ctx context.Context) ([]record, error) {
		body, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		return collect(body, decode)
	}
}

func collect(body io.ReadCloser, decode func(json.RawMessage) (record, error)) ([]record, error) {
	var records []record
	for raw, err := range client.Elements(body) {
		if err != nil {
			return nil, err
		}
		row, err := decode(raw)
		if err != nil {
			return nil, err
		}
		records = append(records, row)
	}
	return records, nil
}

// rowsOf adapts a typed row function to a decoder of raw records.
func rowsOf[T payload](row func(T) record) func(json.RawMessage) (record, error) {
	return func(raw json.RawMessage) (record, error) {
		value, err := decodeRaw[T](raw)
		if err != nil {
			return record{}, err
		}
		return row(value), nil
	}
}

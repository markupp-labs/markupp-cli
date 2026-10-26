package cli

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// NoteLinkReader reads what connects one note to the others: its links, its unlinked
// mentions and the links worth adding. *client.Client satisfies it.
type NoteLinkReader interface {
	NoteLinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteBacklinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteMentions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
	NoteBackmentions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
	NoteSuggestions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
}

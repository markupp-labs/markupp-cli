package cli

import (
	"context"
	"errors"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// PublicReader reads a published vault by its slug, without credentials.
type PublicReader interface {
	PublicNotes(ctx context.Context, slug string, filter client.NoteFilter, page client.Page) (io.ReadCloser, error)
	PublicNote(ctx context.Context, slug, id string) (io.ReadCloser, error)
	PublicSearch(ctx context.Context, slug, text string, page client.Page) (io.ReadCloser, error)
	PublicGraph(ctx context.Context, slug string, filter client.GraphFilter) (io.ReadCloser, error)
	PublicLinks(ctx context.Context, slug string, filter client.LinkFilter, page client.Page) (io.ReadCloser, error)
}

func (a *App) connectVault(context.Context, vaultFlags) (Services, vaultTarget, error) {
	return Services{}, vaultTarget{}, errors.New("not implemented")
}

func publicRead(io.ReadCloser, error) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

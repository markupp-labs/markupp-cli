package cli

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// PublicReader reads a published vault by its slug, without credentials. A command
// that offers --public (addVaultFlags with allowPublic) connects with connectVault and
// calls this port when the target has a PublicSlug, wrapping each call in publicRead.
type PublicReader interface {
	PublicNotes(ctx context.Context, slug string, filter client.NoteFilter, page client.Page) (io.ReadCloser, error)
	PublicNote(ctx context.Context, slug, id string) (io.ReadCloser, error)
	PublicSearch(ctx context.Context, slug, text string, page client.Page) (io.ReadCloser, error)
}

// connectVault resolves --vault and --public and connects the way the target needs:
// with the account credentials for a vault id, without any for a public slug.
func (a *App) connectVault(ctx context.Context, f vaultFlags) (Services, vaultTarget, error) {
	target, err := a.resolveVault(f)
	if err != nil {
		return Services{}, vaultTarget{}, err
	}
	svc, err := a.connect(ctx, Target{Vault: target.ID, Public: target.PublicSlug != ""})
	if err != nil {
		return Services{}, vaultTarget{}, err
	}
	return svc, target, nil
}

// publicReader connects without credentials and returns the reader of public vaults,
// for a command that takes --public alone.
func (a *App) publicReader(ctx context.Context) (PublicReader, error) {
	svc, err := a.connect(ctx, Target{Public: true})
	if err != nil {
		return nil, err
	}
	return publicOf(svc)
}

// publicOf returns the reader of public vaults of svc, and an internal failure when
// it is not wired.
func publicOf(svc Services) (PublicReader, error) {
	if svc.Public == nil {
		return nil, failure.New(failure.Internal, "", "the public reader is not wired, expected Services.Public")
	}
	return svc.Public, nil
}

// publicRead passes the answer of a PublicReader call through, naming the server
// feature when the server has no public routes:
//
//	body, err := publicRead(svc.Public.PublicNotes(ctx, target.PublicSlug, filter, page))
var publicRead = featureRead(featurePublication)

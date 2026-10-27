// Package tui is the interactive interface of mu, mu tui. It is modeled on the
// interaction of K9s: a header, a : command prompt, a / filter, a resource table, a
// view stack with breadcrumbs and a flash line. Nothing outside this package imports
// tview.
//
// The interface reaches the server only through the ports below, which the HTTP
// client of internal/client satisfies, so it is a consumer of the same client layer
// as the commands and never a parallel implementation.
package tui

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// NoteStore reads and writes notes.
type NoteStore interface {
	ListNotes(ctx context.Context, vault string, filter client.NoteFilter, page client.Page) (io.ReadCloser, error)
	GetNote(ctx context.Context, id string) (io.ReadCloser, error)
	SearchNotes(ctx context.Context, text, vault string, page client.Page) (io.ReadCloser, error)
	UpdateNote(ctx context.Context, id string, update schema.NoteUpdateRequest) (io.ReadCloser, error)
	DeleteNote(ctx context.Context, id string) error
}

// RelationReader reads what a note links to, what links to it, its mentions and the
// suggested links.
type RelationReader interface {
	NoteLinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteBacklinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteMentions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
	NoteSuggestions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
}

// GraphReader reads the neighborhood of a note and the organization of a vault.
type GraphReader interface {
	NoteGraph(ctx context.Context, id string, depth int) (io.ReadCloser, error)
	VaultCommunities(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error)
	VaultLinks(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error)
}

// AccountReader reads the account of the token and its tenants.
type AccountReader interface {
	GetAuthenticatedUser(ctx context.Context) (io.ReadCloser, error)
	ListMyTenants(ctx context.Context) (io.ReadCloser, error)
}

// VaultReader lists the vaults of the tenant and the members of one.
type VaultReader interface {
	ListVaults(ctx context.Context) (io.ReadCloser, error)
	ListVaultMembers(ctx context.Context, id string) (io.ReadCloser, error)
}

// TeamReader lists the members of the tenant and the invites addressed to the account.
type TeamReader interface {
	ListMembers(ctx context.Context) (io.ReadCloser, error)
	ListMyInvites(ctx context.Context) (io.ReadCloser, error)
}

// NotificationReader reads the notification feed of the account.
type NotificationReader interface {
	Notifications(ctx context.Context) (io.ReadCloser, error)
}

// Ports are the ports of one connection. A nil port is a resource this build of mu
// cannot reach; its views say so instead of failing.
type Ports struct {
	Notes         NoteStore
	Relations     RelationReader
	Graph         GraphReader
	Account       AccountReader
	Vaults        VaultReader
	Team          TeamReader
	Notifications NotificationReader
}

// Scope is the identity a connection acts as: a tenant of the account, or the public
// vault with PublicSlug, read without credentials.
type Scope struct {
	Tenant     string
	PublicSlug string
}

// Connector builds the ports of a scope. The interface calls it at start and again
// when the tenant changes or a public vault is opened.
type Connector func(ctx context.Context, scope Scope) (Ports, error)

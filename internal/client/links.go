package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// LinkFilter narrows a listing of links. An empty field is not sent.
type LinkFilter struct {
	// Status is resolved or unresolved; only the vault listings take it.
	Status string
	// Type is explicit, embed, typed or inferred.
	Type string
}

func (f LinkFilter) apply(query url.Values) {
	setNonEmpty(query, "status", f.Status)
	setNonEmpty(query, "type", f.Type)
}

// NoteLinks lists the links note id makes, GET /notes/{id}/links.
func (c *Client) NoteLinks(ctx context.Context, id, linkType string, page Page) (io.ReadCloser, error) {
	return c.listNoteRelation(ctx, id, "links", LinkFilter{Type: linkType}, page)
}

// NoteBacklinks lists the links to note id, GET /notes/{id}/backlinks.
func (c *Client) NoteBacklinks(ctx context.Context, id, linkType string, page Page) (io.ReadCloser, error) {
	return c.listNoteRelation(ctx, id, "backlinks", LinkFilter{Type: linkType}, page)
}

// NoteMentions lists the titles note id cites without linking, GET /notes/{id}/mentions.
func (c *Client) NoteMentions(ctx context.Context, id string, page Page) (io.ReadCloser, error) {
	return c.listNoteRelation(ctx, id, "mentions", LinkFilter{}, page)
}

// NoteBackmentions lists the notes that cite a title of note id without linking,
// GET /notes/{id}/backmentions.
func (c *Client) NoteBackmentions(ctx context.Context, id string, page Page) (io.ReadCloser, error) {
	return c.listNoteRelation(ctx, id, "backmentions", LinkFilter{}, page)
}

// NoteSuggestions lists the notes note id could link, GET /notes/{id}/suggestions.
func (c *Client) NoteSuggestions(ctx context.Context, id string, page Page) (io.ReadCloser, error) {
	return c.listNoteRelation(ctx, id, "suggestions", LinkFilter{}, page)
}

// VaultLinks lists the links of every note of vault, GET /vaults/{id}/links.
func (c *Client) VaultLinks(ctx context.Context, vault string, filter LinkFilter, page Page) (io.ReadCloser, error) {
	return c.listLinks(ctx, endpoint("vaults", vault, "links"), filter, page)
}

// PublicVaultLinks lists the links of the published vault slug, without credentials,
// GET /public/vaults/{slug}/links.
func (c *Client) PublicVaultLinks(ctx context.Context, slug string, filter LinkFilter, page Page) (io.ReadCloser, error) {
	return c.listLinks(ctx, endpoint("public", "vaults", slug, "links"), filter, page)
}

func (c *Client) listNoteRelation(ctx context.Context, id, relation string, filter LinkFilter, page Page) (io.ReadCloser, error) {
	return c.listLinks(ctx, endpoint("notes", id, relation), filter, page)
}

func (c *Client) listLinks(ctx context.Context, path string, filter LinkFilter, page Page) (io.ReadCloser, error) {
	query := url.Values{}
	filter.apply(query)
	page.apply(query)
	return c.open(ctx, request{method: http.MethodGet, path: path, query: query})
}

package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// The public reads mirror the vault routes under /public/vaults/{slug}. Build the
// Client with NewPublic, so no token and no tenant are sent.

func publicPath(slug string, segments ...string) string {
	return endpoint(append([]string{"public", "vaults", slug}, segments...)...)
}

func (c *Client) getPublic(ctx context.Context, path string, query url.Values) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: path, query: query})
}

// PublicNotes lists the notes of the public vault slug, without content, a JSON array.
func (c *Client) PublicNotes(ctx context.Context, slug string, filter NoteFilter, page Page) (io.ReadCloser, error) {
	query := url.Values{}
	filter.apply(query)
	page.apply(query)
	return c.getPublic(ctx, publicPath(slug, "notes"), query)
}

// PublicNote returns the note id of the public vault slug, with its content.
func (c *Client) PublicNote(ctx context.Context, slug, id string) (io.ReadCloser, error) {
	return c.getPublic(ctx, publicPath(slug, "notes", id), nil)
}

// PublicSearch searches the public vault slug for text, a JSON array of hits without
// note content.
func (c *Client) PublicSearch(ctx context.Context, slug, text string, page Page) (io.ReadCloser, error) {
	query := url.Values{"q": {text}}
	page.apply(query)
	return c.getPublic(ctx, publicPath(slug, "search"), query)
}

// PublicGraph returns the graph of the public vault slug.
func (c *Client) PublicGraph(ctx context.Context, slug string, filter GraphFilter) (io.ReadCloser, error) {
	query := url.Values{}
	filter.apply(query)
	return c.getPublic(ctx, publicPath(slug, "graph"), query)
}

// PublicLinks lists the links of the public vault slug, a JSON array.
func (c *Client) PublicLinks(ctx context.Context, slug string, filter LinkFilter, page Page) (io.ReadCloser, error) {
	query := url.Values{}
	filter.apply(query)
	page.apply(query)
	return c.getPublic(ctx, publicPath(slug, "links"), query)
}

package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// GraphFilter narrows a graph. The zero value asks for everything: the server shows
// orphans, unresolved nodes and inferred links unless told otherwise.
type GraphFilter struct {
	HideOrphans    bool
	HideUnresolved bool
	HideInferred   bool
	// Community keeps only the notes of that community when it is not nil.
	Community *int64
}

func (f GraphFilter) apply(query url.Values) {
	setFalse(query, "orphans", f.HideOrphans)
	setFalse(query, "unresolved", f.HideUnresolved)
	setFalse(query, "inferred", f.HideInferred)
	if f.Community != nil {
		query.Set("community", strconv.FormatInt(*f.Community, 10))
	}
}

// setFalse sends key=false when hide is set, and nothing otherwise.
func setFalse(query url.Values, key string, hide bool) {
	if hide {
		query.Set(key, "false")
	}
}

// VaultGraph reads the whole graph of vault, GET /vaults/{id}/graph.
func (c *Client) VaultGraph(ctx context.Context, vault string, filter GraphFilter) (io.ReadCloser, error) {
	return c.readGraph(ctx, endpoint("vaults", vault, "graph"), filter, url.Values{})
}

// PublicVaultGraph reads the graph of the published vault slug, without credentials,
// GET /public/vaults/{slug}/graph.
func (c *Client) PublicVaultGraph(ctx context.Context, slug string, filter GraphFilter) (io.ReadCloser, error) {
	return c.readGraph(ctx, endpoint("public", "vaults", slug, "graph"), filter, url.Values{})
}

// NoteGraph reads the notes up to depth hops from note id, GET /notes/{id}/graph. A
// depth of 0 is not sent, so the server default applies.
func (c *Client) NoteGraph(ctx context.Context, id string, depth int, filter GraphFilter) (io.ReadCloser, error) {
	query := url.Values{}
	if depth > 0 {
		query.Set("depth", strconv.Itoa(depth))
	}
	return c.readGraph(ctx, endpoint("notes", id, "graph"), filter, query)
}

func (c *Client) readGraph(ctx context.Context, path string, filter GraphFilter, query url.Values) (io.ReadCloser, error) {
	filter.apply(query)
	return c.open(ctx, request{method: http.MethodGet, path: path, query: query})
}

// VaultCommunities lists the communities of vault, largest first,
// GET /vaults/{id}/communities.
func (c *Client) VaultCommunities(ctx context.Context, vault string, page Page) (io.ReadCloser, error) {
	query := url.Values{}
	page.apply(query)
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("vaults", vault, "communities"), query: query})
}

// VaultCommunity reads community cid of vault, GET /vaults/{id}/communities/{cid}.
func (c *Client) VaultCommunity(ctx context.Context, vault string, cid int64) (io.ReadCloser, error) {
	path := endpoint("vaults", vault, "communities", strconv.FormatInt(cid, 10))
	return c.open(ctx, request{method: http.MethodGet, path: path})
}

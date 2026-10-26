package client

import (
	"context"
	"io"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// listingCall is one client method under test, with the request it must send.
type listingCall struct {
	name      string
	call      func(c *Client) (io.ReadCloser, error)
	wantPath  string
	wantQuery string
}

func runListingCalls(t *testing.T, calls []listingCall) {
	t.Helper()
	for _, tc := range calls {
		server := newFakeServer(t, replyJSON(200, `[{"id":"x"}]`))
		body, err := tc.call(newTestClient(t, server, "t1"))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got := readAll(t, body); got != `[{"id":"x"}]` {
			t.Errorf("%s: body = %q, want the server body unchanged", tc.name, got)
		}
		got := server.seen()[0]
		if got.method != "GET" || got.path != tc.wantPath || got.query != tc.wantQuery {
			t.Errorf("%s: request = %s %s?%s, want GET %s?%s", tc.name, got.method, got.path, got.query, tc.wantPath, tc.wantQuery)
		}
	}
}

func TestNoteRelationListings(t *testing.T) {
	ctx := context.Background()
	page := Page{Limit: 5, Offset: 10}
	runListingCalls(t, []listingCall{
		{"links", func(c *Client) (io.ReadCloser, error) { return c.NoteLinks(ctx, "n/1", "typed", page) },
			"/v1/notes/n%2F1/links", "limit=5&offset=10&type=typed"},
		{"links without filter", func(c *Client) (io.ReadCloser, error) { return c.NoteLinks(ctx, "n1", "", Page{}) },
			"/v1/notes/n1/links", ""},
		{"backlinks", func(c *Client) (io.ReadCloser, error) { return c.NoteBacklinks(ctx, "n1", "embed", Page{}) },
			"/v1/notes/n1/backlinks", "type=embed"},
		{"mentions", func(c *Client) (io.ReadCloser, error) { return c.NoteMentions(ctx, "n1", page) },
			"/v1/notes/n1/mentions", "limit=5&offset=10"},
		{"backmentions", func(c *Client) (io.ReadCloser, error) { return c.NoteBackmentions(ctx, "n1", Page{}) },
			"/v1/notes/n1/backmentions", ""},
		{"suggestions", func(c *Client) (io.ReadCloser, error) { return c.NoteSuggestions(ctx, "n1", Page{Limit: 3}) },
			"/v1/notes/n1/suggestions", "limit=3"},
	})
}

func TestVaultLinkListings(t *testing.T) {
	ctx := context.Background()
	filter := LinkFilter{Status: "unresolved", Type: "inferred"}
	runListingCalls(t, []listingCall{
		{"vault links", func(c *Client) (io.ReadCloser, error) { return c.VaultLinks(ctx, "v1", filter, Page{Limit: 2}) },
			"/v1/vaults/v1/links", "limit=2&status=unresolved&type=inferred"},
		{"vault links without filter", func(c *Client) (io.ReadCloser, error) { return c.VaultLinks(ctx, "v1", LinkFilter{}, Page{}) },
			"/v1/vaults/v1/links", ""},
		{"public links", func(c *Client) (io.ReadCloser, error) { return c.PublicVaultLinks(ctx, "my-blog", filter, Page{}) },
			"/v1/public/vaults/my-blog/links", "status=unresolved&type=inferred"},
	})
}

func TestPublicLinksSendNoCredentials(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	c, err := NewPublic(Options{BaseURL: server.URL, Tenant: "t1", Tokens: StaticToken("tok")})
	if err != nil {
		t.Fatalf("NewPublic(%s): %v", server.URL, err)
	}
	body, err := c.PublicVaultLinks(context.Background(), "blog", LinkFilter{}, Page{})
	if err != nil {
		t.Fatalf("PublicVaultLinks: %v", err)
	}
	_ = body.Close()
	if got := server.seen()[0]; got.auth != "" || got.tenant != "" {
		t.Errorf("public request carried auth %q and tenant %q, want neither", got.auth, got.tenant)
	}
}

func TestNoteLinksMapsNotFound(t *testing.T) {
	server := newFakeServer(t, replyJSON(404, `{"error":"not_found","message":"note n9 not found"}`))
	_, err := newTestClient(t, server, "").NoteLinks(context.Background(), "n9", "", Page{})
	if failure.KindOf(err) != failure.NotFound {
		t.Errorf("NoteLinks on a 404 = %v, want not found", err)
	}
}

package client

import (
	"context"
	"io"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestGraphReads(t *testing.T) {
	ctx := context.Background()
	community := int64(7)
	hideAll := GraphFilter{HideOrphans: true, HideUnresolved: true, HideInferred: true, Community: &community}
	runListingCalls(t, []listingCall{
		{"vault graph", func(c *Client) (io.ReadCloser, error) { return c.VaultGraph(ctx, "v1", GraphFilter{}) },
			"/v1/vaults/v1/graph", ""},
		{"vault graph filtered", func(c *Client) (io.ReadCloser, error) { return c.VaultGraph(ctx, "v1", hideAll) },
			"/v1/vaults/v1/graph", "community=7&inferred=false&orphans=false&unresolved=false"},
		{"public graph", func(c *Client) (io.ReadCloser, error) {
			return c.PublicVaultGraph(ctx, "blog", GraphFilter{HideInferred: true})
		}, "/v1/public/vaults/blog/graph", "inferred=false"},
		{"note graph", func(c *Client) (io.ReadCloser, error) {
			return c.NoteGraph(ctx, "n1", 2, GraphFilter{HideOrphans: true})
		},
			"/v1/notes/n1/graph", "depth=2&orphans=false"},
		{"note graph default depth", func(c *Client) (io.ReadCloser, error) { return c.NoteGraph(ctx, "n1", 0, GraphFilter{}) },
			"/v1/notes/n1/graph", ""},
	})
}

func TestCommunityReads(t *testing.T) {
	ctx := context.Background()
	runListingCalls(t, []listingCall{
		{"communities", func(c *Client) (io.ReadCloser, error) { return c.VaultCommunities(ctx, "v1", Page{Limit: 20}) },
			"/v1/vaults/v1/communities", "limit=20"},
		{"community", func(c *Client) (io.ReadCloser, error) { return c.VaultCommunity(ctx, "v1", 42) },
			"/v1/vaults/v1/communities/42", ""},
	})
}

func TestVaultGraphMapsForbidden(t *testing.T) {
	server := newFakeServer(t, replyJSON(403, `{"error":"vault_forbidden","message":"vault v9"}`))
	_, err := newTestClient(t, server, "").VaultGraph(context.Background(), "v9", GraphFilter{})
	if failure.KindOf(err) != failure.Forbidden {
		t.Errorf("VaultGraph on a 403 = %v, want forbidden", err)
	}
}

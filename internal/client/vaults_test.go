package client

import (
	"context"
	"testing"
)

func TestVaultCallsSendTheirRequests(t *testing.T) {
	checkCalls(t, []apiCall{
		{"list", func(ctx context.Context, c *Client) error { return opened(c.ListVaults(ctx)) },
			recordedRequest{method: "GET", path: "/v1/vaults"}},
		{"create", func(ctx context.Context, c *Client) error { return opened(c.CreateVault(ctx, "Work")) },
			recordedRequest{method: "POST", path: "/v1/vaults", body: `{"name":"Work"}`}},
		{"rename", func(ctx context.Context, c *Client) error { return opened(c.RenameVault(ctx, "v/1", "Home")) },
			recordedRequest{method: "PATCH", path: "/v1/vaults/v%2F1", body: `{"name":"Home"}`}},
		{"delete", func(ctx context.Context, c *Client) error { return c.DeleteVault(ctx, "v1") },
			recordedRequest{method: "DELETE", path: "/v1/vaults/v1"}},
	})
}

func TestVaultMemberCallsSendTheirRequests(t *testing.T) {
	checkCalls(t, []apiCall{
		{"list", func(ctx context.Context, c *Client) error { return opened(c.ListVaultMembers(ctx, "v1")) },
			recordedRequest{method: "GET", path: "/v1/vaults/v1/members"}},
		{"add", func(ctx context.Context, c *Client) error { return opened(c.AddVaultMember(ctx, "v1", "u1")) },
			recordedRequest{method: "POST", path: "/v1/vaults/v1/members", body: `{"user_id":"u1"}`}},
		{"remove", func(ctx context.Context, c *Client) error { return c.RemoveVaultMember(ctx, "v1", "u/1") },
			recordedRequest{method: "DELETE", path: "/v1/vaults/v1/members/u%2F1"}},
	})
}

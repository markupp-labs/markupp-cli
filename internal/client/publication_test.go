package client

import (
	"context"
	"testing"
)

func TestPublicationCallsSendTheirRequests(t *testing.T) {
	checkCalls(t, []apiCall{
		{"get", func(ctx context.Context, c *Client) error { return opened(c.GetPublication(ctx, "v1")) },
			recordedRequest{method: "GET", path: "/v1/vaults/v1/publication"}},
		{"publish", func(ctx context.Context, c *Client) error { return opened(c.PublishVault(ctx, "v1", "docs", "Docs")) },
			recordedRequest{method: "PUT", path: "/v1/vaults/v1/publication", body: `{"slug":"docs","title":"Docs"}`}},
		{"publish without title", func(ctx context.Context, c *Client) error { return opened(c.PublishVault(ctx, "v1", "docs", "")) },
			recordedRequest{method: "PUT", path: "/v1/vaults/v1/publication", body: `{"slug":"docs"}`}},
		{"unpublish", func(ctx context.Context, c *Client) error { return c.UnpublishVault(ctx, "v1") },
			recordedRequest{method: "DELETE", path: "/v1/vaults/v1/publication"}},
	})
}

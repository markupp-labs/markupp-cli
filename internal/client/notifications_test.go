package client

import (
	"context"
	"testing"
)

func TestNotificationsCallSendsItsRequest(t *testing.T) {
	checkCalls(t, []apiCall{
		{"feed", func(ctx context.Context, c *Client) error { return opened(c.Notifications(ctx)) },
			recordedRequest{method: "GET", path: "/v1/me/notifications"}},
	})
}

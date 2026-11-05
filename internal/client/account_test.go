package client

import (
	"context"
	"io"
	"testing"
)

func TestAccountReadsGoUnderV1(t *testing.T) {
	cases := []struct {
		path string
		call func(*Client, context.Context) (io.ReadCloser, error)
	}{
		{"/v1/healthz", (*Client).GetHealth},
		{"/v1/auth/me", (*Client).GetAuthenticatedUser},
		{"/v1/me/tenants", (*Client).ListMyTenants},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			server := newFakeServer(t, replyJSON(200, `{"status":"ok"}`))
			body, err := tc.call(newTestClient(t, server, ""), context.Background())
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			if got := readAll(t, body); got != `{"status":"ok"}` {
				t.Errorf("body = %q, want the payload unchanged", got)
			}
			if got := server.seen()[0]; got.method != "GET" || got.path != tc.path {
				t.Errorf("request = %+v, want GET %s", got, tc.path)
			}
		})
	}
}

func TestHealthOfAPublicClientSendsNoToken(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{"status":"ok"}`))
	c, err := NewPublic(Options{BaseURL: server.URL, Tokens: StaticToken("tok"), Tenant: "t1"})
	if err != nil {
		t.Fatalf("NewPublic: %v", err)
	}
	body, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	_ = body.Close()
	if got := server.seen()[0]; got.auth != "" || got.tenant != "" {
		t.Errorf("request = %+v, want no Authorization and no tenant", got)
	}
}

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// adminCall is one administration method of the client and the request it must send.
type adminCall struct {
	name   string
	call   func(c *Client) error
	method string
	path   string
	body   string
}

func openThenClose(open func() (io.ReadCloser, error)) error {
	body, err := open()
	if err != nil {
		return err
	}
	return body.Close()
}

func adminCalls() []adminCall {
	ctx := context.Background()
	return []adminCall{
		{"ListMyTenants", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.ListMyTenants(ctx) }) }, "GET", "/v1/me/tenants", ""},
		{"CreateTenant", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.CreateTenant(ctx, "Acme") }) }, "POST", "/v1/tenants", `{"name":"Acme"}`},
		{"ListMembers", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.ListMembers(ctx) }) }, "GET", "/v1/users", ""},
		{"ChangeMemberRole", func(c *Client) error { return c.ChangeMemberRole(ctx, "u/1", "admin") }, "PATCH", "/v1/users/u%2F1", `{"role":"admin"}`},
		{"RemoveMember", func(c *Client) error { return c.RemoveMember(ctx, "u1") }, "DELETE", "/v1/users/u1", ""},
		{"InviteMember", func(c *Client) error {
			return openThenClose(func() (io.ReadCloser, error) { return c.InviteMember(ctx, "a@b.c", "viewer") })
		}, "POST", "/v1/users/invites", `{"email":"a@b.c","role":"viewer"}`},
		{"ListTenantInvites", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.ListTenantInvites(ctx) }) }, "GET", "/v1/users/invites", ""},
		{"RevokeInvite", func(c *Client) error { return c.RevokeInvite(ctx, "i1") }, "DELETE", "/v1/users/invites/i1", ""},
		{"ListMyInvites", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.ListMyInvites(ctx) }) }, "GET", "/v1/me/invites", ""},
		{"AcceptInvite", func(c *Client) error { return openThenClose(func() (io.ReadCloser, error) { return c.AcceptInvite(ctx, "i1") }) }, "POST", "/v1/me/invites/i1/accept", ""},
		{"DeclineInvite", func(c *Client) error { return c.DeclineInvite(ctx, "i1") }, "POST", "/v1/me/invites/i1/decline", ""},
	}
}

func TestAdminMethodsSendTheirRoute(t *testing.T) {
	for _, tc := range adminCalls() {
		server := newFakeServer(t, replyJSON(200, `{}`))
		if err := tc.call(newTestClient(t, server, "t1")); err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		got := server.seen()[0]
		if got.method != tc.method || got.path != tc.path || got.body != tc.body || got.tenant != "t1" {
			t.Errorf("%s sent %+v, want %s %s with body %q and tenant t1", tc.name, got, tc.method, tc.path, tc.body)
		}
	}
}

func adminForbidden(t *testing.T, reply http.HandlerFunc) *failure.Error {
	t.Helper()
	server := newFakeServer(t, reply)
	_, err := newTestClient(t, server, "").ListMembers(context.Background())
	var failed *failure.Error
	if !errors.As(err, &failed) || failed.Kind != failure.Forbidden {
		t.Fatalf("ListMembers error = %#v, want a forbidden *failure.Error", err)
	}
	return failed
}

func TestAdminRouteWithoutTheScopeNamesIt(t *testing.T) {
	got := adminForbidden(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope", scope="administracao"`)
		replyJSON(403, `{"error":"insufficient_scope","message":"o token nao tem o escopo administracao"}`)(w, nil)
	})
	if got.Scope != "administracao" || !strings.Contains(got.Hint, "administracao") || !strings.Contains(got.Hint, "mu login") {
		t.Errorf("insufficient_scope mapped to %+v, want the scope and a login hint", got)
	}
}

func TestFirstPartyOnlyRoutePointsToThePanel(t *testing.T) {
	got := adminForbidden(t, replyJSON(403, `{"error":"first_party_only","message":"rota so aceita o token de login"}`))
	if got.Code != "first_party_only" || !strings.Contains(got.Hint, "web panel") {
		t.Errorf("first_party_only mapped to %+v, want a hint about the web panel", got)
	}
}

package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// fakeAdminServer stands for the administration routes of the server: every call is
// recorded, and answers with body, or fails with err.
type fakeAdminServer struct {
	body  string
	err   error
	calls []string
}

func (f *fakeAdminServer) answer(call string) (io.ReadCloser, error) {
	if err := f.call(call); err != nil {
		return nil, err
	}
	return io.NopCloser(strings.NewReader(f.body)), nil
}

func (f *fakeAdminServer) call(call string) error {
	f.calls = append(f.calls, call)
	return f.err
}

func (f *fakeAdminServer) ListMyTenants(context.Context) (io.ReadCloser, error) {
	return f.answer("ListMyTenants")
}

func (f *fakeAdminServer) CreateTenant(_ context.Context, name string) (io.ReadCloser, error) {
	return f.answer("CreateTenant " + name)
}

func (f *fakeAdminServer) ListMembers(context.Context) (io.ReadCloser, error) {
	return f.answer("ListMembers")
}

func (f *fakeAdminServer) ChangeMemberRole(_ context.Context, id, role string) error {
	return f.call("ChangeMemberRole " + id + " " + role)
}

func (f *fakeAdminServer) RemoveMember(_ context.Context, id string) error {
	return f.call("RemoveMember " + id)
}

func (f *fakeAdminServer) InviteMember(_ context.Context, email, role string) (io.ReadCloser, error) {
	return f.answer("InviteMember " + email + " " + role)
}

func (f *fakeAdminServer) ListTenantInvites(context.Context) (io.ReadCloser, error) {
	return f.answer("ListTenantInvites")
}

func (f *fakeAdminServer) RevokeInvite(_ context.Context, id string) error {
	return f.call("RevokeInvite " + id)
}

func (f *fakeAdminServer) ListMyInvites(context.Context) (io.ReadCloser, error) {
	return f.answer("ListMyInvites")
}

func (f *fakeAdminServer) AcceptInvite(_ context.Context, id string) (io.ReadCloser, error) {
	return f.answer("AcceptInvite " + id)
}

func (f *fakeAdminServer) DeclineInvite(_ context.Context, id string) error {
	return f.call("DeclineInvite " + id)
}

// newAdminMu returns a mu with a server URL whose administration routes are fake.
func newAdminMu(body string) (*testMu, *fakeAdminServer) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	fake := &fakeAdminServer{body: body}
	mu.server.services = Services{Tenants: fake, Members: fake, Invites: fake}
	return mu, fake
}

// expectCalls fails unless the fake saw exactly want.
func (f *fakeAdminServer) expectCalls(t *testing.T, want ...string) {
	t.Helper()
	if strings.Join(f.calls, "|") != strings.Join(want, "|") {
		t.Errorf("server calls = %q, want %q", f.calls, want)
	}
}

var errFirstPartyOnly = &failure.Error{
	Kind: failure.Forbidden, Subject: "GET /v1/users", Code: "first_party_only",
	Message: "rota so aceita o token de login", Hint: "this operation is only available in the web panel",
}

package cli

import (
	"context"
	"io"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// callLog records the calls a fake port received, one "name arg arg" string each, and
// answers them with reply or err.
type callLog struct {
	calls []string
	reply string
	err   error
}

func (l *callLog) record(name string, args ...string) {
	l.calls = append(l.calls, strings.TrimSpace(name+" "+strings.Join(args, " ")))
}

func (l *callLog) open(name string, args ...string) (io.ReadCloser, error) {
	l.record(name, args...)
	if l.err != nil {
		return nil, l.err
	}
	return body(l.reply), nil
}

func (l *callLog) exec(name string, args ...string) error {
	l.record(name, args...)
	return l.err
}

// errEndpointAbsent is what the client returns for the plain text 404 of a server
// without the route.
func errEndpointAbsent() error {
	return &failure.Error{Kind: failure.NotFound, Code: client.CodeEndpointAbsent, Subject: "GET /v1/vaults", Message: "the server has no such endpoint"}
}

// fakeVaults is the VaultAdmin port.
type fakeVaults struct{ callLog }

func (f *fakeVaults) ListVaults(context.Context) (io.ReadCloser, error) { return f.open("list") }

func (f *fakeVaults) CreateVault(_ context.Context, name string) (io.ReadCloser, error) {
	return f.open("create", name)
}

func (f *fakeVaults) RenameVault(_ context.Context, id, name string) (io.ReadCloser, error) {
	return f.open("rename", id, name)
}

func (f *fakeVaults) DeleteVault(_ context.Context, id string) error { return f.exec("delete", id) }

// fakeVaultMembers is the VaultMembership port.
type fakeVaultMembers struct{ callLog }

func (f *fakeVaultMembers) ListVaultMembers(_ context.Context, id string) (io.ReadCloser, error) {
	return f.open("list", id)
}

func (f *fakeVaultMembers) AddVaultMember(_ context.Context, id, userID string) (io.ReadCloser, error) {
	return f.open("add", id, userID)
}

func (f *fakeVaultMembers) RemoveVaultMember(_ context.Context, id, userID string) error {
	return f.exec("remove", id, userID)
}

package tui

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// apiCall is one call a fake received: the method and its main argument.
type apiCall struct {
	method string
	arg    string
}

// fakeAPI is every port of the interface. A body is keyed by method, or by method
// and argument ("graph n1") when a test needs one per argument. Calls arrive off the
// event loop, so every field is behind mu.
type fakeAPI struct {
	mu      sync.Mutex
	bodies  map[string]string
	errs    map[string][]error
	calls   []apiCall
	updates []schema.NoteUpdateRequest
}

func newFakeAPI(bodies map[string]string) *fakeAPI {
	return &fakeAPI{bodies: bodies, errs: map[string][]error{}}
}

// failNext makes the next call to method fail with err.
func (f *fakeAPI) failNext(method string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errs[method] = append(f.errs[method], err)
}

func (f *fakeAPI) answer(method, arg string) (io.ReadCloser, error) {
	if err := f.record(method, arg); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	body, ok := f.bodies[method+" "+arg]
	if !ok {
		body = f.bodies[method]
	}
	return io.NopCloser(strings.NewReader(body)), nil
}

// record keeps the call and pops the error queued for it.
func (f *fakeAPI) record(method, arg string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, apiCall{method: method, arg: arg})
	queued := f.errs[method]
	if len(queued) == 0 {
		return nil
	}
	f.errs[method] = queued[1:]
	return queued[0]
}

// callsTo returns the arguments of the calls to method.
func (f *fakeAPI) callsTo(method string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var args []string
	for _, call := range f.calls {
		if call.method == method {
			args = append(args, call.arg)
		}
	}
	return args
}

func (f *fakeAPI) sentUpdates() []schema.NoteUpdateRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]schema.NoteUpdateRequest(nil), f.updates...)
}

func (f *fakeAPI) ListNotes(_ context.Context, vault string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("list", vault)
}

func (f *fakeAPI) GetNote(_ context.Context, id string) (io.ReadCloser, error) {
	return f.answer("get", id)
}

func (f *fakeAPI) SearchNotes(_ context.Context, text, vault string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("search", vault+":"+text)
}

func (f *fakeAPI) UpdateNote(_ context.Context, id string, update schema.NoteUpdateRequest) (io.ReadCloser, error) {
	f.mu.Lock()
	f.updates = append(f.updates, update)
	f.mu.Unlock()
	return f.answer("update", id)
}

func (f *fakeAPI) DeleteNote(_ context.Context, id string) error {
	return f.record("delete", id)
}

func (f *fakeAPI) NoteLinks(_ context.Context, id, _ string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("links", id)
}

func (f *fakeAPI) NoteBacklinks(_ context.Context, id, _ string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("backlinks", id)
}

func (f *fakeAPI) NoteMentions(_ context.Context, id string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("mentions", id)
}

func (f *fakeAPI) NoteSuggestions(_ context.Context, id string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("suggestions", id)
}

func (f *fakeAPI) NoteGraph(_ context.Context, id string, _ int) (io.ReadCloser, error) {
	return f.answer("graph", id)
}

func (f *fakeAPI) VaultCommunities(_ context.Context, vault string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("communities", vault)
}

func (f *fakeAPI) VaultLinks(_ context.Context, vault string, _ client.Page) (io.ReadCloser, error) {
	return f.answer("vault-links", vault)
}

func (f *fakeAPI) GetAuthenticatedUser(context.Context) (io.ReadCloser, error) {
	return f.answer("me", "")
}

func (f *fakeAPI) ListMyTenants(context.Context) (io.ReadCloser, error) {
	return f.answer("tenants", "")
}

func (f *fakeAPI) ListVaults(context.Context) (io.ReadCloser, error) {
	return f.answer("vaults", "")
}

func (f *fakeAPI) ListVaultMembers(_ context.Context, id string) (io.ReadCloser, error) {
	return f.answer("vault-members", id)
}

func (f *fakeAPI) ListMembers(context.Context) (io.ReadCloser, error) {
	return f.answer("members", "")
}

func (f *fakeAPI) ListMyInvites(context.Context) (io.ReadCloser, error) {
	return f.answer("invites", "")
}

func (f *fakeAPI) Notifications(context.Context) (io.ReadCloser, error) {
	return f.answer("notifications", "")
}

// fakeConnector hands out the ports of one fakeAPI and records the scopes.
type fakeConnector struct {
	mu     sync.Mutex
	api    *fakeAPI
	ports  Ports
	err    error
	scopes []Scope
}

// newFakeConnector connects every port that exists on the dev server to api.
func newFakeConnector(api *fakeAPI) *fakeConnector {
	return &fakeConnector{api: api, ports: Ports{Notes: api, Relations: api, Graph: api, Account: api}}
}

// withAll also connects the ports of the open server pull requests.
func (c *fakeConnector) withAll() *fakeConnector {
	c.ports.Vaults, c.ports.Team, c.ports.Notifications = c.api, c.api, c.api
	return c
}

func (c *fakeConnector) connect(_ context.Context, scope Scope) (Ports, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scopes = append(c.scopes, scope)
	return c.ports, c.err
}

func (c *fakeConnector) connected() []Scope {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Scope(nil), c.scopes...)
}

package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/config"
)

// muRun is what one execution of mu left behind.
type muRun struct {
	stdout string
	stderr string
	code   int
}

// fakeEnv is the process environment of a test.
type fakeEnv map[string]string

func (e fakeEnv) lookup(key string) (string, bool) {
	value, ok := e[key]
	return value, ok
}

// fakeConfigStore keeps the settings in memory.
type fakeConfigStore struct {
	path     string
	settings config.Settings
	loadErr  error
	updates  int
}

func (s *fakeConfigStore) Path() (string, error) { return s.path, s.loadErr }

func (s *fakeConfigStore) Load() (config.Settings, error) { return s.settings, s.loadErr }

func (s *fakeConfigStore) Save(settings config.Settings) error {
	s.settings = settings
	return s.loadErr
}

func (s *fakeConfigStore) Lock(context.Context) (func() error, error) {
	return func() error { return nil }, s.loadErr
}

func (s *fakeConfigStore) Update(_ context.Context, mutate func(*config.Settings) error) error {
	if s.loadErr != nil {
		return s.loadErr
	}
	next := s.settings
	if err := mutate(&next); err != nil {
		return err
	}
	s.settings = next
	s.updates++
	return nil
}

// fakeServer is a ServiceFactory that records the connection and hands out services.
// privateErr fails only the connections that need a credential.
type fakeServer struct {
	services    Services
	err         error
	privateErr  error
	connections []Connection
}

func (f *fakeServer) connect(_ context.Context, conn Connection) (Services, error) {
	f.connections = append(f.connections, conn)
	if !conn.Public && f.privateErr != nil {
		return Services{}, f.privateErr
	}
	return f.services, f.err
}

// fakeAuthenticator is an AuthFactory that records the setup and the logins.
type fakeAuthenticator struct {
	setups  []AuthSetup
	servers []string
	cred    config.Credential
	removed bool
	err     error
	logouts int
}

func (f *fakeAuthenticator) build(setup AuthSetup) Authenticator {
	f.setups = append(f.setups, setup)
	return f
}

func (f *fakeAuthenticator) Login(_ context.Context, server string) (config.Credential, error) {
	f.servers = append(f.servers, server)
	return f.cred, f.err
}

func (f *fakeAuthenticator) Logout(context.Context) (bool, error) {
	f.logouts++
	return f.removed, f.err
}

// testMu is one isolated mu: its environment, config file and server.
type testMu struct {
	env    fakeEnv
	store  *fakeConfigStore
	server *fakeServer
	auth   *fakeAuthenticator
}

func newTestMu() *testMu {
	return &testMu{
		env:    fakeEnv{},
		store:  &fakeConfigStore{path: "/home/u/.config/markupp/cli.json"},
		server: &fakeServer{},
		auth:   &fakeAuthenticator{},
	}
}

func (m *testMu) deps(stdout, stderr *bytes.Buffer) Deps {
	return Deps{
		Stdout: stdout, Stderr: stderr, Env: m.env.lookup, Store: m.store,
		Version: "1.2.3", Connect: m.server.connect, Auth: m.auth.build,
	}
}

// run executes mu with args and returns its stdout, stderr and exit code.
func (m *testMu) run(t *testing.T, args ...string) muRun {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Execute(context.Background(), m.deps(&stdout, &stderr), args)
	return muRun{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func (r muRun) expectCode(t *testing.T, want int) {
	t.Helper()
	if r.code != want {
		t.Fatalf("exit code = %d, want %d\nstdout: %s\nstderr: %s", r.code, want, r.stdout, r.stderr)
	}
}

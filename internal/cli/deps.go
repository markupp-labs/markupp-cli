package cli

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/session"
)

// Deps is everything mu takes from the process. main fills it; tests fill it with
// buffers and fakes.
type Deps struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Env     config.LookupEnv
	Store   ConfigStore
	Version string
	// Connect builds the Services; nil means the real server through internal/client.
	Connect ServiceFactory
	// Auth builds the Authenticator of mu login and logout; nil means the session.
	Auth AuthFactory
	// Clock is the current time for the token expiry; nil means time.Now.
	Clock func() time.Time
	// OpenURL opens the login page in a browser, best effort; nil only prints it.
	OpenURL func(url string) error
	// Stdin is read only when a command is told to, as with --file -.
	Stdin io.Reader
	// ReadFile reads a file named by a flag; nil means os.ReadFile.
	ReadFile func(name string) ([]byte, error)
}

// Services are the ports the commands reach the server through. Each port is a small
// interface declared next to the commands that use it; the real *client.Client
// satisfies all of them. To add one: a field here and a line in newServices.
type Services struct {
	Account   AccountReader
	Notes     NoteService
	NoteLinks NoteLinkReader
	Graph     GraphReader
	Tenants   TenantAdmin
	Members   MemberAdmin
	Invites   InviteInbox
}

// newServices wires the real client into every port.
func newServices(api *client.Client) Services {
	return Services{
		Account:   api,
		Notes:     api,
		NoteLinks: api,
		Graph:     api,
		Tenants:   api,
		Members:   api,
		Invites:   api,
	}
}

// Target says which server identity a command needs.
type Target struct {
	// Vault is the --vault flag of the command, empty when it has none.
	Vault string
	// Public reads a public vault: no token and no tenant are sent.
	Public bool
}

// Connection is what a ServiceFactory needs to build Services.
type Connection struct {
	Resolved       config.Resolved
	Settings       config.Settings
	Store          ConfigStore
	AllowPlainHTTP bool
	Public         bool
	Logger         *slog.Logger
	UserAgent      string
	Clock          func() time.Time
}

// ServiceFactory builds the Services for one connection.
type ServiceFactory func(ctx context.Context, conn Connection) (Services, error)

// connectServer is the real ServiceFactory.
func connectServer(ctx context.Context, conn Connection) (Services, error) {
	api, err := newAPIClient(ctx, conn)
	if err != nil {
		return Services{}, err
	}
	return newServices(api), nil
}

func newAPIClient(ctx context.Context, conn Connection) (*client.Client, error) {
	opts := client.Options{
		BaseURL:        conn.Resolved.URL.Value,
		Tenant:         conn.Resolved.Tenant.Value,
		AllowPlainHTTP: conn.AllowPlainHTTP,
		Logger:         conn.Logger,
		UserAgent:      conn.UserAgent,
	}
	if conn.Public {
		return client.NewPublic(opts)
	}
	tokens, err := tokenSource(ctx, conn)
	if err != nil {
		return nil, err
	}
	opts.Tokens = tokens
	return client.New(opts)
}

// tokenSource picks the credential: MARKUPP_TOKEN wins and is never refreshed, then
// the stored login, refreshed under the config lock by the session.
func tokenSource(_ context.Context, conn Connection) (client.TokenSource, error) {
	server := conn.Resolved.URL.Value
	if token := conn.Resolved.Token.Value; token != "" {
		return client.StaticToken(token), nil
	}
	if auth := conn.Settings.Auth; auth != nil && auth.AccessToken != "" {
		oauth := session.NewOAuth(client.NewHTTPClient(nil, conn.Logger), conn.AllowPlainHTTP)
		return session.NewSource(conn.Store, oauth, server, conn.Clock), nil
	}
	return nil, failure.New(failure.Unauthenticated, server, "not logged in").
		WithHint("run mu login " + server + ", or set " + config.EnvToken)
}

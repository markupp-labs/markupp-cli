package cli

import (
	"context"
	"io"
	"log/slog"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
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
}

// Services are the ports the commands reach the server through. Each port is a small
// interface declared next to the commands that use it; the real *client.Client
// satisfies all of them. To add one: a field here and a line in newServices.
type Services struct {
	NoteLinks NoteLinkReader
}

// newServices wires the real client into every port.
func newServices(api *client.Client) Services {
	return Services{
		NoteLinks: api,
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

// tokenSource picks the credential: MARKUPP_TOKEN wins, then the stored login.
// The session package replaces the stored login branch with a refreshing source.
func tokenSource(_ context.Context, conn Connection) (client.TokenSource, error) {
	if token := conn.Resolved.Token.Value; token != "" {
		return client.StaticToken(token), nil
	}
	if auth := conn.Settings.Auth; auth != nil && auth.AccessToken != "" {
		return client.StaticToken(auth.AccessToken), nil
	}
	return nil, failure.New(failure.Unauthenticated, conn.Resolved.URL.Value, "not logged in").
		WithHint("run mu login, or set " + config.EnvToken)
}

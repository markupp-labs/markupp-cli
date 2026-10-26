package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/session"
)

// Authenticator logs in and out of a server. session.Authenticator satisfies it.
type Authenticator interface {
	Login(ctx context.Context, server string) (config.Credential, error)
	Logout(ctx context.Context) (removed bool, err error)
}

// AuthSetup is what an AuthFactory needs to build an Authenticator.
type AuthSetup struct {
	Store          ConfigStore
	AllowPlainHTTP bool
	Logger         *slog.Logger
	Clock          func() time.Time
	Notice         io.Writer
	OpenURL        func(url string) error
}

// AuthFactory builds the Authenticator of one execution.
type AuthFactory func(setup AuthSetup) Authenticator

const loginLong = `Log in to a Markupp server with OAuth through the browser.

mu opens the consent page of the server and waits up to 5 minutes for the browser
to come back to http://127.0.0.1:<port>/callback, on this machine. The page URL is
also printed on stderr, for when no browser opens.

The server URL and the login are stored in the config file (mode 0600); without URL
the configured server is used. The access token is refreshed on its own. ` + config.EnvToken + `,
when set, wins over the stored login.`

func newLoginCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "login [URL]",
		Short: "Log in to a server through the browser",
		Long:  loginLong,
		Example: `  mu login https://markupp.example.com
  mu login http://localhost:8080`,
		Args: upToArgs("URL"),
		RunE: app.login,
	}
}

func (a *App) login(cmd *cobra.Command, args []string) error {
	server, err := a.loginServer(args)
	if err != nil {
		return err
	}
	cred, err := a.authenticator().Login(cmd.Context(), server)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(a.deps.Stderr, "Logged in to %s; the access token expires at %s and is refreshed on its own.\n",
		cred.Issuer, cred.ExpiresAt.UTC().Format(time.RFC3339))
	a.warnTokenOverride()
	return nil
}

// loginServer is the URL argument, or the configured server without one.
func (a *App) loginServer(args []string) (string, error) {
	if len(args) == 1 {
		return config.NormalizeURL(args[0])
	}
	resolved, _, err := a.resolve("")
	if err != nil {
		return "", err
	}
	if resolved.URL.Value == "" {
		return "", errNoServer()
	}
	return resolved.URL.Value, nil
}

func (a *App) warnTokenOverride() {
	if token, _ := a.deps.Env(config.EnvToken); token != "" {
		_, _ = fmt.Fprintf(a.deps.Stderr, "Warning: %s is set and wins over this login until it is unset.\n", config.EnvToken)
	}
}

// authenticator builds the Authenticator of this execution, the session by default.
func (a *App) authenticator() Authenticator {
	factory := a.deps.Auth
	if factory == nil {
		factory = newSessionAuthenticator
	}
	return factory(AuthSetup{
		Store: a.deps.Store, AllowPlainHTTP: a.global.allowPlainHTTP, Logger: a.requestLogger(),
		Clock: a.deps.Clock, Notice: a.deps.Stderr, OpenURL: a.deps.OpenURL,
	})
}

func newSessionAuthenticator(setup AuthSetup) Authenticator {
	return session.Authenticator{
		OAuth:  session.NewOAuth(client.NewHTTPClient(nil, setup.Logger), setup.AllowPlainHTTP),
		Store:  setup.Store,
		Now:    setup.Clock,
		Notice: setup.Notice,
		Open:   setup.OpenURL,
	}
}

// upToArgs accepts at most the named positional arguments, and names them and what
// was given otherwise.
func upToArgs(names ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) <= len(names) {
			return nil
		}
		message := fmt.Sprintf("expected at most %d arguments (%v), got %d: %q", len(names), names, len(args), args)
		return failure.New(failure.Usage, cmd.CommandPath(), message)
	}
}

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

// AccountReader reads who and where mu is. *client.Client satisfies it.
type AccountReader interface {
	GetHealth(ctx context.Context) (io.ReadCloser, error)
	GetAuthenticatedUser(ctx context.Context) (io.ReadCloser, error)
	ListMyTenants(ctx context.Context) (io.ReadCloser, error)
}

// Sources of the credential shown by mu status.
const (
	credentialFromEnv   = "env"
	credentialFromLogin = "login"
	credentialNone      = "none"
)

// statusView is what mu status prints. health, account and tenants are the API
// payloads unchanged, null when they could not be read.
type statusView struct {
	Server  string          `json:"server"`
	Health  json.RawMessage `json:"health"`
	Token   credentialView  `json:"token"`
	Account json.RawMessage `json:"account"`
	Tenants json.RawMessage `json:"tenants"`
}

// credentialView says where the token comes from and when it expires, never the token.
type credentialView struct {
	Source    string     `json:"source"`
	Issuer    string     `json:"issuer,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

const statusLong = `Show the server, its health, where the token comes from and when it expires, the
account and its tenants. The token itself is never shown.

The health probe needs no login. Without a valid login mu prints what it could read
and exits 3. With --json the output is an object whose health, account and tenants
are the payloads of GET /healthz, /auth/me and /me/tenants, unchanged.`

func newStatusCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the server, the login and the account; tokens are never shown",
		Long:    statusLong,
		Example: `  mu status
  mu whoami --json | jq -r .account.email`,
		Args: exactArgs(),
		RunE: app.showStatus,
	}
}

// showStatus prints what it could read, then returns the error that stopped it.
func (a *App) showStatus(cmd *cobra.Command, _ []string) error {
	view, err := a.statusOf(cmd.Context())
	if view.Server == "" {
		return err
	}
	if printErr := a.printStatus(view); printErr != nil {
		return printErr
	}
	return err
}

func (a *App) statusOf(ctx context.Context) (statusView, error) {
	resolved, settings, err := a.resolve("")
	if err != nil {
		return statusView{}, err
	}
	if resolved.URL.Value == "" {
		return statusView{}, errNoServer()
	}
	view := statusView{Server: resolved.URL.Value, Token: credentialOf(resolved, settings)}
	if view.Health, err = a.readHealth(ctx); err != nil {
		return view, err
	}
	return view, a.readAccount(ctx, &view)
}

func credentialOf(resolved config.Resolved, settings config.Settings) credentialView {
	if resolved.Token.Value != "" {
		return credentialView{Source: credentialFromEnv}
	}
	if auth := settings.Auth; auth != nil {
		expires := auth.ExpiresAt.UTC()
		return credentialView{Source: credentialFromLogin, Issuer: auth.Issuer, ExpiresAt: &expires}
	}
	return credentialView{Source: credentialNone}
}

// readHealth reads the probe without credentials, so it works before any login.
func (a *App) readHealth(ctx context.Context) (json.RawMessage, error) {
	svc, err := a.connect(ctx, Target{Public: true})
	if err != nil {
		return nil, err
	}
	return readPayload(svc.Account.GetHealth(ctx))
}

func (a *App) readAccount(ctx context.Context, view *statusView) error {
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return err
	}
	if view.Account, err = readPayload(svc.Account.GetAuthenticatedUser(ctx)); err != nil {
		return err
	}
	view.Tenants, err = readPayload(svc.Account.ListMyTenants(ctx))
	return err
}

// readPayload reads a small body whole; the three reads of status are bounded.
func readPayload(body io.ReadCloser, err error) (json.RawMessage, error) {
	if err != nil {
		return nil, err
	}
	return readObject(body)
}

func (a *App) printStatus(view statusView) error {
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	if r.mode != output.JSON {
		return printStatusText(r.printer, view)
	}
	raw, err := json.Marshal(view)
	if err != nil {
		return fmt.Errorf("encoding the status of %s: %w", view.Server, err)
	}
	return r.printer.JSON(bytes.NewReader(raw))
}

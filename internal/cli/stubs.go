// Package cli holds the cobra command tree of mu.
package cli

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

type ConfigStore interface {
	Path() (string, error)
	Load() (config.Settings, error)
	Update(ctx context.Context, mutate func(*config.Settings) error) error
}

type Services struct{}

type Target struct{ Public bool }

type Connection struct {
	Resolved       config.Resolved
	Settings       config.Settings
	Store          ConfigStore
	AllowPlainHTTP bool
	Public         bool
	Logger         *slog.Logger
	UserAgent      string
}

type ServiceFactory func(ctx context.Context, conn Connection) (Services, error)

type Deps struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Env     config.LookupEnv
	Store   ConfigStore
	Version string
	Connect ServiceFactory
}

type App struct {
	deps   Deps
	global globalFlags
}

type globalFlags struct {
	url, tenant                   string
	json, verbose, allowPlainHTTP bool
}

type renderer struct {
	mode    output.Mode
	printer *output.Printer
}

type rowFunc func(raw json.RawMessage) ([]string, error)
type objectFunc func(p *output.Printer, raw json.RawMessage) error

type listFlags struct {
	ndjson bool
	limit  int
	offset int
}

type vaultFlags struct {
	vault  string
	public string
}

type vaultTarget struct {
	ID         string
	PublicSlug string
}

func Execute(ctx context.Context, deps Deps, args []string) int        { return 0 }
func newApp(deps Deps) *App                                            { return &App{deps: deps} }
func (a *App) connect(ctx context.Context, t Target) (Services, error) { return Services{}, nil }
func (a *App) renderer(ndjson bool) (renderer, error)                  { return renderer{}, nil }
func (a *App) resolveVault(f vaultFlags) (vaultTarget, error)          { return vaultTarget{}, nil }
func (r renderer) list(body io.ReadCloser, columns []output.Column, row rowFunc) error {
	return nil
}
func (r renderer) object(body io.ReadCloser, render objectFunc) error { return nil }
func (f listFlags) page() (client.Page, error)                        { return client.Page{}, nil }
func addListFlags(cmd *cobra.Command, f *listFlags)                   {}
func addVaultFlags(cmd *cobra.Command, f *vaultFlags)                 {}
func requireYes(yes bool, action, subject string) error               { return nil }
func connectServer(ctx context.Context, conn Connection) (Services, error) {
	return Services{}, nil
}
func newLogger(verbose bool, stderr io.Writer) *slog.Logger { return nil }

func newRoot(app *App) *cobra.Command                      { return &cobra.Command{Use: "mu"} }
func malformedRecord(raw json.RawMessage, err error) error { return err }

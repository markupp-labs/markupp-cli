package cli

import (
	"context"
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

// App is the state shared by every command of one execution: the injected
// dependencies and the global flags.
type App struct {
	deps   Deps
	global globalFlags
	logger *slog.Logger
	ran    bool
}

// globalFlags are the persistent flags of the root command.
type globalFlags struct {
	url            string
	tenant         string
	json           bool
	verbose        bool
	allowPlainHTTP bool
}

func newApp(deps Deps) *App {
	return &App{deps: deps, logger: newLogger(false, deps.Stderr)}
}

func (a *App) addGlobalFlags(root *cobra.Command) {
	flags := root.PersistentFlags()
	flags.StringVar(&a.global.url, "url", "", "server URL (overrides "+config.EnvURL+" and the config file)")
	flags.StringVar(&a.global.tenant, "tenant", "", "tenant id, for accounts in several tenants (overrides "+config.EnvTenant+")")
	flags.BoolVar(&a.global.json, "json", false, "print the API payload as JSON, unchanged")
	flags.BoolVarP(&a.global.verbose, "verbose", "v", false, "log each request as JSON on stderr")
	flags.BoolVar(&a.global.allowPlainHTTP, "allow-plain-http", false, "allow http:// outside localhost, sending the token unencrypted")
}

// prepare runs after cobra parsed the flags and arguments, before any command.
func (a *App) prepare(*cobra.Command, []string) error {
	a.ran = true
	a.logger = newLogger(a.global.verbose, a.deps.Stderr)
	return nil
}

// newLogger returns a JSON logger on stderr at debug level when verbose, and a logger
// that drops everything otherwise.
func newLogger(verbose bool, stderr io.Writer) *slog.Logger {
	if !verbose {
		return slog.New(slog.DiscardHandler)
	}
	return slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func (a *App) printer() *output.Printer { return output.NewPrinter(a.deps.Stdout) }

// resolve loads the config file and applies flag > environment > file. vault is the
// --vault flag of the command, empty when it has none.
func (a *App) resolve(vault string) (config.Resolved, config.Settings, error) {
	settings, err := a.deps.Store.Load()
	if err != nil {
		return config.Resolved{}, config.Settings{}, err
	}
	overrides := config.Overrides{URL: a.global.url, Tenant: a.global.tenant, Vault: vault}
	return config.Resolve(overrides, a.deps.Env, settings), settings, nil
}

// connect resolves the settings and builds the Services of target. Without a server
// URL it explains how to configure one, which is what a first run sees.
func (a *App) connect(ctx context.Context, target Target) (Services, error) {
	resolved, settings, err := a.resolve(target.Vault)
	if err != nil {
		return Services{}, err
	}
	if resolved.URL.Value == "" {
		return Services{}, errNoServer()
	}
	return a.factory()(ctx, Connection{
		Resolved: resolved, Settings: settings, Store: a.deps.Store,
		AllowPlainHTTP: a.global.allowPlainHTTP, Public: target.Public,
		Logger: a.requestLogger(), UserAgent: "mu/" + a.deps.Version, Clock: a.deps.Clock,
	})
}

func (a *App) factory() ServiceFactory {
	if a.deps.Connect != nil {
		return a.deps.Connect
	}
	return connectServer
}

// requestLogger is the logger handed to the client, nil unless -v was given so the
// client skips the logging transport.
func (a *App) requestLogger() *slog.Logger {
	if !a.global.verbose {
		return nil
	}
	return a.logger
}

func errNoServer() error {
	return failure.New(failure.Usage, "", "no Markupp server configured").
		WithHint("run mu login https://markupp.example.com or mu config set url https://markupp.example.com, set " + config.EnvURL + ", or pass --url")
}

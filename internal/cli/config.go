package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

const configLong = `Show and change the mu configuration file.

Each setting resolves as flag > environment > file:
  url     --url      ` + config.EnvURL + `
  tenant  --tenant   ` + config.EnvTenant + `
  vault   --vault    ` + config.EnvVault + `
` + config.EnvToken + ` holds a raw access token that wins over the stored login and is
never refreshed. The file is $XDG_CONFIG_HOME/markupp/cli.json (or
~/.config/markupp/cli.json), mode 0600.`

func newConfigCmd(app *App) *cobra.Command {
	cmd := newGroup("config", "Show and change the configuration", "  mu config show\n  mu config set url https://markupp.example.com")
	cmd.Long = configLong
	cmd.AddCommand(newConfigShowCmd(app), newConfigSetCmd(app), newConfigUnsetCmd(app), newConfigPathCmd(app))
	return cmd
}

func newConfigShowCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show each setting and its source; tokens are never shown",
		Example: `  mu config show
  mu config show --json | jq -r .url.value`,
		Args: exactArgs(),
		RunE: func(*cobra.Command, []string) error { return app.showConfig() },
	}
}

func newConfigSetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "set KEY VALUE",
		Short: "Store a setting: url, tenant or vault",
		Example: `  mu config set url https://markupp.example.com
  mu config set vault 3b241101-e2bb-4255-8caf-4136c566a962`,
		Args: exactArgs("KEY", "VALUE"),
		RunE: app.setConfig,
	}
}

func newConfigUnsetCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "unset KEY",
		Short:   "Remove a setting: url, tenant or vault",
		Example: "  mu config unset tenant",
		Args:    exactArgs("KEY"),
		RunE:    app.unsetConfig,
	}
}

func newConfigPathCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "path",
		Short:   "Print the location of the configuration file",
		Example: "  mu config path",
		Args:    exactArgs(),
		RunE:    app.printConfigPath,
	}
}

func (a *App) setConfig(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	set := func(s *config.Settings) error { return s.Set(key, value) }
	return a.updateConfig(cmd.Context(), "saved", key, set)
}

func (a *App) unsetConfig(cmd *cobra.Command, args []string) error {
	key := args[0]
	unset := func(s *config.Settings) error { return s.Unset(key) }
	return a.updateConfig(cmd.Context(), "removed", key, unset)
}

func (a *App) printConfigPath(*cobra.Command, []string) error {
	path, err := a.deps.Store.Path()
	if err != nil {
		return err
	}
	return a.printer().Line(path)
}

// updateConfig applies mutate under the config lock and confirms on stderr.
func (a *App) updateConfig(ctx context.Context, verb, key string, mutate func(*config.Settings) error) error {
	if err := a.deps.Store.Update(ctx, mutate); err != nil {
		return err
	}
	path, _ := a.deps.Store.Path()
	_, _ = fmt.Fprintf(a.deps.Stderr, "%s %s in %s\n", key, verb, path)
	return nil
}

// configView is what mu config show prints: values and sources, never a token.
type configView struct {
	Path   string       `json:"path"`
	URL    sourcedValue `json:"url"`
	Tenant sourcedValue `json:"tenant"`
	Vault  sourcedValue `json:"vault"`
	Token  tokenState   `json:"token"`
	Login  *loginState  `json:"login"`
}

type sourcedValue struct {
	Value  string        `json:"value"`
	Source config.Source `json:"source"`
}

type tokenState struct {
	Set    bool          `json:"set"`
	Source config.Source `json:"source"`
}

type loginState struct {
	Issuer    string    `json:"issuer"`
	ExpiresAt time.Time `json:"expires_at"`
	Scope     string    `json:"scope,omitempty"`
}

func (a *App) showConfig() error {
	view, err := a.configView()
	if err != nil {
		return err
	}
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	if r.mode == output.JSON {
		return printConfigJSON(r.printer, view)
	}
	return printConfigText(r.printer, view)
}

func (a *App) configView() (configView, error) {
	path, err := a.deps.Store.Path()
	if err != nil {
		return configView{}, err
	}
	resolved, settings, err := a.resolve("")
	if err != nil {
		return configView{}, err
	}
	return configView{
		Path:   path,
		URL:    sourcedValue(resolved.URL),
		Tenant: sourcedValue(resolved.Tenant),
		Vault:  sourcedValue(resolved.Vault),
		Token:  tokenState{Set: resolved.Token.Value != "", Source: resolved.Token.Source},
		Login:  loginOf(settings.Auth),
	}, nil
}

func loginOf(auth *config.Credential) *loginState {
	if auth == nil {
		return nil
	}
	return &loginState{Issuer: auth.Issuer, ExpiresAt: auth.ExpiresAt.UTC(), Scope: auth.Scope}
}

func printConfigJSON(p *output.Printer, view configView) error {
	raw, err := json.Marshal(view)
	if err != nil {
		return fmt.Errorf("encoding the config view: %w", err)
	}
	return p.JSON(bytes.NewReader(raw))
}

func printConfigText(p *output.Printer, view configView) error {
	table := p.Table(output.Column{Name: "KEY", Width: 6}, output.Column{Name: "SOURCE", Width: 6}, output.Column{Name: "VALUE"})
	rows := [][]string{
		{"url", string(view.URL.Source), view.URL.Value},
		{"tenant", string(view.Tenant.Source), view.Tenant.Value},
		{"vault", string(view.Vault.Source), view.Vault.Value},
		tokenRow(view.Token),
		loginRow(view.Login),
		{"path", "", view.Path},
	}
	if err := table.Header(); err != nil {
		return err
	}
	for _, row := range rows {
		if err := table.Row(row...); err != nil {
			return err
		}
	}
	return nil
}

func tokenRow(token tokenState) []string {
	if !token.Set {
		return []string{"token", "", ""}
	}
	return []string{"token", string(token.Source), "set"}
}

func loginRow(login *loginState) []string {
	if login == nil {
		return []string{"login", "", ""}
	}
	text := fmt.Sprintf("issuer %s, expires %s", login.Issuer, login.ExpiresAt.Format(time.RFC3339))
	return []string{"login", string(config.FromFile), text}
}

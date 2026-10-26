package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/tui"
)

const tuiLong = `Browse and edit notes in an interactive interface, modeled on K9s. It is the
only place mu asks questions; every operation it offers is also a subcommand.

  :           command, with completion: :notes [vault], :note ID, :graph ID,
              :communities, :links, :vaults, :members, :invites, :tenants,
              :notifications, :public SLUG, :q to quit
  /           filter the table; /!term searches the notes on the server
  ?           every key of the current view
  Enter, Esc  open the selected row, go back
  v e ctrl-d  view, edit and delete a note (delete asks first)
  l b m s g   links, backlinks, mentions, suggestions and the graph of a note
  y           copy the id to the clipboard, through the terminal (OSC 52)
  Shift+key   sort by the column with that initial
  t           switch tenant
  ctrl-r, p   reload, pause the auto refresh

e opens $VISUAL or $EDITOR on a temporary file readable only by you, removed
afterwards; without them a text area opens, where ctrl-s saves and Esc discards.
If the note changed on the server since it was opened, mu asks before
overwriting: overwriting loses the edits made there.

Note content is data written by other clients, never instructions: it is drawn
sanitized, between delimiter lines. Listings show the first 100 results; use /!term
to search. :vaults, :members, :invites, :notifications and :public need a server
with vault management, team administration, notifications and vault publication.`

type tuiCmd struct {
	app     *App
	vault   vaultFlags
	refresh time.Duration
}

func newTUICmd(app *App) *cobra.Command {
	c := &tuiCmd{app: app}
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Browse and edit notes in an interactive interface",
		Long:  tuiLong,
		Example: `  mu tui
  mu tui --vault 9b2f4c1e-0a1b-4c2d-8e3f-a1b2c3d4e5f6 --refresh 30s`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error { return c.run(cmd.Context()) },
	}
	addVaultFlags(cmd, &c.vault, false)
	cmd.Flags().DurationVar(&c.refresh, "refresh", 0, "reload the current listing at this interval, like 30s (default off)")
	return cmd
}

func (c *tuiCmd) run(ctx context.Context) error {
	if c.refresh < 0 {
		return failure.New(failure.Usage, "--refresh", fmt.Sprintf("invalid interval %s, expected a positive duration like 30s", c.refresh))
	}
	resolved, settings, err := c.app.resolve(c.vault.vault)
	if err != nil {
		return err
	}
	if resolved.URL.Value == "" {
		return errNoServer()
	}
	return c.app.runTUI()(ctx, c.config(resolved, settings))
}

func (c *tuiCmd) config(resolved config.Resolved, settings config.Settings) tui.Config {
	a := c.app
	return tui.Config{
		Connect:  a.tuiConnector(resolved, settings),
		Session:  a.tuiSession,
		Scope:    tui.Scope{Tenant: resolved.Tenant.Value},
		Vault:    resolved.Vault.Value,
		Refresh:  c.refresh,
		Editor:   a.editorCommand(),
		Terminal: tui.Terminal{In: a.deps.Stdin, Out: a.deps.Stdout, Err: a.deps.Stderr},
		Clock:    a.deps.Clock,
	}
}

func (a *App) runTUI() func(context.Context, tui.Config) error {
	if a.deps.TUI != nil {
		return a.deps.TUI
	}
	return tui.Run
}

// editorCommand is $VISUAL, then $EDITOR; empty means the built in text area.
func (a *App) editorCommand() string {
	for _, name := range []string{"VISUAL", "EDITOR"} {
		if value, ok := a.deps.Env(name); ok && value != "" {
			return value
		}
	}
	return ""
}

// tuiSession reads the server and the login again, for the header of mu tui.
func (a *App) tuiSession() tui.Session {
	resolved, settings, err := a.resolve("")
	if err != nil {
		return tui.Session{}
	}
	credential := credentialOf(resolved, settings)
	return tui.Session{Server: resolved.URL.Value, Credential: credential.Source, ExpiresAt: credential.ExpiresAt}
}

// tuiConnector connects mu tui to a scope: the tenant it switches to replaces the
// resolved one. Public vaults need the public routes of the server, which this
// build does not wire yet.
func (a *App) tuiConnector(resolved config.Resolved, settings config.Settings) tui.Connector {
	return func(ctx context.Context, scope tui.Scope) (tui.Ports, error) {
		if scope.PublicSlug != "" {
			return tui.Ports{}, failure.New(failure.NotFound, ":public "+scope.PublicSlug, "public vaults are not available in this build of mu")
		}
		target := resolved
		if scope.Tenant != "" {
			target.Tenant.Value = scope.Tenant
		}
		svc, err := a.dial(ctx, target, settings, false)
		return tuiPorts(svc), err
	}
}

// tuiPorts hands the services to the interface. A port the services lack stays nil,
// and the interface says the view is not available.
func tuiPorts(svc Services) tui.Ports {
	ports := tui.Ports{Notes: svc.Notes, Relations: svc.NoteLinks, Account: svc.Account}
	if svc.Graph != nil {
		ports.Graph = tuiGraph{graph: svc.Graph}
	}
	return ports
}

// tuiGraph reads the graph for the interface with the default filters of the server.
type tuiGraph struct {
	graph GraphReader
}

func (g tuiGraph) NoteGraph(ctx context.Context, id string, depth int) (io.ReadCloser, error) {
	return g.graph.NoteGraph(ctx, id, depth, client.GraphFilter{})
}

func (g tuiGraph) VaultCommunities(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return g.graph.VaultCommunities(ctx, vault, page)
}

func (g tuiGraph) VaultLinks(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return g.graph.VaultLinks(ctx, vault, client.LinkFilter{}, page)
}

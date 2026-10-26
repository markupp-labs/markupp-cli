package cli

import (
	"context"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
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

func newTUICmd(app *App) *cobra.Command { return *new(*cobra.Command) }

func (c *tuiCmd) run(ctx context.Context) error { return *new(error) }

func (c *tuiCmd) config(resolved config.Resolved, settings config.Settings) tui.Config {
	return *new(tui.Config)
}

func (a *App) runTUI() func(context.Context, tui.Config) error {
	return *new(func(context.Context, tui.Config) error)
}

// editorCommand is $VISUAL, then $EDITOR; empty means the built in text area.
func (a *App) editorCommand() string { return *new(string) }

// tuiSession reads the server and the login again, for the header of mu tui.
func (a *App) tuiSession() tui.Session { return *new(tui.Session) }

// tuiConnector connects mu tui to a scope: the tenant it switches to replaces the
// resolved one. Public vaults need the public routes of the server, which this
// build does not wire yet.
func (a *App) tuiConnector(resolved config.Resolved, settings config.Settings) tui.Connector {
	return *new(tui.Connector)
}

// tuiPorts hands the services to the interface. A port the services lack stays nil,
// and the interface says the view is not available.
func tuiPorts(svc Services) tui.Ports { return *new(tui.Ports) }

// tuiGraph reads the graph for the interface with the default filters of the server.
type tuiGraph struct {
	graph GraphReader
}

func (g tuiGraph) NoteGraph(ctx context.Context, id string, depth int) (io.ReadCloser, error) {
	return *new(io.ReadCloser), *new(error)
}

func (g tuiGraph) VaultCommunities(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return *new(io.ReadCloser), *new(error)
}

func (g tuiGraph) VaultLinks(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error) {
	return *new(io.ReadCloser), *new(error)
}

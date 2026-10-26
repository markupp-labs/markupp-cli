package tui

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// shell is the frame around the views: header, prompt, main frame, crumbs and flash.
// It knows no resource; the views and commandTable do.
type shell struct {
	cfg    Config
	ctx    context.Context
	app    *tview.Application
	pages  *tview.Pages
	layout *tview.Flex
	body   *tview.Flex
	header *header
	prompt *prompt
	crumbs *tview.TextView
	flash  *flash
	stack  stack
	ports  Ports
	scope  Scope
	// vault is the vault :notes opens without an argument.
	vault string
	// tenantName is the name of the tenant once known, for the header and crumbs.
	tenantName string
	account    string
	session    Session
	paused     bool
	// overlay is true while the prompt or a dialog has the keyboard.
	overlay   bool
	runEditor func(argv []string) error
}

// Run starts mu tui and blocks until the user quits or ctx ends. It connects before
// drawing anything, so a missing login fails like any other command.
func Run(ctx context.Context, cfg Config) error { return *new(error) }

func newShell(ctx context.Context, cfg Config) (*shell, error) { return *new(*shell), *new(error) }

// build lays the widgets out, top to bottom.
func (s *shell) build() {}

const mainPage = "main"

func (s *shell) run() error { return *new(error) }

func (s *shell) stopOnDone() {}

// async runs work off the event loop and applies its result on it. Views change only
// on the event loop, so they need no locks.
func (s *shell) async(work func(ctx context.Context) (func(), error)) {}

func (s *shell) settle(apply func(), err error) {}

// push shows v on top of the stack and loads it.
func (s *shell) push(v View) {}

// reset replaces the stack with v, as a : command does.
func (s *shell) reset(v View) {}

// back pops the top view and shows the one below, as it was.
func (s *shell) back() {}

// reload reads the top view again.
func (s *shell) reload() {}

func (s *shell) show(v View) {}

func (s *shell) focusTop() {}

// refreshChrome redraws everything around the view: title, header and crumbs.
func (s *shell) refreshChrome() {}

// capture routes a key: the prompt and dialogs keep the keyboard while open, then
// the keys of a typing view, then the global keys, then the keys of the view.
func (s *shell) capture(event *tcell.EventKey) *tcell.EventKey { return *new(*tcell.EventKey) }

func (s *shell) typingKey(event *tcell.EventKey) *tcell.EventKey { return *new(*tcell.EventKey) }

func (s *shell) viewKey(event *tcell.EventKey) *tcell.EventKey { return *new(*tcell.EventKey) }

// Typing is a view where printable keys type text instead of running hotkeys.
type Typing interface {
	Typing() bool
}

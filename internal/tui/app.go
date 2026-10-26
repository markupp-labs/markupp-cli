package tui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/failure"
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
func Run(ctx context.Context, cfg Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s, err := newShell(ctx, cfg)
	if err != nil {
		return err
	}
	return s.run()
}

func newShell(ctx context.Context, cfg Config) (*shell, error) {
	if cfg.Connect == nil {
		return nil, failure.New(failure.Internal, "", "mu tui needs a connector, got none")
	}
	ports, err := cfg.Connect(ctx, cfg.Scope)
	if err != nil {
		return nil, err
	}
	s := &shell{cfg: cfg, ctx: ctx, app: tview.NewApplication(), ports: ports, scope: cfg.Scope, vault: cfg.Vault}
	s.runEditor = s.execEditor
	s.build()
	return s, nil
}

// build lays the widgets out, top to bottom.
func (s *shell) build() {
	s.header = newHeader()
	s.prompt = newPrompt(s)
	s.body = tview.NewFlex()
	s.body.SetBorder(true)
	s.crumbs = tview.NewTextView().SetDynamicColors(true)
	s.flash = newFlash(s)
	s.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(s.header.widget, headerHeight, 0, false).
		AddItem(s.prompt.field, 0, 0, false).
		AddItem(s.body, 0, 1, true).
		AddItem(s.crumbs, 1, 0, false).
		AddItem(s.flash.view, 1, 0, false)
	s.pages = tview.NewPages().AddPage(mainPage, s.layout, true, true)
	s.app.SetRoot(s.pages, true).SetInputCapture(s.capture)
}

const mainPage = "main"

func (s *shell) run() error {
	if s.cfg.Screen != nil {
		s.app.SetScreen(s.cfg.Screen)
	}
	s.session = s.cfg.session()
	s.reset(s.homeView())
	s.loadAccount()
	go s.stopOnDone()
	s.startRefresh()
	if err := s.app.Run(); err != nil {
		return failure.Wrap(failure.Usage, "mu tui", fmt.Errorf("starting the terminal interface, expected a terminal: %w", err))
	}
	return nil
}

func (s *shell) stopOnDone() {
	<-s.ctx.Done()
	s.app.Stop()
}

// async runs work off the event loop and applies its result on it. Views change only
// on the event loop, so they need no locks.
func (s *shell) async(work func(ctx context.Context) (func(), error)) {
	go func() {
		apply, err := work(s.ctx)
		s.app.QueueUpdateDraw(func() { s.settle(apply, err) })
	}()
}

func (s *shell) settle(apply func(), err error) {
	if err != nil {
		s.flashError(err)
		return
	}
	apply()
}

// push shows v on top of the stack and loads it.
func (s *shell) push(v View) {
	s.stack.push(v)
	s.show(v)
	v.Load()
}

// reset replaces the stack with v, as a : command does.
func (s *shell) reset(v View) {
	s.stack.reset(v)
	s.show(v)
	v.Load()
}

// back pops the top view and shows the one below, as it was.
func (s *shell) back() {
	if s.stack.pop() {
		s.show(s.stack.top())
	}
}

// reload reads the top view again.
func (s *shell) reload() {
	if top := s.stack.top(); top != nil {
		top.Load()
	}
}

func (s *shell) show(v View) {
	s.body.Clear()
	s.body.AddItem(v.Widget(), 0, 1, true)
	s.focusTop()
	s.refreshChrome()
}

func (s *shell) focusTop() {
	if top := s.stack.top(); top != nil {
		s.app.SetFocus(top.Widget())
	}
}

// refreshChrome redraws everything around the view: title, header and crumbs.
func (s *shell) refreshChrome() {
	top := s.stack.top()
	if top == nil {
		return
	}
	s.body.SetTitle(" " + plain(top.Title()) + " ")
	s.header.render(s.headerState(top))
	s.crumbs.SetText(crumbText(s.scopeLabel(), s.stack.crumbs()))
}

// capture routes a key: the prompt and dialogs keep the keyboard while open, then
// the keys of a typing view, then the global keys, then the keys of the view.
func (s *shell) capture(event *tcell.EventKey) *tcell.EventKey {
	if s.overlay || s.stack.top() == nil {
		return event
	}
	if typing, ok := s.stack.top().(Typing); ok && typing.Typing() {
		return s.typingKey(event)
	}
	if s.globalKey(event) {
		return nil
	}
	return s.viewKey(event)
}

func (s *shell) typingKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		return event
	}
	return s.viewKey(event)
}

func (s *shell) viewKey(event *tcell.EventKey) *tcell.EventKey {
	key, ok := findHotkey(s.stack.top().Hotkeys(), event)
	if !ok {
		return event
	}
	key.Run()
	return nil
}

// Typing is a view where printable keys type text instead of running hotkeys.
type Typing interface {
	Typing() bool
}

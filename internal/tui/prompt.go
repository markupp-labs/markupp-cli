package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// promptKind is what the prompt line is for: a : command or a / filter.
type promptKind rune

const (
	promptCommand promptKind = ':'
	promptFilter  promptKind = '/'
)

// searchPrefix turns a / filter into a server search: /!term.
const searchPrefix = "!"

// prompt is the line that opens under the header for : and /.
type prompt struct {
	shell *shell
	field *tview.InputField
	kind  promptKind
}

func newPrompt(s *shell) *prompt {
	p := &prompt{shell: s, field: tview.NewInputField()}
	p.field.SetFieldBackgroundColor(tcell.ColorDefault)
	p.field.SetAutocompleteFunc(p.complete)
	p.field.SetAutocompletedFunc(p.completed)
	p.field.SetDoneFunc(p.done)
	return p
}

// completed takes the chosen command; with Enter it also runs it, as K9s does.
func (p *prompt) completed(text string, _, source int) bool {
	if source == tview.AutocompletedNavigate {
		return false
	}
	p.field.SetText(text)
	if source == tview.AutocompletedEnter {
		p.done(tcell.KeyEnter)
	}
	return true
}

// open shows the prompt and gives it the keyboard.
func (p *prompt) open(kind promptKind) {
	p.kind = kind
	p.field.SetLabel(string(rune(kind))).SetText("")
	p.shell.layout.ResizeItem(p.field, 1, 0)
	p.shell.overlay = true
	p.shell.app.SetFocus(p.field)
}

func (p *prompt) close() {
	p.shell.layout.ResizeItem(p.field, 0, 0)
	p.shell.overlay = false
	p.shell.focusTop()
}

func (p *prompt) done(key tcell.Key) {
	line := strings.TrimSpace(p.field.GetText())
	p.close()
	if key != tcell.KeyEnter {
		return
	}
	if p.kind == promptCommand {
		p.shell.execute(line)
		return
	}
	p.shell.applyFilter(line)
}

// complete offers the commands and aliases that start with the typed word.
func (p *prompt) complete(text string) []string {
	if p.kind != promptCommand || text == "" || strings.Contains(text, " ") {
		return nil
	}
	return completions(text)
}

// execute runs a : command line.
func (s *shell) execute(line string) {
	word, arg := parseCommand(line)
	if word == "" {
		return
	}
	if isQuit(word) {
		s.app.Stop()
		return
	}
	cmd, ok := findCommand(word)
	if !ok {
		s.flashError(fmt.Errorf("unknown command %q, expected one of: %s", word, strings.Join(commandNames(), ", ")))
		return
	}
	if err := cmd.open(s, arg); err != nil {
		s.flashError(err)
	}
}

// applyFilter filters the view locally, or searches the server with /!term.
func (s *shell) applyFilter(line string) {
	if term, ok := strings.CutPrefix(line, searchPrefix); ok {
		s.search(strings.TrimSpace(term))
		return
	}
	f, ok := s.stack.top().(Filterable)
	if !ok {
		s.flashError(fmt.Errorf("%s cannot be filtered, expected a listing", s.stack.top().Crumb()))
		return
	}
	f.SetFilter(line)
}

func (s *shell) search(term string) {
	searchable, ok := s.stack.top().(Searchable)
	if !ok || term == "" {
		s.flashError(fmt.Errorf("search %q needs a listing of notes and a term, like /!meeting", term))
		return
	}
	searchable.Search(term)
}

// parseCommand splits a command line into its word and the rest.
func parseCommand(line string) (word, arg string) {
	word, arg, _ = strings.Cut(strings.TrimSpace(line), " ")
	return strings.ToLower(word), strings.TrimSpace(arg)
}

func isQuit(word string) bool {
	return word == "q" || word == "quit" || word == "q!"
}

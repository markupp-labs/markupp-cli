package tui

import (
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

func newPrompt(s *shell) *prompt { return *new(*prompt) }

// completed takes the chosen command; with Enter it also runs it, as K9s does.
func (p *prompt) completed(text string, _, source int) bool { return *new(bool) }

// open shows the prompt and gives it the keyboard.
func (p *prompt) open(kind promptKind) {}

func (p *prompt) close() {}

func (p *prompt) done(key tcell.Key) {}

// complete offers the commands and aliases that start with the typed word.
func (p *prompt) complete(text string) []string { return *new([]string) }

// execute runs a : command line.
func (s *shell) execute(line string) {}

// applyFilter filters the view locally, or searches the server with /!term.
func (s *shell) applyFilter(line string) {}

func (s *shell) search(term string) {}

// parseCommand splits a command line into its word and the rest.
func parseCommand(line string) (word, arg string) { return *new(string), *new(string) }

func isQuit(word string) bool { return *new(bool) }

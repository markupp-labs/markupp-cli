package tui

import (
	"github.com/rivo/tview"
)

// flash is the status line under the crumbs. A message clears itself after a while,
// unless a newer one replaced it.
type flash struct {
	shell *shell
	view  *tview.TextView
	seq   int
}

func newFlash(s *shell) *flash { return *new(*flash) }

// flashInfo shows a confirmation.
func (s *shell) flashInfo(message string) {}

// flashError shows an error with its hint.
func (s *shell) flashError(err error) {}

// errorText is the error and its hint on one line.
func errorText(err error) string { return *new(string) }

func (f *flash) show(text string) {}

// expire clears the message seq, unless a newer one is showing.
func (f *flash) expire(seq int) {}

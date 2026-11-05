package tui

import (
	"time"

	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// flash is the status line under the crumbs. A message clears itself after a while,
// unless a newer one replaced it.
type flash struct {
	shell *shell
	view  *tview.TextView
	seq   int
}

func newFlash(s *shell) *flash {
	return &flash{shell: s, view: tview.NewTextView().SetDynamicColors(true)}
}

// flashInfo shows a confirmation.
func (s *shell) flashInfo(message string) {
	s.flash.show("[green]" + plain(message) + "[-]")
}

// flashError shows an error with its hint.
func (s *shell) flashError(err error) {
	s.flash.show("[red]" + plain(errorText(err)) + "[-]")
}

// errorText is the error and its hint on one line.
func errorText(err error) string {
	text := "error: " + err.Error()
	if hint := failure.HintOf(err); hint != "" {
		text += " (hint: " + hint + ")"
	}
	return text
}

func (f *flash) show(text string) {
	f.seq++
	f.view.SetText(text)
	seq := f.seq
	time.AfterFunc(f.shell.cfg.flashFor(), func() {
		f.shell.app.QueueUpdateDraw(func() { f.expire(seq) })
	})
}

// expire clears the message seq, unless a newer one is showing.
func (f *flash) expire(seq int) {
	if f.seq == seq {
		f.view.SetText("")
	}
}

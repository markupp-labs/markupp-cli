package tui

import (
	"github.com/rivo/tview"
)

const (
	dialogPage   = "dialog"
	cancelButton = "Cancel"
)

// confirm asks before action runs. Cancel has the focus, so Enter alone cancels and
// runs cancel, which may be nil. It is the only question mu ever asks, and it lives
// inside mu tui.
func (s *shell) confirm(text, action string, run, cancel func()) {
	modal := tview.NewModal().SetText(text).AddButtons([]string{cancelButton, action})
	modal.SetDoneFunc(func(_ int, label string) {
		s.closeDialog()
		switch {
		case label == action:
			run()
		case cancel != nil:
			cancel()
		}
	})
	s.overlay = true
	s.pages.AddPage(dialogPage, modal, true, true)
	s.app.SetFocus(modal)
}

func (s *shell) closeDialog() {
	s.pages.RemovePage(dialogPage)
	s.overlay = false
	s.focusTop()
}

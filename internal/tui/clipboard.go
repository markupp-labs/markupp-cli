package tui

import (
	"encoding/base64"
	"fmt"
)

// osc52 is the escape sequence that asks the terminal to put text on the clipboard.
// It works over ssh and needs no clipboard program.
func osc52(text string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\a"
}

// copyText puts text on the clipboard of the terminal.
func (s *shell) copyText(text string) {
	if text == "" {
		s.flashError(fmt.Errorf("nothing to copy, expected a selected row"))
		return
	}
	out := s.cfg.Terminal.Out
	if out == nil {
		s.flashError(fmt.Errorf("no terminal to copy %q to", text))
		return
	}
	if _, err := fmt.Fprint(out, osc52(text)); err != nil {
		s.flashError(fmt.Errorf("copying %q to the clipboard: %w", text, err))
		return
	}
	s.flashInfo("copied " + text)
}

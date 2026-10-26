package tui

import (
	"github.com/gdamore/tcell/v2"
)

// globalHotkeys are the keys of every view, listed in ? help.
func (s *shell) globalHotkeys() []Hotkey { return *new([]Hotkey) }

// globalKey runs a global hotkey, reporting whether event was one.
func (s *shell) globalKey(event *tcell.EventKey) bool { return *new(bool) }

// escape clears the filter of the view first, then goes back.
func (s *shell) escape() {}

func (s *shell) togglePause() {}

// open runs a : command without argument, flashing its error.
func (s *shell) open(command func(s *shell, arg string) error) {}

func (s *shell) showHelp() {}

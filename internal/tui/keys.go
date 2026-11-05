package tui

import (
	"github.com/gdamore/tcell/v2"
)

// globalHotkeys are the keys of every view, listed in ? help.
func (s *shell) globalHotkeys() []Hotkey {
	return []Hotkey{
		runeKey(':', "command", func() { s.prompt.open(promptCommand) }),
		runeKey('/', "filter, /!term searches", func() { s.prompt.open(promptFilter) }),
		runeKey('?', "help", s.showHelp),
		ctrlKey(tcell.KeyEscape, "back", s.escape),
		ctrlKey(tcell.KeyCtrlR, "refresh", s.reload),
		runeKey('t', "switch tenant", func() { s.open(openTenants) }),
		runeKey('p', "pause auto refresh", s.togglePause),
	}
}

// globalKey runs a global hotkey, reporting whether event was one.
func (s *shell) globalKey(event *tcell.EventKey) bool {
	key, ok := findHotkey(s.globalHotkeys(), event)
	if ok {
		key.Run()
	}
	return ok
}

// escape clears the filter of the view first, then goes back.
func (s *shell) escape() {
	if f, ok := s.stack.top().(Filterable); ok && f.Filter() != "" {
		f.SetFilter("")
		return
	}
	s.back()
}

func (s *shell) togglePause() {
	s.paused = !s.paused
	if s.paused {
		s.flashInfo("auto refresh paused, p resumes")
	} else {
		s.flashInfo("auto refresh resumed")
	}
	s.refreshChrome()
}

// open runs a : command without argument, flashing its error.
func (s *shell) open(command func(s *shell, arg string) error) {
	if err := command(s, ""); err != nil {
		s.flashError(err)
	}
}

func (s *shell) showHelp() {
	s.push(newHelpView(s, s.stack.top()))
}

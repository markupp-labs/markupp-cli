package tui

const (
	dialogPage   = "dialog"
	cancelButton = "Cancel"
)

// confirm asks before action runs. Cancel has the focus, so Enter alone cancels.
// It is the only question mu ever asks, and it lives inside mu tui.
func (s *shell) confirm(text, action string, run func()) {}

func (s *shell) closeDialog() {}

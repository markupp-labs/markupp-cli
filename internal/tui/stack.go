package tui

// stack is the view stack: Enter pushes, Esc pops, : replaces the whole stack.
type stack struct {
	views []View
}

func (s *stack) push(v View) {}

// pop removes the top view, keeping the last one: there is always a view to show.
func (s *stack) pop() bool { return *new(bool) }

// reset makes v the only view.
func (s *stack) reset(v View) {}

// top returns the current view, nil while the stack is empty.
func (s *stack) top() View { return *new(View) }

// crumbs are the labels of the views, bottom first.
func (s *stack) crumbs() []string { return *new([]string) }

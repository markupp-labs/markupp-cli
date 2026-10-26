package tui

// stack is the view stack: Enter pushes, Esc pops, : replaces the whole stack.
type stack struct {
	views []View
}

func (s *stack) push(v View) { s.views = append(s.views, v) }

// pop removes the top view, keeping the last one: there is always a view to show.
func (s *stack) pop() bool {
	if len(s.views) < 2 {
		return false
	}
	s.views = s.views[:len(s.views)-1]
	return true
}

// reset makes v the only view.
func (s *stack) reset(v View) { s.views = []View{v} }

// top returns the current view, nil while the stack is empty.
func (s *stack) top() View {
	if len(s.views) == 0 {
		return nil
	}
	return s.views[len(s.views)-1]
}

// crumbs are the labels of the views, bottom first.
func (s *stack) crumbs() []string {
	labels := make([]string, 0, len(s.views))
	for _, v := range s.views {
		labels = append(labels, v.Crumb())
	}
	return labels
}

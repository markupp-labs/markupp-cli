package tui

import (
	"context"

	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// graphDepth is the depth of each read: the direct neighbors of a note. Deeper levels
// load when their node is expanded.
const graphDepth = 1

// openGraph is :graph ID.
func openGraph(s *shell, id string) error { return *new(error) }

// pushGraph opens the graph of a note on top of the stack.
func (s *shell) pushGraph(id, label string) {}

// graphView is the XRay of a note: a tree of its links, in and out, where each
// neighbor expands into its own links on Enter.
type graphView struct {
	shell   *shell
	id      string
	label   string
	tree    *tview.TreeView
	entries map[*tview.TreeNode]*graphEntry
}

// graphEntry is the note behind a tree node; id is empty for an unresolved title.
type graphEntry struct {
	id     string
	label  string
	loaded bool
}

func newGraphView(s *shell, id, label string) *graphView { return *new(*graphView) }

func (v *graphView) Crumb() string { return *new(string) }

func (v *graphView) Title() string { return *new(string) }

func (v *graphView) Widget() tview.Primitive { return *new(tview.Primitive) }

// Load reads the neighbors of the root again, collapsing what was expanded.
func (v *graphView) Load() {}

func (v *graphView) Hotkeys() []Hotkey { return *new([]Hotkey) }

// onCurrent binds a note action to the note of the current node.
func (v *graphView) onCurrent(run func(r record)) func() { return *new(func()) }

func (v *graphView) current() graphEntry { return *new(graphEntry) }

// expand loads the neighbors of node the first time, then folds and unfolds it.
func (v *graphView) expand(node *tview.TreeNode) {}

func (v *graphView) read(ctx context.Context, id string) (schema.Graph, error) {
	return *new(schema.Graph), *new(error)
}

// grow replaces the children of node with the neighbors of its note.
func (v *graphView) grow(node *tview.TreeNode, entry *graphEntry, graph schema.Graph) {}

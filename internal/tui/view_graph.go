package tui

import (
	"context"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// graphDepth is the depth of each read: the direct neighbors of a note. Deeper levels
// load when their node is expanded.
const graphDepth = 1

// openGraph is :graph ID.
func openGraph(s *shell, id string) error {
	if err := s.need(s.ports.Graph != nil, "graph"); err != nil {
		return err
	}
	if id == "" {
		return failure.New(failure.Usage, ":graph", "missing note id, expected :graph <note id>, or g on a note")
	}
	s.reset(newGraphView(s, id, id))
	return nil
}

// pushGraph opens the graph of a note on top of the stack.
func (s *shell) pushGraph(id, label string) {
	if err := s.need(s.ports.Graph != nil, "graph"); err != nil {
		s.flashError(err)
		return
	}
	s.push(newGraphView(s, id, label))
}

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

func newGraphView(s *shell, id, label string) *graphView {
	root := tview.NewTreeNode(plain(label)).SetColor(tcell.ColorOrange)
	v := &graphView{shell: s, id: id, label: label, entries: map[*tview.TreeNode]*graphEntry{}}
	v.entries[root] = &graphEntry{id: id, label: label}
	v.tree = tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	return v
}

func (v *graphView) Crumb() string { return "graph" }

func (v *graphView) Title() string { return "graph(" + v.label + ")" }

func (v *graphView) Widget() tview.Primitive { return v.tree }

// Load reads the neighbors of the root again, collapsing what was expanded.
func (v *graphView) Load() {
	root := v.tree.GetRoot()
	v.entries[root].loaded = false
	v.expand(root)
}

func (v *graphView) Hotkeys() []Hotkey {
	keys := []Hotkey{ctrlKey(tcell.KeyEnter, "expand", func() { v.expand(v.tree.GetCurrentNode()) })}
	keys = append(keys, v.shell.noteActions(v.onCurrent)...)
	return append(keys, runeKey('y', "copy id", func() { v.shell.copyText(v.current().id) }))
}

// onCurrent binds a note action to the note of the current node.
func (v *graphView) onCurrent(run func(r record)) func() {
	return func() {
		entry := v.current()
		if entry.id == "" {
			v.shell.flashError(failure.New(failure.Usage, entry.label, "unresolved title, expected a node that is a note"))
			return
		}
		run(record{id: entry.id, note: entry.id, label: entry.label})
	}
}

func (v *graphView) current() graphEntry {
	if entry, ok := v.entries[v.tree.GetCurrentNode()]; ok {
		return *entry
	}
	return graphEntry{}
}

// expand loads the neighbors of node the first time, then folds and unfolds it.
func (v *graphView) expand(node *tview.TreeNode) {
	entry, ok := v.entries[node]
	if !ok || entry.id == "" {
		return
	}
	if entry.loaded {
		node.SetExpanded(!node.IsExpanded())
		return
	}
	v.shell.async(func(ctx context.Context) (func(), error) {
		graph, err := v.read(ctx, entry.id)
		return func() { v.grow(node, entry, graph) }, err
	})
}

func (v *graphView) read(ctx context.Context, id string) (schema.Graph, error) {
	body, err := v.shell.ports.Graph.NoteGraph(ctx, id, graphDepth)
	if err != nil {
		return schema.Graph{}, err
	}
	return decodeBody[schema.Graph](body)
}

// grow replaces the children of node with the neighbors of its note.
func (v *graphView) grow(node *tview.TreeNode, entry *graphEntry, graph schema.Graph) {
	node.ClearChildren()
	for _, n := range neighbors(graph, entry.id) {
		child := tview.NewTreeNode(plain(n.text())).SetExpanded(false)
		v.entries[child] = &graphEntry{id: n.noteID(), label: n.label()}
		node.AddChild(child)
	}
	entry.loaded = true
	node.SetExpanded(true)
	v.shell.refreshChrome()
}

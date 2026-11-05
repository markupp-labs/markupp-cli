package tui

import (
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// Arrows of the XRay tree: a link from the note, and a link to it.
const (
	arrowOut = "->"
	arrowIn  = "<-"
)

// neighbor is a node linked to the center of a graph, and the direction of the link.
type neighbor struct {
	arrow    string
	linkType string
	node     schema.GraphNode
}

// neighbors are the nodes linked to the note center, outgoing links first, in the
// order of the server.
func neighbors(graph schema.Graph, center string) []neighbor {
	index, ok := nodeIndex(graph.Nodes, center)
	if !ok {
		return nil
	}
	out := linked(graph, index, arrowOut)
	return append(out, linked(graph, index, arrowIn)...)
}

func nodeIndex(nodes []schema.GraphNode, id string) (int, bool) {
	for i, node := range nodes {
		if node.ID != nil && *node.ID == id {
			return i, true
		}
	}
	return 0, false
}

// linked are the far ends of the links of node index in one direction.
func linked(graph schema.Graph, index int, arrow string) []neighbor {
	var found []neighbor
	for _, link := range graph.Links {
		near, far := link.Source, link.Target
		if arrow == arrowIn {
			near, far = far, near
		}
		if near == index && far >= 0 && far < len(graph.Nodes) {
			found = append(found, neighbor{arrow: arrow, linkType: link.Type, node: graph.Nodes[far]})
		}
	}
	return found
}

// text is the line of the neighbor in the tree.
func (n neighbor) text() string {
	return n.arrow + " " + n.linkType + "  " + n.label()
}

// label is the path of the note, or the cited title of an unresolved node.
func (n neighbor) label() string {
	if n.node.Path != nil && *n.node.Path != "" {
		return *n.node.Path
	}
	if n.node.Unresolved {
		return n.node.Title + " (unresolved)"
	}
	return n.node.Title
}

func (n neighbor) noteID() string {
	if n.node.ID == nil {
		return ""
	}
	return *n.node.ID
}

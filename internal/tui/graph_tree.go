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
func neighbors(graph schema.Graph, center string) []neighbor { return *new([]neighbor) }

func nodeIndex(nodes []schema.GraphNode, id string) (int, bool) { return *new(int), *new(bool) }

// linked are the far ends of the links of node index in one direction.
func linked(graph schema.Graph, index int, arrow string) []neighbor { return *new([]neighbor) }

// text is the line of the neighbor in the tree.
func (n neighbor) text() string { return *new(string) }

// label is the path of the note, or the cited title of an unresolved node.
func (n neighbor) label() string { return *new(string) }

func (n neighbor) noteID() string { return *new(string) }

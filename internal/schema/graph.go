package schema

// Graph is a graph to draw, GraphResponse in the API: links refer to nodes by index.
type Graph struct {
	Nodes []GraphNode
	Links []GraphLink
}

// GraphNode is a note, or a cited title without a note when Unresolved is true.
type GraphNode struct {
	// ID and Path are nil on a node without a note.
	ID    *string
	Title string
	Path  *string
	NoteRank
	Unresolved bool
}

// GraphLink is a link between two nodes, by their index in Graph.Nodes.
type GraphLink struct {
	Source    int
	Target    int
	Type      string
	Predicate *string
}

// Community is a community of the notes of a vault, CommunityResponse in the API.
type Community struct {
	ID   int64
	Size int
	Hub  NoteSummary
}

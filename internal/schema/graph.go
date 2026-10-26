package schema

// Graph is a graph to draw, GraphResponse in the API: links refer to nodes by index.
type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// GraphNode is a note, or a cited title without a note when Unresolved is true.
type GraphNode struct {
	// ID and Path are nil on a node without a note.
	ID    *string `json:"id"`
	Title string  `json:"title"`
	Path  *string `json:"path"`
	NoteRank
	Unresolved bool `json:"unresolved"`
}

// GraphLink is a link between two nodes, by their index in Graph.Nodes.
type GraphLink struct {
	Source    int     `json:"source"`
	Target    int     `json:"target"`
	Type      string  `json:"type"`
	Predicate *string `json:"predicate"`
}

// Community is a community of the notes of a vault, CommunityResponse in the API.
type Community struct {
	ID   int64       `json:"id"`
	Size int         `json:"size"`
	Hub  NoteSummary `json:"hub"`
}

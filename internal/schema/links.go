package schema

// Link is one link between notes, LinkResponse in the API.
type Link struct {
	Source NoteSummary `json:"source"`
	// Target is nil when the link is unresolved.
	Target     *NoteSummary `json:"target"`
	CitedTitle string       `json:"cited_title"`
	// Type is explicit, embed, typed or inferred.
	Type string `json:"type"`
	// Predicate is the frontmatter key of a typed link, nil for the other types.
	Predicate *string `json:"predicate"`
	// InferredTarget is the indexer's guess for an unresolved link, nil otherwise.
	InferredTarget *NoteSummary `json:"inferred_target"`
}

// Mention is a title cited without a link, MentionResponse in the API.
type Mention struct {
	Note       NoteSummary `json:"note"`
	CitedTitle string      `json:"cited_title"`
	Prior      float64     `json:"prior"`
}

// LinkSuggestion is a note worth linking, LinkSuggestionResponse in the API.
type LinkSuggestion struct {
	Note               NoteSummary `json:"note"`
	Score              float64     `json:"score"`
	ResourceAllocation float64     `json:"resource_allocation"`
	Cosine             float64     `json:"cosine"`
}

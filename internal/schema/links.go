package schema

// Link is one link between notes, LinkResponse in the API.
type Link struct {
	Source NoteSummary
	// Target is nil when the link is unresolved.
	Target     *NoteSummary
	CitedTitle string
	// Type is explicit, embed, typed or inferred.
	Type string
	// Predicate is the frontmatter key of a typed link, nil for the other types.
	Predicate *string
	// InferredTarget is the indexer's guess for an unresolved link, nil otherwise.
	InferredTarget *NoteSummary
}

// Mention is a title cited without a link, MentionResponse in the API.
type Mention struct {
	Note       NoteSummary
	CitedTitle string
	Prior      float64
}

// LinkSuggestion is a note worth linking, LinkSuggestionResponse in the API.
type LinkSuggestion struct {
	Note               NoteSummary
	Score              float64
	ResourceAllocation float64
	Cosine             float64
}

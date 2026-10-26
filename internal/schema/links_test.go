package schema

import (
	"encoding/json"
	"testing"
)

// decodable lists the types the tests decode, instead of an unconstrained any.
type decodable interface {
	Link | Mention | LinkSuggestion | Graph | Community
}

func decode[T decodable](t *testing.T, body string) T {
	t.Helper()
	var got T
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	return got
}

func TestLinkDecodesATypedResolvedLink(t *testing.T) {
	got := decode[Link](t, `{"source":{"id":"n1","title":"a","path":"a.md"},`+
		`"target":{"id":"n2","title":"b","path":"b.md"},"cited_title":"B","type":"typed",`+
		`"predicate":"author","inferred_target":null}`)
	if got.Source.ID != "n1" || got.Target == nil || got.Target.Path != "b.md" {
		t.Fatalf("ends = %+v -> %+v", got.Source, got.Target)
	}
	if got.CitedTitle != "B" || got.Type != "typed" || got.Predicate == nil || *got.Predicate != "author" {
		t.Errorf("link = %+v", got)
	}
	if got.InferredTarget != nil {
		t.Errorf("inferred_target = %+v, want nil", got.InferredTarget)
	}
}

func TestLinkDecodesAnUnresolvedLinkWithAGuess(t *testing.T) {
	got := decode[Link](t, `{"source":{"id":"n1"},"target":null,"cited_title":"Meting",`+
		`"type":"explicit","predicate":null,"inferred_target":{"id":"n3","path":"meeting.md"}}`)
	if got.Target != nil || got.Predicate != nil {
		t.Errorf("target %+v and predicate %v, want nil", got.Target, got.Predicate)
	}
	if got.InferredTarget == nil || got.InferredTarget.ID != "n3" {
		t.Errorf("inferred_target = %+v, want n3", got.InferredTarget)
	}
}

func TestMentionDecodesPrior(t *testing.T) {
	got := decode[Mention](t, `{"note":{"id":"n2","path":"k.md"},"cited_title":"kubernetes","prior":0.33}`)
	want := Mention{Note: NoteSummary{ID: "n2", Path: "k.md"}, CitedTitle: "kubernetes", Prior: 0.33}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLinkSuggestionDecodesScores(t *testing.T) {
	got := decode[LinkSuggestion](t, `{"note":{"id":"n4"},"score":0.0325,"resource_allocation":0.5,"cosine":0.38}`)
	want := LinkSuggestion{Note: NoteSummary{ID: "n4"}, Score: 0.0325, ResourceAllocation: 0.5, Cosine: 0.38}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

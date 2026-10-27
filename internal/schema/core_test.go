package schema

import (
	"encoding/json"
	"testing"
)

func TestErrorResponseDecodesCodeAndMessage(t *testing.T) {
	var got ErrorResponse
	body := `{"error":"duplicate_path","message":"path taken: a.md"}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.Error != "duplicate_path" || got.Message != "path taken: a.md" {
		t.Errorf("got %+v", got)
	}
}

func TestNoteSummaryDecodesSnakeCaseFields(t *testing.T) {
	var got NoteSummary
	body := `{"id":"n1","title":"meeting","path":"notas/meeting.md"}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	want := NoteSummary{ID: "n1", Title: "meeting", Path: "notas/meeting.md"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestNoteRankKeepsNullAsNil(t *testing.T) {
	var got NoteRank
	if err := json.Unmarshal([]byte(`{"degree":null,"pagerank":null,"community":null}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Degree != nil || got.PageRank != nil || got.Community != nil {
		t.Errorf("null fields decoded as %+v, want all nil", got)
	}
}

func TestNoteRankDecodesValues(t *testing.T) {
	var got NoteRank
	if err := json.Unmarshal([]byte(`{"degree":15,"pagerank":0.0123,"community":7}`), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Degree == nil || *got.Degree != 15 {
		t.Fatalf("degree = %v, want 15", got.Degree)
	}
	if got.PageRank == nil || *got.PageRank != 0.0123 {
		t.Fatalf("pagerank = %v, want 0.0123", got.PageRank)
	}
	if got.Community == nil || *got.Community != 7 {
		t.Fatalf("community = %v, want 7", got.Community)
	}
}

func TestNoteSummaryEncodesWithAPIFieldNames(t *testing.T) {
	got, err := json.Marshal(NoteSummary{ID: "n1", Title: "t", Path: "t.md"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(got) != `{"id":"n1","title":"t","path":"t.md"}` {
		t.Errorf("Marshal = %s", got)
	}
}

func TestNoteRankEncodesWithAPIFieldNames(t *testing.T) {
	got, err := json.Marshal(NoteRank{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(got) != `{"degree":null,"pagerank":null,"community":null}` {
		t.Errorf("Marshal = %s", got)
	}
}

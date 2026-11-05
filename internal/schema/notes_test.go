package schema

import (
	"encoding/json"
	"testing"
)

func TestNoteDecodesContentAndRank(t *testing.T) {
	var got Note
	body := `{"id":"n1","vault_id":"v1","path":"a.md","content":"# A","content_sha256":"ab",` +
		`"created_at":"2026-05-14T12:00:00Z","updated_at":"2026-05-15T12:00:00Z","degree":3,"pagerank":null,"community":null}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.ID != "n1" || got.VaultID != "v1" || got.Content != "# A" || got.UpdatedAt != "2026-05-15T12:00:00Z" {
		t.Errorf("got %+v", got)
	}
	if got.Degree == nil || *got.Degree != 3 || got.PageRank != nil {
		t.Errorf("rank = %+v, want degree 3 and no pagerank", got.NoteRank)
	}
}

func TestNoteListItemDecodesSummaryAndUpdatedAt(t *testing.T) {
	var got NoteListItem
	body := `{"id":"n1","title":"a","path":"a.md","updated_at":"2026-05-15T12:00:00Z","community":7}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.ID != "n1" || got.Title != "a" || got.UpdatedAt != "2026-05-15T12:00:00Z" || *got.Community != 7 {
		t.Errorf("got %+v", got)
	}
}

func TestSearchHitDecodesNestedNote(t *testing.T) {
	var got SearchHit
	body := `{"note":{"id":"n1","title":"a","path":"a.md"},"vault_id":"v1","rank":0.5,"snippet":"x"}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.Note.ID != "n1" || got.VaultID != "v1" || got.Rank != 0.5 || got.Snippet != "x" {
		t.Errorf("got %+v", got)
	}
}

func TestNoteUpdateRequestEncodesAPIFieldNames(t *testing.T) {
	got, err := json.Marshal(NoteUpdateRequest{Path: "a.md", Content: "x", LastModifiedAt: "2026-05-15T12:00:00Z"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"path":"a.md","content":"x","lastModifiedAt":"2026-05-15T12:00:00Z","force":false}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

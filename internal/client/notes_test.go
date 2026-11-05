package client

import (
	"context"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

func TestListNotesReadsTheVaultCollection(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	body, err := newTestClient(t, server, "").ListNotes(context.Background(), "v/1", NoteFilter{}, Page{Limit: 5, Offset: 2})
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if got := readAll(t, body); got != `[]` {
		t.Errorf("body = %q", got)
	}
	got := server.seen()[0]
	if got.method != "GET" || got.path != "/v1/vaults/v%2F1/notes" || got.query != "limit=5&offset=2" {
		t.Errorf("request = %+v", got)
	}
}

func TestListNotesSendsTheFilters(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	orphan, community := false, int64(3)
	filter := NoteFilter{Orphan: &orphan, Community: &community, Sort: "degree", Direction: "asc"}
	if err := opened(newTestClient(t, server, "").ListNotes(context.Background(), "v1", filter, Page{Limit: 5})); err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	if got := server.seen()[0].query; got != "community=3&direction=asc&limit=5&orphan=false&sort=degree" {
		t.Errorf("query = %q", got)
	}
}

func TestGetNoteReadsOneNote(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{"id":"n1"}`))
	body, err := newTestClient(t, server, "").GetNote(context.Background(), "n1")
	if err != nil {
		t.Fatalf("GetNote: %v", err)
	}
	if got := readAll(t, body); got != `{"id":"n1"}` {
		t.Errorf("body = %q", got)
	}
	if got := server.seen()[0]; got.method != "GET" || got.path != "/v1/notes/n1" {
		t.Errorf("request = %+v", got)
	}
}

func TestGetNoteMapsNotFound(t *testing.T) {
	server := newFakeServer(t, replyJSON(404, `{"error":"not_found","message":"no note"}`))
	_, err := newTestClient(t, server, "").GetNote(context.Background(), "n9")
	if failure.KindOf(err) != failure.NotFound {
		t.Errorf("GetNote error = %v, want not found", err)
	}
}

func TestSearchNotesSendsQueryVaultAndPage(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	body, err := newTestClient(t, server, "").SearchNotes(context.Background(), "a b", "v1", Page{Limit: 3})
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	_ = body.Close()
	got := server.seen()[0]
	if got.method != "GET" || got.path != "/v1/search/notes" || got.query != "limit=3&q=a+b&vault=v1" {
		t.Errorf("request = %+v", got)
	}
}

func TestSearchNotesWithoutVaultSearchesEveryVault(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `[]`))
	body, err := newTestClient(t, server, "").SearchNotes(context.Background(), "x", "", Page{})
	if err != nil {
		t.Fatalf("SearchNotes: %v", err)
	}
	_ = body.Close()
	if got := server.seen()[0].query; got != "q=x" {
		t.Errorf("query = %q, want q=x", got)
	}
}

func TestCreateNotePostsToTheVault(t *testing.T) {
	server := newFakeServer(t, replyJSON(201, `{"id":"n1"}`))
	note := schema.NoteCreateRequest{Path: "a.md", Content: "# A"}
	body, err := newTestClient(t, server, "").CreateNote(context.Background(), "v1", note)
	if err != nil {
		t.Fatalf("CreateNote: %v", err)
	}
	if got := readAll(t, body); got != `{"id":"n1"}` {
		t.Errorf("body = %q", got)
	}
	got := server.seen()[0]
	if got.method != "POST" || got.path != "/v1/vaults/v1/notes" || got.body != `{"path":"a.md","content":"# A"}` {
		t.Errorf("request = %+v", got)
	}
}

func TestUpdateNotePutsTheVersion(t *testing.T) {
	server := newFakeServer(t, replyJSON(200, `{"id":"n1"}`))
	update := schema.NoteUpdateRequest{Path: "a.md", Content: "x", LastModifiedAt: "2026-05-15T12:00:00Z"}
	body, err := newTestClient(t, server, "").UpdateNote(context.Background(), "n1", update)
	if err != nil {
		t.Fatalf("UpdateNote: %v", err)
	}
	_ = body.Close()
	got := server.seen()[0]
	want := `{"path":"a.md","content":"x","lastModifiedAt":"2026-05-15T12:00:00Z","force":false}`
	if got.method != "PUT" || got.path != "/v1/notes/n1" || got.body != want {
		t.Errorf("request = %+v", got)
	}
}

func TestUpdateNoteConflictIsNotRetried(t *testing.T) {
	server := newFakeServer(t, replyJSON(409, `{"error":"conflict","message":"changed"}`))
	_, err := newTestClient(t, server, "").UpdateNote(context.Background(), "n1", schema.NoteUpdateRequest{})
	if failure.KindOf(err) != failure.Conflict {
		t.Errorf("UpdateNote error = %v, want conflict", err)
	}
	if n := len(server.seen()); n != 1 {
		t.Errorf("server saw %d requests, want 1", n)
	}
}

func TestDeleteNoteSendsDelete(t *testing.T) {
	server := newFakeServer(t, replyJSON(204, ""))
	if err := newTestClient(t, server, "").DeleteNote(context.Background(), "n1"); err != nil {
		t.Fatalf("DeleteNote: %v", err)
	}
	if got := server.seen()[0]; got.method != "DELETE" || got.path != "/v1/notes/n1" {
		t.Errorf("request = %+v", got)
	}
}

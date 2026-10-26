package schema

import (
	"encoding/json"
	"testing"
)

func TestMyTenantDecodesTheSummary(t *testing.T) {
	var got []MyTenant
	body := `[{"id":"t1","name":"Acme","role":"owner","notes_count":3,"last_note_at":"2026-10-01T10:00:00Z"},{"id":"t2","name":"Empty","role":"viewer","notes_count":0,"last_note_at":null}]`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if len(got) != 2 || got[0].Role != "owner" || got[0].NotesCount != 3 || got[0].LastNoteAt == nil || got[1].LastNoteAt != nil {
		t.Errorf("got %+v", got)
	}
}

func TestUserDecodesTheAccount(t *testing.T) {
	var got User
	body := `{"id":"u1","email":"ana@example.com","email_verified":true,"password_login":true}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if want := (User{ID: "u1", Email: "ana@example.com", EmailVerified: true}); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

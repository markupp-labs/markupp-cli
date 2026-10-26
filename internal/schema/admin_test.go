package schema

import (
	"encoding/json"
	"testing"
)

func TestIsRoleAcceptsOnlyTheSchemaRoles(t *testing.T) {
	for _, role := range []string{"owner", "admin", "member", "viewer"} {
		if !IsRole(role) {
			t.Errorf("IsRole(%q) = false, want true", role)
		}
	}
	for _, role := range []string{"", "Owner", "editor", "admin "} {
		if IsRole(role) {
			t.Errorf("IsRole(%q) = true, want false", role)
		}
	}
}

func TestMyTenantDecodesTheEmbeddedTenant(t *testing.T) {
	var got MyTenant
	body := `{"id":"t1","name":"Acme","role":"owner","notes_count":3,"last_note_at":null}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.ID != "t1" || got.Name != "Acme" || got.Role != "owner" || got.NotesCount != 3 || got.LastNoteAt != nil {
		t.Errorf("got %+v", got)
	}
}

func TestIssuedInviteDecodesTheInviteAndTheToken(t *testing.T) {
	var got IssuedInvite
	body := `{"id":"i1","email":"a@b.c","role":"member","expires_at":"2026-11-02T00:00:00Z","token":"tk"}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	if got.ID != "i1" || got.Email != "a@b.c" || got.Token != "tk" || got.AcceptedAt != "" {
		t.Errorf("got %+v", got)
	}
}

func TestPendingInviteDecodesSnakeCaseFields(t *testing.T) {
	var got PendingInvite
	body := `{"id":"i1","tenant_id":"t1","tenant_name":"Acme","role":"viewer","expires_at":"x"}`
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("Unmarshal(%s): %v", body, err)
	}
	want := PendingInvite{ID: "i1", TenantID: "t1", TenantName: "Acme", Role: "viewer", ExpiresAt: "x"}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

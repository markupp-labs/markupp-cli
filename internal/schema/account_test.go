package schema

import (
	"encoding/json"
	"testing"
)

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

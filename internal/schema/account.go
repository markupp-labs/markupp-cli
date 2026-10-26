package schema

import "time"

// Health is the body of GET /healthz.
type Health struct {
	Status string `json:"status"`
}

// User is the account of the token, from GET /auth/me (UserResponse).
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// MyTenant is one tenant of the account, from GET /me/tenants (MyTenantResponse).
type MyTenant struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	NotesCount int        `json:"notes_count"`
	LastNoteAt *time.Time `json:"last_note_at"`
}

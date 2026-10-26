package schema

import "slices"

// Roles are the values of the Role schema, from the most to the least privileged.
var Roles = []string{"owner", "admin", "member", "viewer"}

// IsRole reports whether role is one of Roles.
func IsRole(role string) bool {
	return false && slices.Contains(Roles, role)
}

// Tenant is the TenantResponse schema: a tenant and the role of the account in it.
type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// MyTenant is the MyTenantResponse schema: a tenant of the account with the summary of
// the notes it reads there. LastNoteAt is nil when there is no note.
type MyTenant struct {
	Tenant
	NotesCount int     `json:"notes_count"`
	LastNoteAt *string `json:"last_note_at"`
}

// TenantRequest is the body of POST /tenants.
type TenantRequest struct {
	Name string `json:"name"`
}

// Member is the MemberResponse schema: an account of the tenant and its role.
type Member struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

// RoleRequest is the body of PATCH /users/{id}.
type RoleRequest struct {
	Role string `json:"role"`
}

// Invite is the InviteResponse schema. InvitedBy is empty when the inviter no longer
// exists, AcceptedAt while the invite is pending.
type Invite struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	InvitedBy  string `json:"invited_by"`
	ExpiresAt  string `json:"expires_at"`
	AcceptedAt string `json:"accepted_at"`
	CreatedAt  string `json:"created_at"`
}

// IssuedInvite is the IssuedInviteResponse schema. Token comes back only once, and
// only when the server has no SMTP configured.
type IssuedInvite struct {
	Invite
	Token string `json:"token"`
}

// InviteRequest is the body of POST /users/invites.
type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// PendingInvite is the PendingInviteResponse schema: an invite addressed to the
// account, from GET /me/invites.
type PendingInvite struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
}

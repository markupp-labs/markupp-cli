package schema

// NotificationInvite is the kind of a notification about a pending tenant invite.
const NotificationInvite = "invite"

// NotificationFeed is what asks for the attention of the account, oldest first.
type NotificationFeed struct {
	Items  []Notification `json:"items"`
	Unread int            `json:"unread"`
}

// Notification is one item of the feed. Kind says which fields come; with
// NotificationInvite they are those of a pending invite, and ID is the invite id.
type Notification struct {
	Kind       string `json:"kind"`
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
}

package tui

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// memberRecord is a member of the tenant or of a vault (MemberResponse).
type memberRecord struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joined_at"`
}

// inviteRecord is an invite addressed to the account (PendingInviteResponse).
type inviteRecord struct {
	ID         string `json:"id"`
	TenantName string `json:"tenant_name"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
}

// notificationFeed is the feed of the account (NotificationFeedResponse).
type notificationFeed struct {
	Items  []notificationRecord `json:"items"`
	Unread int                  `json:"unread"`
}

// notificationRecord is one item of the feed; an invite carries the invite fields.
type notificationRecord struct {
	Kind string `json:"kind"`
	inviteRecord
}

// homeView is the first view of a scope: the notes of the vault, the vaults, or the
// tenants, whichever this build and scope can show.
func (s *shell) homeView() View { return *new(View) }

// openTenants is :tenants.
func openTenants(s *shell, _ string) error { return *new(error) }

// newTenantsView lists the tenants of the account; Enter switches to one.
func newTenantsView(s *shell) *listView { return *new(*listView) }

func tenantRow(tenant schema.MyTenant) record { return *new(record) }

func (s *shell) enterTenant(r record) {}

// copyKeys is y on a row that is not a note.
func (s *shell) copyKeys(selected func() (record, bool)) []Hotkey { return *new([]Hotkey) }

// openMembers is :members, the accounts of the tenant.
func openMembers(s *shell, _ string) error { return *new(error) }

func newMembersView(s *shell, resource, context string, fetch fetchBody) *listView {
	return *new(*listView)
}

func memberRow(member memberRecord) record { return *new(record) }

// openInvites is :invites, the invites addressed to the account.
func openInvites(s *shell, _ string) error { return *new(error) }

func inviteRow(invite inviteRecord) record { return *new(record) }

// openNotifications is :notifications, the feed of the account.
func openNotifications(s *shell, _ string) error { return *new(error) }

func (s *shell) readNotifications(ctx context.Context) ([]record, error) {
	return *new([]record), *new(error)
}

// openPublic is :public SLUG: the public vault, read without credentials.
func openPublic(s *shell, slug string) error { return *new(error) }

func (s *shell) publicHome(slug string) View { return *new(View) }

// fetchOf adapts a call with one argument to a fetchBody.
func fetchOf(call func(ctx context.Context, id string) (io.ReadCloser, error), id string) fetchBody {
	return *new(fetchBody)
}

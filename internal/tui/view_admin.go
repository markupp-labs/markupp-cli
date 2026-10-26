package tui

import (
	"context"
	"io"
	"strconv"

	"github.com/markupp-labs/markupp-cli/internal/failure"
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
func (s *shell) homeView() View {
	switch {
	case s.vault != "" && s.ports.Notes != nil:
		return newNotesView(s, s.vault)
	case s.ports.Vaults != nil:
		return newVaultsView(s)
	case s.ports.Account != nil:
		return newTenantsView(s)
	default:
		return newHelpView(s, newNoteView(s, ""))
	}
}

// openTenants is :tenants.
func openTenants(s *shell, _ string) error {
	if err := s.need(s.ports.Account != nil && s.scope.PublicSlug == "", "tenants"); err != nil {
		return err
	}
	s.reset(newTenantsView(s))
	return nil
}

// newTenantsView lists the tenants of the account; Enter switches to one.
func newTenantsView(s *shell) *listView {
	return newListView(s, listing{
		resource: "tenants",
		columns:  []string{"NAME", "ROLE", "NOTES", "ID"},
		load:     arrayOf(s.ports.Account.ListMyTenants, rowsOf(tenantRow)),
		enter:    s.enterTenant,
		keys:     s.copyKeys,
	})
}

func tenantRow(tenant schema.MyTenant) record {
	cells := []string{tenant.Name, tenant.Role, strconv.Itoa(tenant.NotesCount), tenant.ID}
	return record{id: tenant.ID, label: tenant.Name, cells: cells}
}

func (s *shell) enterTenant(r record) {
	s.switchScope(Scope{Tenant: r.id}, r.label, s.homeView)
}

// copyKeys is y on a row that is not a note.
func (s *shell) copyKeys(selected func() (record, bool)) []Hotkey {
	return []Hotkey{runeKey('y', "copy id", func() { s.copySelected(selected) })}
}

// openMembers is :members, the accounts of the tenant.
func openMembers(s *shell, _ string) error {
	if err := s.need(s.ports.Team != nil, "members"); err != nil {
		return err
	}
	s.reset(newMembersView(s, "members", "", s.ports.Team.ListMembers))
	return nil
}

func newMembersView(s *shell, resource, context string, fetch fetchBody) *listView {
	return newListView(s, listing{
		resource: resource, context: context,
		columns: []string{"EMAIL", "ROLE", "JOINED", "ID"},
		load:    arrayOf(fetch, rowsOf(memberRow)),
		keys:    s.copyKeys,
	})
}

func memberRow(member memberRecord) record {
	return record{id: member.ID, label: member.Email, cells: []string{member.Email, member.Role, member.JoinedAt, member.ID}}
}

// openInvites is :invites, the invites addressed to the account.
func openInvites(s *shell, _ string) error {
	if err := s.need(s.ports.Team != nil, "invites"); err != nil {
		return err
	}
	s.reset(newListView(s, listing{
		resource: "invites",
		columns:  []string{"TENANT", "ROLE", "EXPIRES", "ID"},
		load:     arrayOf(s.ports.Team.ListMyInvites, rowsOf(inviteRow)),
		keys:     s.copyKeys,
	}))
	return nil
}

func inviteRow(invite inviteRecord) record {
	return record{id: invite.ID, label: invite.TenantName, cells: []string{invite.TenantName, invite.Role, invite.ExpiresAt, invite.ID}}
}

// openNotifications is :notifications, the feed of the account.
func openNotifications(s *shell, _ string) error {
	if err := s.need(s.ports.Notifications != nil, "notifications"); err != nil {
		return err
	}
	s.reset(newListView(s, listing{
		resource: "notifications",
		columns:  []string{"KIND", "TENANT", "ROLE", "EXPIRES", "ID"},
		load:     s.readNotifications,
		keys:     s.copyKeys,
	}))
	return nil
}

func (s *shell) readNotifications(ctx context.Context) ([]record, error) {
	body, err := s.ports.Notifications.Notifications(ctx)
	if err != nil {
		return nil, err
	}
	feed, err := decodeBody[notificationFeed](body)
	rows := make([]record, 0, len(feed.Items))
	for _, item := range feed.Items {
		row := inviteRow(item.inviteRecord)
		rows = append(rows, record{id: row.id, label: row.label, cells: append([]string{item.Kind}, row.cells...)})
	}
	return rows, err
}

// openPublic is :public SLUG: the public vault, read without credentials.
func openPublic(s *shell, slug string) error {
	if slug == "" {
		return failure.New(failure.Usage, ":public", "missing slug, expected :public <slug>")
	}
	s.switchScope(Scope{PublicSlug: slug}, "", func() View { return s.publicHome(slug) })
	return nil
}

func (s *shell) publicHome(slug string) View {
	s.vault = slug
	if s.ports.Notes == nil {
		s.flashError(s.need(false, "public"))
		return newHelpView(s, newNoteView(s, ""))
	}
	return newNotesView(s, slug)
}

// fetchOf adapts a call with one argument to a fetchBody.
func fetchOf(call func(ctx context.Context, id string) (io.ReadCloser, error), id string) fetchBody {
	return func(ctx context.Context) (io.ReadCloser, error) { return call(ctx, id) }
}

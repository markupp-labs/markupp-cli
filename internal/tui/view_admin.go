package tui

import (
	"context"
	"io"
	"strconv"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

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

func memberRow(member schema.Member) record {
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

func inviteRow(invite schema.PendingInvite) record {
	return invitedRow(invite.ID, invite.TenantName, invite.Role, invite.ExpiresAt)
}

// invitedRow is the row of an invite to tenantName, alone or as a notification.
func invitedRow(id, tenantName, role, expiresAt string) record {
	return record{id: id, label: tenantName, cells: []string{tenantName, role, expiresAt, id}}
}

// notificationRow is an item of the feed: its kind, then the fields of the invite.
func notificationRow(item schema.Notification) record {
	row := invitedRow(item.ID, item.TenantName, item.Role, item.ExpiresAt)
	row.cells = append([]string{item.Kind}, row.cells...)
	return row
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
	feed, err := decodeBody[schema.NotificationFeed](body)
	rows := make([]record, 0, len(feed.Items))
	for _, item := range feed.Items {
		rows = append(rows, notificationRow(item))
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

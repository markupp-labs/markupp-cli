package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// MemberAdmin manages the members and the invites of the tenant. *client.Client
// satisfies it.
type MemberAdmin interface {
	ListMembers(ctx context.Context) (io.ReadCloser, error)
	ChangeMemberRole(ctx context.Context, id, role string) error
	RemoveMember(ctx context.Context, id string) error
	InviteMember(ctx context.Context, email, role string) (io.ReadCloser, error)
	ListTenantInvites(ctx context.Context) (io.ReadCloser, error)
	RevokeInvite(ctx context.Context, id string) error
}

const membersLong = `Manage the members and the invites of the tenant chosen by --tenant or
` + config.EnvTenant + `. Needs the owner or admin role in the tenant: nobody changes a
member above their own role or grants a role above it, and the tenant always keeps an
owner. Roles: owner, admin, member, viewer.

` + adminScopeNote

const removeConsequence = "the vaults where the member is the only one are deleted with their notes, and the pending invites to the member's email stop working"

func newMembersCmd(app *App) *cobra.Command {
	cmd := newGroup("members", "Manage the members and invites of the tenant", "  mu members list\n  mu members invite ana@example.com --role member")
	cmd.Long = membersLong
	cmd.AddCommand(
		newMembersListCmd(app), newMembersRoleCmd(app), newMembersRemoveCmd(app),
		newMembersInviteCmd(app), newMembersInvitesCmd(app), newMembersRevokeCmd(app),
	)
	return cmd
}

func newMembersRoleCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "role ID ROLE",
		Short:   "Change the role of a member: owner, admin, member or viewer",
		Example: "  mu members role 0b7f3c1e-5d2a-4f6b-9c8e-1a2b3c4d5e6f admin",
		Args:    exactArgs("ID", "ROLE"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.changeRole(cmd, args[0], args[1])
		},
	}
}

func (a *App) changeRole(cmd *cobra.Command, id, role string) error {
	if err := checkRole("ROLE", role); err != nil {
		return err
	}
	call := func(ctx context.Context, svc Services) error { return svc.Members.ChangeMemberRole(ctx, id, role) }
	return a.callServer(cmd, call, fmt.Sprintf("member %s is now %s", id, role))
}

func newMembersRemoveCmd(app *App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "remove ID",
		Short: "Remove a member from the tenant (requires --yes)",
		Long: "Remove a member from the tenant. It needs --yes, because " + removeConsequence + ".\n\n" +
			adminScopeNote,
		Example: "  mu members remove 0b7f3c1e-5d2a-4f6b-9c8e-1a2b3c4d5e6f --yes",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.removeMember(cmd, args[0], yes)
		},
	}
	addYesFlag(cmd, &yes, removeConsequence)
	return cmd
}

func (a *App) removeMember(cmd *cobra.Command, id string, yes bool) error {
	if err := requireYes(yes, "remove member", id); err != nil {
		return err
	}
	call := func(ctx context.Context, svc Services) error { return svc.Members.RemoveMember(ctx, id) }
	return a.callServer(cmd, call, "removed member "+id)
}

func newMembersListCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the members of the tenant, by email, with their roles",
		Example: `  mu members list
  mu members list --json | jq -r '.[] | select(.role == "viewer") | .email'`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.listFrom(cmd, ndjson, fetchMembers, memberColumns, memberRow)
		},
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func fetchMembers(ctx context.Context, svc Services) (io.ReadCloser, error) {
	return svc.Members.ListMembers(ctx)
}

var memberColumns = []output.Column{
	{Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "JOINED_AT", Width: timeColumn}, {Name: "EMAIL"},
}

func memberRow(raw json.RawMessage) ([]string, error) {
	member, err := decodeRecord[schema.Member](raw)
	if err != nil {
		return nil, err
	}
	return []string{member.ID, member.Role, shortTime(member.JoinedAt), member.Email}, nil
}

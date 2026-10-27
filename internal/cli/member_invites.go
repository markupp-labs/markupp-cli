package cli

import (
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const inviteLong = `Invite an email to the tenant with a role no higher than your own. An account
that already exists sees the invite in mu invites list. The invite lasts 7 days.

Without SMTP on the server, the answer carries the invite token, shown this one time
only: pass it to the invitee to sign up. With SMTP the server emails it instead.

` + adminScopeNote

const revokeConsequence = "the invite stops working and its token can no longer be used to join"

var tenantInviteColumns = []output.Column{
	{Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "EXPIRES_AT", Width: timeColumn},
	{Name: "ACCEPTED_AT", Width: timeColumn}, {Name: "EMAIL"},
}

func newMembersInviteCmd(app *App) *cobra.Command {
	var role string
	cmd := &cobra.Command{
		Use:     "invite EMAIL",
		Short:   "Invite an email to the tenant with --role",
		Long:    inviteLong,
		Example: "  mu members invite ana@example.com --role member\n  mu members invite ana@example.com --role viewer --json | jq -r .token",
		Args:    exactArgs("EMAIL"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.inviteMember(cmd, args[0], role)
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "role of the invitee: owner, admin, member or viewer (required)")
	return cmd
}

func (a *App) inviteMember(cmd *cobra.Command, email, role string) error {
	if err := checkRole("--role", role); err != nil {
		return err
	}
	fetch := func(ctx context.Context, svc Services) (io.ReadCloser, error) {
		return svc.Members.InviteMember(ctx, email, role)
	}
	return a.objectFrom(cmd, fetch, a.printIssuedInvite)
}

// printIssuedInvite prints the invite and, when the token came back, warns on stderr
// that it is shown only once.
func (a *App) printIssuedInvite(p *output.Printer, raw json.RawMessage) error {
	invite, err := decodeRecord[schema.IssuedInvite](raw)
	if err != nil {
		return err
	}
	fields := []output.Field{
		{Key: "id", Value: invite.ID}, {Key: "email", Value: invite.Email},
		{Key: "role", Value: invite.Role}, {Key: "expires_at", Value: shortTime(invite.ExpiresAt)},
	}
	if invite.Token == "" {
		return p.Fields(fields...)
	}
	a.confirm("the invite token is shown only this once: pass it to " + invite.Email + " to sign up")
	return p.Fields(append(fields, output.Field{Key: "token", Value: invite.Token})...)
}

func newMembersInvitesCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:   "invites",
		Short: "List the invites of the tenant, newest first; tokens are never listed",
		Example: `  mu members invites
  mu members invites --json | jq -r '.[] | select(.accepted_at == null) | .email'`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.listFrom(cmd, ndjson, fetchTenantInvites, tenantInviteColumns, tenantInviteRow)
		},
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func fetchTenantInvites(ctx context.Context, svc Services) (io.ReadCloser, error) {
	return svc.Members.ListTenantInvites(ctx)
}

func tenantInviteRow(raw json.RawMessage) ([]string, error) {
	invite, err := decodeRecord[schema.Invite](raw)
	if err != nil {
		return nil, err
	}
	return []string{invite.ID, invite.Role, shortTime(invite.ExpiresAt), shortTime(invite.AcceptedAt), invite.Email}, nil
}

func newMembersRevokeCmd(app *App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "revoke ID",
		Short:   "Revoke an unanswered invite of the tenant (requires --yes)",
		Long:    "Revoke an unanswered invite of the tenant. It needs --yes: " + revokeConsequence + ".\n\n" + adminScopeNote,
		Example: "  mu members revoke 7c9e6679-7425-40de-944b-e07fc1f90ae7 --yes",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.revokeInvite(cmd, args[0], yes)
		},
	}
	addYesFlag(cmd, &yes, revokeConsequence)
	return cmd
}

func (a *App) revokeInvite(cmd *cobra.Command, id string, yes bool) error {
	if err := requireYes(yes, "revoke invite", id); err != nil {
		return err
	}
	call := func(ctx context.Context, svc Services) error { return svc.Members.RevokeInvite(ctx, id) }
	return a.callServer(cmd, call, "revoked invite "+id)
}

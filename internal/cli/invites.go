package cli

import (
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// InviteInbox answers the invites addressed to the account. *client.Client
// satisfies it.
type InviteInbox interface {
	ListMyInvites(ctx context.Context) (io.ReadCloser, error)
	AcceptInvite(ctx context.Context, id string) (io.ReadCloser, error)
	DeclineInvite(ctx context.Context, id string) error
}

const invitesLong = `List, accept and decline the invites addressed to the email of the account.
Only an account with a verified email sees them.

` + adminScopeNote

var pendingInviteColumns = []output.Column{
	{Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "EXPIRES_AT", Width: timeColumn},
	{Name: "TENANT_ID", Width: 36}, {Name: "TENANT"},
}

func newInvitesCmd(app *App) *cobra.Command {
	cmd := newGroup("invites", "Answer the invites addressed to the account", "  mu invites list\n  mu invites accept 7c9e6679-7425-40de-944b-e07fc1f90ae7")
	cmd.Long = invitesLong
	cmd.AddCommand(newInvitesListCmd(app), newInvitesAcceptCmd(app), newInvitesDeclineCmd(app))
	return cmd
}

func newInvitesListCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List the pending invites addressed to the account",
		Example: "  mu invites list\n  mu invites list --json | jq -r '.[].tenant_name'",
		Args:    exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.listFrom(cmd, ndjson, fetchMyInvites, pendingInviteColumns, pendingInviteRow)
		},
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func newInvitesAcceptCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "accept ID",
		Short:   "Join the tenant of an invite and print the tenant id",
		Example: "  mu invites accept 7c9e6679-7425-40de-944b-e07fc1f90ae7\n  mu config set tenant \"$(mu invites accept 7c9e6679-7425-40de-944b-e07fc1f90ae7)\"",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.objectFrom(cmd, acceptInvite(args[0]), app.printJoinedTenant)
		},
	}
}

func newInvitesDeclineCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "decline ID",
		Short:   "Decline an invite addressed to the account",
		Example: "  mu invites decline 7c9e6679-7425-40de-944b-e07fc1f90ae7",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			call := func(ctx context.Context, svc Services) error { return svc.Invites.DeclineInvite(ctx, id) }
			return app.callServer(cmd, call, "declined invite "+id)
		},
	}
}

func fetchMyInvites(ctx context.Context, svc Services) (io.ReadCloser, error) {
	return svc.Invites.ListMyInvites(ctx)
}

func acceptInvite(id string) bodyFetch {
	return func(ctx context.Context, svc Services) (io.ReadCloser, error) {
		return svc.Invites.AcceptInvite(ctx, id)
	}
}

func pendingInviteRow(raw json.RawMessage) ([]string, error) {
	invite, err := decodeRecord[schema.PendingInvite](raw)
	if err != nil {
		return nil, err
	}
	return []string{invite.ID, invite.Role, shortTime(invite.ExpiresAt), invite.TenantID, invite.TenantName}, nil
}

// printJoinedTenant prints the id of the joined tenant and confirms on stderr.
func (a *App) printJoinedTenant(p *output.Printer, raw json.RawMessage) error {
	tenant, err := decodeRecord[schema.Tenant](raw)
	if err != nil {
		return err
	}
	a.confirm("joined tenant " + tenant.Name + " as " + tenant.Role)
	return p.Line(tenant.ID)
}

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// VaultMembership lists, adds and removes the members of a vault.
type VaultMembership interface {
	ListVaultMembers(ctx context.Context, id string) (io.ReadCloser, error)
	AddVaultMember(ctx context.Context, id, userID string) (io.ReadCloser, error)
	RemoveVaultMember(ctx context.Context, id, userID string) error
}

func newVaultMembersCmd(app *App) *cobra.Command {
	cmd := newGroup("members", "Manage who belongs to a vault", "  mu vaults members list 3b241101-e2bb-4255-8caf-4136c566a962")
	cmd.Long = requiresServer(featureVaults, "List, add and remove the members of a vault. Only accounts of the tenant can join.")
	cmd.AddCommand(newVaultMembersListCmd(app), newVaultMembersAddCmd(app), newVaultMembersRemoveCmd(app))
	return cmd
}

func newVaultMembersListCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:     "list ID",
		Short:   "List the members of a vault with their tenant role",
		Long:    requiresServer(featureVaults, "List the members of the vault ID by email, with the role of each one in the tenant."),
		Example: "  mu vaults members list 3b241101-e2bb-4255-8caf-4136c566a962\n  mu vaults members list 3b241101-e2bb-4255-8caf-4136c566a962 --json | jq -r '.[].email'",
		Args:    exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listVaultMembers(cmd.Context(), args[0], ndjson)
		},
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func newVaultMembersAddCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "add ID USER_ID",
		Short:   "Add an account of the tenant to a vault",
		Long:    requiresServer(featureVaults, "Add the account USER_ID, already a member of the tenant, to the vault ID."),
		Example: "  mu vaults members add 3b241101-e2bb-4255-8caf-4136c566a962 9f0c6a52-6d1e-4b8a-a0c2-55e2f1d3c7b4",
		Args:    exactArgs("ID", "USER_ID"),
		RunE:    app.addVaultMember,
	}
}

func newVaultMembersRemoveCmd(app *App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "remove ID USER_ID",
		Short:   "Remove an account from a vault",
		Long:    requiresServer(featureVaults, "Remove the account USER_ID from the vault ID. Removing your own account leaves the vault."),
		Example: "  mu vaults members remove 3b241101-e2bb-4255-8caf-4136c566a962 9f0c6a52-6d1e-4b8a-a0c2-55e2f1d3c7b4 --yes",
		Args:    exactArgs("ID", "USER_ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.removeVaultMember(cmd.Context(), args[0], args[1], yes)
		},
	}
	addYesFlag(cmd, &yes, "the account loses access to the notes of the vault")
	return cmd
}

var vaultMemberColumns = []output.Column{{Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "EMAIL", Width: 32}, {Name: "JOINED"}}

func vaultMemberRow(raw json.RawMessage) ([]string, error) {
	member, err := decodeRecord[schema.VaultMember](raw)
	return []string{member.ID, member.Role, member.Email, member.JoinedAt}, err
}

func printVaultMember(p *output.Printer, raw json.RawMessage) error {
	member, err := decodeRecord[schema.VaultMember](raw)
	if err != nil {
		return err
	}
	return p.Fields(
		output.Field{Key: "id", Value: member.ID}, output.Field{Key: "email", Value: member.Email},
		output.Field{Key: "role", Value: member.Role}, output.Field{Key: "joined", Value: member.JoinedAt},
	)
}

func (a *App) listVaultMembers(ctx context.Context, id string, ndjson bool) error {
	r, err := a.renderer(ndjson)
	if err != nil {
		return err
	}
	body, err := a.openFeature(ctx, featureVaults, func(svc Services) (io.ReadCloser, error) {
		return svc.VaultMembers.ListVaultMembers(ctx, id)
	})
	if err != nil {
		return err
	}
	return r.list(body, vaultMemberColumns, vaultMemberRow)
}

func (a *App) addVaultMember(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	return a.showFeature(ctx, featureVaults, printVaultMember, func(svc Services) (io.ReadCloser, error) {
		return svc.VaultMembers.AddVaultMember(ctx, args[0], args[1])
	})
}

func (a *App) removeVaultMember(ctx context.Context, id, userID string, yes bool) error {
	if err := requireYes(yes, "remove from vault "+id+" the account", userID); err != nil {
		return err
	}
	remove := func(svc Services) error { return svc.VaultMembers.RemoveVaultMember(ctx, id, userID) }
	if err := a.execFeature(ctx, featureVaults, remove); err != nil {
		return err
	}
	a.confirmDone(fmt.Sprintf("removed account %q from vault %q", userID, id))
	return nil
}

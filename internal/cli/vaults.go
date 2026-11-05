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

// VaultAdmin lists, creates, renames and deletes the vaults of the tenant.
type VaultAdmin interface {
	ListVaults(ctx context.Context) (io.ReadCloser, error)
	CreateVault(ctx context.Context, name string) (io.ReadCloser, error)
	RenameVault(ctx context.Context, id, name string) (io.ReadCloser, error)
	DeleteVault(ctx context.Context, id string) error
}

const vaultsLong = `List, create, rename and delete the vaults of the tenant, manage who belongs
to each one, and publish a vault for reading without an account.`

func newVaultsCmd(app *App) *cobra.Command {
	cmd := newGroup("vaults", "Manage vaults, their members and their publication", "  mu vaults list\n  mu vaults members list 3b241101-e2bb-4255-8caf-4136c566a962")
	cmd.Long = requiresServer(featureVaults, vaultsLong) + " Publishing requires " + featurePublication + " too."
	cmd.AddCommand(newVaultsListCmd(app), newVaultsCreateCmd(app), newVaultsRenameCmd(app), newVaultsDeleteCmd(app), newVaultMembersCmd(app),
		newVaultsPublicationCmd(app), newVaultsPublishCmd(app), newVaultsUnpublishCmd(app))
	return cmd
}

func newVaultsListCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List the vaults the account belongs to",
		Long:    requiresServer(featureVaults, "List the vaults of the tenant that the account is a member of, by name."),
		Example: "  mu vaults list\n  mu vaults list --json | jq -r '.[].id'",
		Args:    exactArgs(),
		RunE:    func(cmd *cobra.Command, _ []string) error { return app.listVaults(cmd.Context(), ndjson) },
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func newVaultsCreateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "create NAME",
		Short:   "Create a vault and print its id",
		Long:    requiresServer(featureVaults, "Create a vault named NAME (1 to 100 characters) and print its id. The account becomes its first member."),
		Example: "  mu vaults create Research\n  mu config set vault \"$(mu vaults create Research)\"",
		Args:    exactArgs("NAME"),
		RunE:    app.createVault,
	}
}

func newVaultsRenameCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:     "rename ID NAME",
		Short:   "Rename a vault",
		Long:    requiresServer(featureVaults, "Rename the vault ID to NAME (1 to 100 characters)."),
		Example: "  mu vaults rename 3b241101-e2bb-4255-8caf-4136c566a962 Archive",
		Args:    exactArgs("ID", "NAME"),
		RunE:    app.renameVault,
	}
}

func newVaultsDeleteCmd(app *App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:     "delete ID",
		Short:   "Delete a vault with all its notes",
		Long:    requiresServer(featureVaults, "Delete the vault ID, its notes and everything derived from them. There is no undo."),
		Example: "  mu vaults delete 3b241101-e2bb-4255-8caf-4136c566a962 --yes",
		Args:    exactArgs("ID"),
		RunE:    func(cmd *cobra.Command, args []string) error { return app.deleteVault(cmd.Context(), args[0], yes) },
	}
	addYesFlag(cmd, &yes, "the vault, its notes and everything derived from them are deleted for good")
	return cmd
}

var vaultColumns = []output.Column{{Name: "ID", Width: 36}, {Name: "NAME", Width: 30}, {Name: "CREATED"}}

func vaultRow(raw json.RawMessage) ([]string, error) {
	vault, err := decodeRecord[schema.Vault](raw)
	return []string{vault.ID, vault.Name, vault.CreatedAt}, err
}

func printVault(p *output.Printer, raw json.RawMessage) error {
	vault, err := decodeRecord[schema.Vault](raw)
	if err != nil {
		return err
	}
	return p.Fields(output.Field{Key: "id", Value: vault.ID}, output.Field{Key: "name", Value: vault.Name}, output.Field{Key: "created", Value: vault.CreatedAt})
}

func printVaultID(p *output.Printer, raw json.RawMessage) error {
	vault, err := decodeRecord[schema.Vault](raw)
	if err != nil {
		return err
	}
	return p.Line(vault.ID)
}

func (a *App) listVaults(ctx context.Context, ndjson bool) error {
	r, err := a.renderer(ndjson)
	if err != nil {
		return err
	}
	body, err := a.openFeature(ctx, featureVaults, func(svc Services) (io.ReadCloser, error) { return svc.Vaults.ListVaults(ctx) })
	if err != nil {
		return err
	}
	return r.list(body, vaultColumns, vaultRow)
}

func (a *App) createVault(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	return a.showFeature(ctx, featureVaults, printVaultID, func(svc Services) (io.ReadCloser, error) {
		return svc.Vaults.CreateVault(ctx, args[0])
	})
}

func (a *App) renameVault(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	return a.showFeature(ctx, featureVaults, printVault, func(svc Services) (io.ReadCloser, error) {
		return svc.Vaults.RenameVault(ctx, args[0], args[1])
	})
}

func (a *App) deleteVault(ctx context.Context, id string, yes bool) error {
	if err := requireYes(yes, "delete vault", id); err != nil {
		return err
	}
	if err := a.execFeature(ctx, featureVaults, func(svc Services) error { return svc.Vaults.DeleteVault(ctx, id) }); err != nil {
		return err
	}
	a.inform(fmt.Sprintf("deleted vault %q", id))
	return nil
}

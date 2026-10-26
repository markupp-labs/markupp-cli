package cli

import (
	"context"
	"encoding/json"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// TenantAdmin lists and creates the tenants of the account. *client.Client
// satisfies it.
type TenantAdmin interface {
	ListMyTenants(ctx context.Context) (io.ReadCloser, error)
	CreateTenant(ctx context.Context, name string) (io.ReadCloser, error)
}

const tenantsLong = `List and create the tenants of the account.

The ids are the values of --tenant and ` + config.EnvTenant + `, needed when the account is in
several tenants. Deleting a tenant is only available in the web panel.`

var tenantColumns = []output.Column{
	{Name: "ID", Width: 36}, {Name: "ROLE", Width: 6}, {Name: "NOTES", Width: 5},
	{Name: "LAST_NOTE_AT", Width: timeColumn}, {Name: "NAME"},
}

func newTenantsCmd(app *App) *cobra.Command {
	cmd := newGroup("tenants", "List and create the tenants of the account", "  mu tenants list\n  mu tenants create \"Acme Research\"")
	cmd.Long = tenantsLong
	cmd.AddCommand(newTenantsListCmd(app), newTenantsCreateCmd(app))
	return cmd
}

func newTenantsListCmd(app *App) *cobra.Command {
	var ndjson bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the tenants of the account, with the role and notes read in each",
		Example: `  mu tenants list
  mu tenants list --json | jq -r '.[] | select(.role == "owner") | .id'`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.listFrom(cmd, ndjson, fetchMyTenants, tenantColumns, myTenantRow)
		},
	}
	addNDJSONFlag(cmd, &ndjson)
	return cmd
}

func newTenantsCreateCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "create NAME",
		Short: "Create a tenant owned by the account and print its id",
		Long: `Create a tenant owned by the account, with a vault of its own, and print its id.
The name has at most 100 characters.

` + adminScopeNote,
		Example: `  mu tenants create "Acme Research"
  mu config set tenant "$(mu tenants create "Acme Research")"`,
		Args: exactArgs("NAME"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.objectFrom(cmd, createTenant(args[0]), printTenantID)
		},
	}
}

func fetchMyTenants(ctx context.Context, svc Services) (io.ReadCloser, error) {
	return svc.Tenants.ListMyTenants(ctx)
}

func createTenant(name string) bodyFetch {
	return func(ctx context.Context, svc Services) (io.ReadCloser, error) {
		return svc.Tenants.CreateTenant(ctx, name)
	}
}

func myTenantRow(raw json.RawMessage) ([]string, error) {
	var tenant schema.MyTenant
	if err := json.Unmarshal(raw, &tenant); err != nil {
		return nil, malformedRecord(raw, err)
	}
	lastNote := ""
	if tenant.LastNoteAt != nil {
		lastNote = shortTime(*tenant.LastNoteAt)
	}
	return []string{tenant.ID, tenant.Role, strconv.Itoa(tenant.NotesCount), lastNote, tenant.Name}, nil
}

// printTenantID prints the id of a TenantResponse, for the caller to pipe.
func printTenantID(p *output.Printer, raw json.RawMessage) error {
	var tenant schema.Tenant
	if err := json.Unmarshal(raw, &tenant); err != nil {
		return malformedRecord(raw, err)
	}
	return p.Line(tenant.ID)
}

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// flagError makes every flag parsing error a usage failure naming the command.
func flagError(cmd *cobra.Command, err error) error {
	return failure.Wrap(failure.Usage, cmd.CommandPath(), err)
}

// exactArgs accepts exactly the named positional arguments, and names them and what
// was given otherwise.
func exactArgs(names ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == len(names) {
			return nil
		}
		expected := "no arguments"
		if len(names) > 0 {
			expected = fmt.Sprintf("%d arguments (%s)", len(names), strings.Join(names, " "))
		}
		message := fmt.Sprintf("expected %s, got %d: %q", expected, len(args), args)
		return failure.New(failure.Usage, cmd.CommandPath(), message)
	}
}

// newGroup returns a command that only holds subcommands: alone it prints its help,
// with an unknown subcommand it is a usage error.
func newGroup(use, short, example string) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Short:   short,
		Example: example,
		Args:    cobra.ArbitraryArgs,
		RunE:    runGroup,
	}
}

func runGroup(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	message := fmt.Sprintf("unknown command %q, expected one of: %s", args[0], subcommandNames(cmd))
	return failure.New(failure.Usage, cmd.CommandPath(), message)
}

func subcommandNames(cmd *cobra.Command) string {
	var names []string
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	return strings.Join(names, ", ")
}

// listFlags are the flags of a listing: streaming output and pagination.
type listFlags struct {
	ndjson bool
	limit  int
	offset int
}

func addListFlags(cmd *cobra.Command, f *listFlags) {
	addNDJSONFlag(cmd, &f.ndjson)
	flags := cmd.Flags()
	flags.IntVar(&f.limit, "limit", 0, fmt.Sprintf("results per page (server default %d, values above %d are clamped)", client.DefaultLimit, client.MaxLimit))
	flags.IntVar(&f.offset, "offset", 0, fmt.Sprintf("results to skip (values above %d are clamped)", client.MaxOffset))
}

// addNDJSONFlag adds --ndjson alone, for a listing the server does not paginate.
func addNDJSONFlag(cmd *cobra.Command, ndjson *bool) {
	cmd.Flags().BoolVar(ndjson, "ndjson", false, "print one JSON object per line, as the results stream in")
}

// page validates the pagination flags. The server clamps large values itself.
func (f listFlags) page() (client.Page, error) {
	if f.limit < 0 {
		return client.Page{}, failure.New(failure.Usage, "--limit", fmt.Sprintf("invalid value %d, expected 1 to %d", f.limit, client.MaxLimit))
	}
	if f.offset < 0 {
		return client.Page{}, failure.New(failure.Usage, "--offset", fmt.Sprintf("invalid value %d, expected 0 to %d", f.offset, client.MaxOffset))
	}
	return client.Page{Limit: f.limit, Offset: f.offset}, nil
}

// vaultFlags select the vault a command acts on: one of the account (--vault) or a
// public one by its slug (--public).
type vaultFlags struct {
	vault  string
	public string
}

// vaultTarget is the vault after resolution: an id, or the slug of a public vault.
type vaultTarget struct {
	ID         string
	PublicSlug string
}

// addVaultFlags adds --vault, and --public when the command can read a public vault.
func addVaultFlags(cmd *cobra.Command, f *vaultFlags, allowPublic bool) {
	cmd.Flags().StringVar(&f.vault, "vault", "", "vault id (overrides "+config.EnvVault+" and the config file)")
	if allowPublic {
		addPublicFlag(cmd, &f.public)
	}
}

// addPublicFlag adds --public alone, for a command that takes no vault id.
func addPublicFlag(cmd *cobra.Command, slug *string) {
	cmd.Flags().StringVar(slug, "public", "", "read the public vault with this slug, without credentials (requires a server with "+featurePublication+")")
}

// exclusive refuses --vault and --public together.
func (f vaultFlags) exclusive() error {
	if f.public == "" || f.vault == "" {
		return nil
	}
	message := fmt.Sprintf("--vault %q and --public %q are exclusive, expected one of them", f.vault, f.public)
	return failure.New(failure.Usage, "", message)
}

// resolveVault applies --public, then --vault > MARKUPP_VAULT > config file.
func (a *App) resolveVault(f vaultFlags) (vaultTarget, error) {
	if err := f.exclusive(); err != nil {
		return vaultTarget{}, err
	}
	if f.public != "" {
		return vaultTarget{PublicSlug: f.public}, nil
	}
	resolved, _, err := a.resolve(f.vault)
	if err != nil {
		return vaultTarget{}, err
	}
	if resolved.Vault.Value == "" {
		return vaultTarget{}, errNoVault()
	}
	return vaultTarget{ID: resolved.Vault.Value}, nil
}

func errNoVault() error {
	return failure.New(failure.Usage, "", "no vault selected").
		WithHint("pass --vault ID, set " + config.EnvVault + ", or run mu config set vault ID")
}

// addYesFlag adds --yes to a destructive command; consequence says what is lost.
func addYesFlag(cmd *cobra.Command, yes *bool, consequence string) {
	cmd.Flags().BoolVar(yes, "yes", false, "confirm the operation: "+consequence)
}

// requireYes refuses a destructive action unless --yes was given.
func requireYes(yes bool, action, subject string) error {
	if yes {
		return nil
	}
	return failure.New(failure.Usage, subject, fmt.Sprintf("refusing to %s %q without --yes", action, subject))
}

package cli

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestPageFromFlags(t *testing.T) {
	got, err := listFlags{limit: 5, offset: 20}.page()
	if err != nil || got != (client.Page{Limit: 5, Offset: 20}) {
		t.Errorf("page = %+v, %v", got, err)
	}
	for _, bad := range []listFlags{{limit: -1}, {offset: -3}} {
		_, err := bad.page()
		if failure.KindOf(err) != failure.Usage || !strings.Contains(err.Error(), "-") {
			t.Errorf("page(%+v) error = %v, want usage naming the value", bad, err)
		}
	}
}

func TestRequireYes(t *testing.T) {
	if err := requireYes(true, "delete note", "n1"); err != nil {
		t.Errorf("requireYes(true) = %v", err)
	}
	err := requireYes(false, "delete note", "n1")
	if failure.KindOf(err) != failure.Usage || !strings.Contains(err.Error(), "--yes") || !strings.Contains(err.Error(), "n1") {
		t.Errorf("requireYes(false) = %v, want usage naming --yes and the subject", err)
	}
}

func TestResolveVaultPrecedence(t *testing.T) {
	mu := newTestMu()
	mu.store.settings.Vault = "v-file"
	app := newApp(mu.deps(nil, nil))
	if got, _ := app.resolveVault(vaultFlags{}); got.ID != "v-file" {
		t.Errorf("vault from file = %+v", got)
	}
	mu.env["MARKUPP_VAULT"] = "v-env"
	if got, _ := app.resolveVault(vaultFlags{}); got.ID != "v-env" {
		t.Errorf("vault from env = %+v", got)
	}
	if got, _ := app.resolveVault(vaultFlags{vault: "v-flag"}); got.ID != "v-flag" {
		t.Errorf("vault from flag = %+v", got)
	}
	if got, _ := app.resolveVault(vaultFlags{public: "my-slug"}); got != (vaultTarget{PublicSlug: "my-slug"}) {
		t.Errorf("public target = %+v", got)
	}
}

func TestResolveVaultErrors(t *testing.T) {
	app := newApp(newTestMu().deps(nil, nil))
	_, err := app.resolveVault(vaultFlags{})
	if failure.KindOf(err) != failure.Usage || !strings.Contains(failure.HintOf(err), "--vault") {
		t.Errorf("no vault error = %v hint %q, want usage hinting --vault", err, failure.HintOf(err))
	}
	_, err = app.resolveVault(vaultFlags{vault: "v1", public: "s"})
	if failure.KindOf(err) != failure.Usage {
		t.Errorf("--vault with --public error = %v, want usage", err)
	}
}

func TestVerboseLoggerWritesJSONToStderr(t *testing.T) {
	var stderr bytes.Buffer
	newLogger(true, &stderr).Debug("http request", slog.String("path", "/v1/notes"))
	if !strings.Contains(stderr.String(), `"path":"/v1/notes"`) {
		t.Errorf("verbose log = %q, want a JSON record", stderr.String())
	}
	stderr.Reset()
	newLogger(false, &stderr).Debug("hidden")
	if stderr.Len() != 0 {
		t.Errorf("quiet logger wrote %q", stderr.String())
	}
}

func parsedCommand(t *testing.T, register func(*cobra.Command), args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "x"}
	register(cmd)
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags(%q): %v", args, err)
	}
	return cmd
}

func TestAddListFlagsBindsPaginationAndNDJSON(t *testing.T) {
	var f listFlags
	cmd := parsedCommand(t, func(c *cobra.Command) { addListFlags(c, &f) }, "--ndjson", "--limit", "5", "--offset", "7")
	if f != (listFlags{ndjson: true, limit: 5, offset: 7}) {
		t.Errorf("flags = %+v", f)
	}
	if usage := cmd.Flags().Lookup("limit").Usage; !strings.Contains(usage, "100") {
		t.Errorf("--limit help %q does not document the cap", usage)
	}
}

func TestAddVaultFlagsOffersPublicOnlyWhenAllowed(t *testing.T) {
	var f vaultFlags
	parsedCommand(t, func(c *cobra.Command) { addVaultFlags(c, &f, true) }, "--vault", "v1", "--public", "s")
	if f != (vaultFlags{vault: "v1", public: "s"}) {
		t.Errorf("flags = %+v", f)
	}
	private := &cobra.Command{Use: "x"}
	addVaultFlags(private, &vaultFlags{}, false)
	if private.Flags().Lookup("public") != nil {
		t.Error("--public registered on a command that cannot read public vaults")
	}
}

func TestAddYesFlagSaysWhatIsLost(t *testing.T) {
	var yes bool
	cmd := parsedCommand(t, func(c *cobra.Command) { addYesFlag(c, &yes, "the note is deleted for good") }, "--yes")
	if !yes {
		t.Error("--yes did not set the flag")
	}
	if usage := cmd.Flags().Lookup("yes").Usage; !strings.Contains(usage, "deleted for good") {
		t.Errorf("--yes help %q does not say what is lost", usage)
	}
}

func TestExactArgsNamesWhatWasGiven(t *testing.T) {
	cmd := &cobra.Command{Use: "get"}
	err := exactArgs("ID")(cmd, []string{"a", "b"})
	if failure.KindOf(err) != failure.Usage || !strings.Contains(err.Error(), "ID") || !strings.Contains(err.Error(), `"b"`) {
		t.Errorf("exactArgs error = %v", err)
	}
	if err := exactArgs()(cmd, nil); err != nil {
		t.Errorf("exactArgs() with no args = %v", err)
	}
}

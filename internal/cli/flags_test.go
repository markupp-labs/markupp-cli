package cli

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

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

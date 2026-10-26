package cli

import (
	"context"

	"github.com/spf13/cobra"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestRootHelpFitsOneScreen(t *testing.T) {
	run := newTestMu().run(t, "--help")
	run.expectCode(t, 0)
	if lines := strings.Count(run.stdout, "\n"); lines >= 40 {
		t.Errorf("mu --help has %d lines, the contract is under 40", lines)
	}
	outputtest.Golden(t, "root_help", []byte(run.stdout))
}

func TestBareMuPrintsHelp(t *testing.T) {
	run := newTestMu().run(t)
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "Available Commands") {
		t.Errorf("bare mu printed %q, want the help", run.stdout)
	}
}

func TestVersionFlag(t *testing.T) {
	run := newTestMu().run(t, "--version")
	run.expectCode(t, 0)
	if run.stdout != "mu 1.2.3\n" {
		t.Errorf("--version printed %q, want %q", run.stdout, "mu 1.2.3\n")
	}
}

func TestUnknownCommandIsAUsageError(t *testing.T) {
	run := newTestMu().run(t, "frobnicate")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, `"frobnicate"`) || run.stdout != "" {
		t.Errorf("stderr %q should name the command, stdout %q should be empty", run.stderr, run.stdout)
	}
}

func TestUnknownFlagIsAUsageError(t *testing.T) {
	run := newTestMu().run(t, "config", "show", "--bogus")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "--bogus") || !strings.Contains(run.stderr, "mu config show --help") {
		t.Errorf("stderr %q should name the flag and point to the help", run.stderr)
	}
}

func TestEveryCommandHasAnExample(t *testing.T) {
	app := newApp(newTestMu().deps(nil, nil))
	var missing []string
	walk(newRoot(app), func(path string, example string) {
		if strings.TrimSpace(example) == "" {
			missing = append(missing, path)
		}
	})
	if len(missing) > 0 {
		t.Errorf("commands without Example: %v", missing)
	}
}

func TestFirstRunWithoutURLExplainsHowToConfigure(t *testing.T) {
	mu := newTestMu()
	_, err := newApp(mu.deps(nil, nil)).connect(context.Background(), Target{})
	if failure.KindOf(err) != failure.Usage {
		t.Fatalf("connect without url error = %v, want usage", err)
	}
	hint := failure.HintOf(err)
	for _, want := range []string{"mu config set url", "MARKUPP_URL", "--url"} {
		if !strings.Contains(hint, want) {
			t.Errorf("hint %q does not mention %s", hint, want)
		}
	}
	if len(mu.server.connections) != 0 {
		t.Error("connect reached the server factory without a URL")
	}
}

func TestConnectPassesResolvedSettings(t *testing.T) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://env.example.com"
	mu.store.settings.Tenant = "t-file"
	app := newApp(mu.deps(nil, nil))
	if _, err := app.connect(context.Background(), Target{Public: true}); err != nil {
		t.Fatalf("connect: %v", err)
	}
	conn := mu.server.connections[0]
	if conn.Resolved.URL.Value != "https://env.example.com" || conn.Resolved.Tenant.Value != "t-file" || !conn.Public {
		t.Errorf("connection = %+v", conn)
	}
	if conn.UserAgent != "mu/1.2.3" {
		t.Errorf("user agent = %q, want mu/1.2.3", conn.UserAgent)
	}
}

func TestErrorReportCarriesHintAndIsSanitized(t *testing.T) {
	mu := newTestMu()
	run := mu.run(t, "config", "set", "url", "ftp://\x1b[31mx")
	run.expectCode(t, 2)
	if strings.Contains(run.stderr, "\x1b") {
		t.Errorf("stderr %q carries an escape character", run.stderr)
	}
	if !strings.HasPrefix(run.stderr, "Error: ") {
		t.Errorf("stderr %q does not start with Error:", run.stderr)
	}
}

// walk calls visit with the path and example of cmd and every command under it, the
// generated help and completion commands excepted.
func walk(cmd *cobra.Command, visit func(path, example string)) {
	if cmd.Name() == "help" || cmd.Name() == "completion" {
		return
	}
	visit(cmd.CommandPath(), cmd.Example)
	for _, child := range cmd.Commands() {
		walk(child, visit)
	}
}

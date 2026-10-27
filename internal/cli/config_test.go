package cli

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func configuredMu() *testMu {
	mu := newTestMu()
	mu.store.settings = config.Settings{
		URL: "http://localhost:8080", Vault: "v-file",
		Auth: &config.Credential{
			Issuer: "http://localhost:8080", AccessToken: "secret-access", RefreshToken: "secret-refresh",
			ExpiresAt: time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	mu.env["MARKUPP_TENANT"] = "t-env"
	mu.env["MARKUPP_TOKEN"] = "secret-env-token"
	return mu
}

func TestConfigShowText(t *testing.T) {
	run := configuredMu().run(t, "config", "show", "--url", "https://flag.example.com")
	run.expectCode(t, 0)
	if strings.Contains(run.stdout, "secret") {
		t.Fatalf("config show leaked a token: %s", run.stdout)
	}
	outputtest.Golden(t, "config_show", []byte(run.stdout))
}

func TestConfigShowJSON(t *testing.T) {
	run := configuredMu().run(t, "config", "show", "--json")
	run.expectCode(t, 0)
	if strings.Contains(run.stdout, "secret") {
		t.Fatalf("config show --json leaked a token: %s", run.stdout)
	}
	outputtest.Golden(t, "config_show_json", []byte(run.stdout))
}

func TestConfigShowWithNothingConfigured(t *testing.T) {
	run := newTestMu().run(t, "config", "show")
	run.expectCode(t, 0)
	outputtest.Golden(t, "config_show_empty", []byte(run.stdout))
}

func TestConfigShowFailsWhenTheFileIsBroken(t *testing.T) {
	mu := newTestMu()
	mu.store.loadErr = errors.New("config file /x is not valid")
	run := mu.run(t, "config", "show")
	run.expectCode(t, 1)
	if !strings.Contains(run.stderr, "/x") {
		t.Errorf("stderr %q does not name the file", run.stderr)
	}
}

func TestConfigSetNormalizesAndStores(t *testing.T) {
	mu := newTestMu()
	run := mu.run(t, "config", "set", "url", "https://markupp.example.com/")
	run.expectCode(t, 0)
	if run.stdout != "" {
		t.Errorf("config set wrote %q to stdout, want nothing", run.stdout)
	}
	if mu.store.settings.URL != "https://markupp.example.com" {
		t.Errorf("stored url = %q", mu.store.settings.URL)
	}
}

func TestConfigSetRejectsBadInput(t *testing.T) {
	cases := map[string][]string{
		"unknown key":  {"config", "set", "token", "x"},
		"invalid url":  {"config", "set", "url", "markupp.example.com"},
		"missing arg":  {"config", "set", "url"},
		"extra arg":    {"config", "set", "url", "https://a", "b"},
		"unset no key": {"config", "unset"},
	}
	for name, args := range cases {
		mu := newTestMu()
		run := mu.run(t, args...)
		if run.code != 2 || mu.store.updates != 0 {
			t.Errorf("%s: exit %d, %d updates, want exit 2 and no update; stderr %q", name, run.code, mu.store.updates, run.stderr)
		}
	}
}

func TestConfigUnsetClearsTheKey(t *testing.T) {
	mu := configuredMu()
	mu.run(t, "config", "unset", "vault").expectCode(t, 0)
	if mu.store.settings.Vault != "" || mu.store.settings.URL == "" {
		t.Errorf("settings after unset vault = %+v", mu.store.settings)
	}
}

func TestConfigPathPrintsTheFile(t *testing.T) {
	run := newTestMu().run(t, "config", "path")
	run.expectCode(t, 0)
	if run.stdout != "/home/u/.config/markupp/cli.json\n" {
		t.Errorf("config path printed %q", run.stdout)
	}
}

func TestConfigUnknownSubcommandIsAUsageError(t *testing.T) {
	run := newTestMu().run(t, "config", "frob")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, `"frob"`) {
		t.Errorf("stderr %q does not name the subcommand", run.stderr)
	}
}

func TestConfigGroupWithoutSubcommandPrintsHelp(t *testing.T) {
	run := newTestMu().run(t, "config")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "mu config show") {
		t.Errorf("mu config printed %q, want its help", run.stdout)
	}
}

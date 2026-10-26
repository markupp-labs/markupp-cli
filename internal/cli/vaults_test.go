package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const (
	vaultsBody  = `[{"id":"3b241101-e2bb-4255-8caf-4136c566a962","name":"Work","created_at":"2026-05-01T12:00:00Z"},{"id":"v2","name":"A vault with a name longer than the column","created_at":"2026-05-02T08:30:00.5Z"}]`
	vaultBody   = `{"id":"3b241101-e2bb-4255-8caf-4136c566a962","name":"Work","created_at":"2026-05-01T12:00:00Z"}`
	vaultMembersBody = `[{"id":"u1","email":"ana@example.com","role":"owner","joined_at":"2026-05-01T12:00:00Z"},{"id":"u2","email":"bo@example.com","role":"viewer","joined_at":"2026-05-03T09:00:00Z"}]`
	memberBody  = `{"id":"u2","email":"bo@example.com","role":"viewer","joined_at":"2026-05-03T09:00:00Z"}`
)

// vaultsMu is a test mu with a server URL and fake vault ports.
func vaultsMu() (*testMu, *fakeVaults, *fakeVaultMembers) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	vaults, members := &fakeVaults{}, &fakeVaultMembers{}
	mu.server.services = Services{Vaults: vaults, VaultMembers: members}
	return mu, vaults, members
}

func expectCallNames(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("calls = %q, want %q", got, want)
	}
}

func TestVaultsListPrintsATable(t *testing.T) {
	mu, vaults, _ := vaultsMu()
	vaults.reply = vaultsBody
	run := mu.run(t, "vaults", "list")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_list", []byte(run.stdout))
	run = mu.run(t, "vaults", "list", "--json")
	run.expectCode(t, 0)
	if run.stdout != vaultsBody+"\n" {
		t.Errorf("--json printed %q, want the body unchanged", run.stdout)
	}
}

func TestVaultsCreatePrintsTheID(t *testing.T) {
	mu, vaults, _ := vaultsMu()
	vaults.reply = vaultBody
	run := mu.run(t, "vaults", "create", "Work")
	run.expectCode(t, 0)
	if run.stdout != "3b241101-e2bb-4255-8caf-4136c566a962\n" {
		t.Errorf("create printed %q, want the id", run.stdout)
	}
	expectCallNames(t, vaults.calls, "create Work")
}

func TestVaultsRenamePrintsTheVault(t *testing.T) {
	mu, vaults, _ := vaultsMu()
	vaults.reply = vaultBody
	run := mu.run(t, "vaults", "rename", "v1", "Work")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_rename", []byte(run.stdout))
	expectCallNames(t, vaults.calls, "rename v1 Work")
}

func TestVaultsDeleteRequiresYes(t *testing.T) {
	mu, vaults, _ := vaultsMu()
	run := mu.run(t, "vaults", "delete", "v1")
	run.expectCode(t, 2)
	if len(vaults.calls) != 0 || !strings.Contains(run.stderr, "--yes") {
		t.Errorf("delete without --yes called %q, stderr %q", vaults.calls, run.stderr)
	}
	run = mu.run(t, "vaults", "delete", "v1", "--yes")
	run.expectCode(t, 0)
	expectCallNames(t, vaults.calls, "delete v1")
	if run.stdout != "" || !strings.Contains(run.stderr, `"v1"`) {
		t.Errorf("delete stdout %q, stderr %q, want the confirmation on stderr only", run.stdout, run.stderr)
	}
}

func TestVaultsWithoutTheEndpointSayWhatTheServerLacks(t *testing.T) {
	mu, vaults, _ := vaultsMu()
	vaults.err = errEndpointAbsent()
	run := mu.run(t, "vaults", "list")
	run.expectCode(t, 5)
	if !strings.Contains(run.stderr, "endpoint absent") || !strings.Contains(run.stderr, "requires a server with vault management") {
		t.Errorf("stderr %q should say the endpoint is absent and what the server lacks", run.stderr)
	}
}

func TestVaultMembersListPrintsATable(t *testing.T) {
	mu, _, members := vaultsMu()
	members.reply = vaultMembersBody
	run := mu.run(t, "vaults", "members", "list", "v1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_members_list", []byte(run.stdout))
	expectCallNames(t, members.calls, "list v1")
}

func TestVaultMembersAddPrintsTheMember(t *testing.T) {
	mu, _, members := vaultsMu()
	members.reply = memberBody
	run := mu.run(t, "vaults", "members", "add", "v1", "u2")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_members_add", []byte(run.stdout))
	expectCallNames(t, members.calls, "add v1 u2")
}

func TestVaultMembersRemoveRequiresYes(t *testing.T) {
	mu, _, members := vaultsMu()
	mu.run(t, "vaults", "members", "remove", "v1", "u2").expectCode(t, 2)
	run := mu.run(t, "vaults", "members", "remove", "v1", "u2", "--yes")
	run.expectCode(t, 0)
	expectCallNames(t, members.calls, "remove v1 u2")
	if run.stdout != "" || !strings.Contains(run.stderr, `"u2"`) {
		t.Errorf("remove stdout %q, stderr %q", run.stdout, run.stderr)
	}
}

func TestVaultCommandsNameTheirArguments(t *testing.T) {
	mu, _, _ := vaultsMu()
	run := mu.run(t, "vaults", "rename", "v1")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "ID NAME") {
		t.Errorf("stderr %q should name the expected arguments", run.stderr)
	}
}

func TestVaultCommandsSayWhichServerTheyNeed(t *testing.T) {
	root := newRoot(newApp(newTestMu().deps(nil, nil)))
	for _, path := range [][]string{
		{"vaults"}, {"vaults", "list"}, {"vaults", "create"}, {"vaults", "rename"}, {"vaults", "delete"},
		{"vaults", "members"}, {"vaults", "members", "list"}, {"vaults", "members", "add"}, {"vaults", "members", "remove"},
	} {
		cmd, _, err := root.Find(path)
		if err != nil || cmd.Name() != path[len(path)-1] {
			t.Errorf("command %q not found: %v", path, err)
			continue
		}
		if !strings.Contains(cmd.Long, "Requires a server with vault management") {
			t.Errorf("help of %q does not say which server it needs: %q", path, cmd.Long)
		}
	}
}

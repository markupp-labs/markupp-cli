package cli

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func connectedMu() *testMu {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	return mu
}

func TestConnectVaultReadsAPublicVaultWithoutCredentials(t *testing.T) {
	mu := connectedMu()
	_, target, err := newApp(mu.deps(nil, nil)).connectVault(context.Background(), vaultFlags{public: "docs"})
	if err != nil || target != (vaultTarget{PublicSlug: "docs"}) {
		t.Fatalf("connectVault(--public docs) = %+v, %v", target, err)
	}
	if conn := mu.server.connections[0]; !conn.Public {
		t.Errorf("connection = %+v, want a public one", conn)
	}
}

func TestConnectVaultReadsAnAccountVaultWithCredentials(t *testing.T) {
	mu := connectedMu()
	_, target, err := newApp(mu.deps(nil, nil)).connectVault(context.Background(), vaultFlags{vault: "v1"})
	if err != nil || target != (vaultTarget{ID: "v1"}) {
		t.Fatalf("connectVault(--vault v1) = %+v, %v", target, err)
	}
	if conn := mu.server.connections[0]; conn.Public || conn.Resolved.Vault.Value != "v1" {
		t.Errorf("connection = %+v, want a private one on v1", conn)
	}
}

func TestConnectVaultWithoutAVaultConnectsNothing(t *testing.T) {
	mu := connectedMu()
	_, _, err := newApp(mu.deps(nil, nil)).connectVault(context.Background(), vaultFlags{})
	if failure.KindOf(err) != failure.Usage || len(mu.server.connections) != 0 {
		t.Errorf("connectVault without a vault = %v after %d connections, want usage and none", err, len(mu.server.connections))
	}
}

func TestPublicReadNamesTheMissingFeature(t *testing.T) {
	_, err := publicRead(nil, errEndpointAbsent())
	if !strings.Contains(failure.HintOf(err), "vault publication") || failure.KindOf(err) != failure.NotFound {
		t.Errorf("publicRead error = %v, hint %q", err, failure.HintOf(err))
	}
	other := errors.New("boom")
	if _, err := publicRead(nil, other); !errors.Is(err, other) {
		t.Errorf("publicRead changed an unrelated error: %v", err)
	}
	got, err := publicRead(body("[]"), nil)
	if err != nil || got == nil {
		t.Errorf("publicRead of a body = %v, %v", got, err)
	}
}

func TestPublicFlagSaysWhichServerItNeeds(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	addVaultFlags(cmd, &vaultFlags{}, true)
	if usage := cmd.Flags().Lookup("public").Usage; !strings.Contains(usage, "requires a server with vault publication") {
		t.Errorf("--public help %q does not say which server it needs", usage)
	}
}

func TestServicesWireThePublicReader(t *testing.T) {
	api, err := client.NewPublic(client.Options{BaseURL: "https://markupp.example.com"})
	if err != nil {
		t.Fatalf("NewPublic: %v", err)
	}
	svc := newServices(api)
	if svc.Public == nil || svc.Publication == nil || svc.Vaults == nil || svc.VaultMembers == nil {
		t.Errorf("services = %+v, want every vault port wired", svc)
	}
}

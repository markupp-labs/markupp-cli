package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const myTenantsBody = `[{"id":"5f0c2d1e-8a7b-4c3d-9e1f-2a3b4c5d6e7f","name":"Acme Research","role":"owner","notes_count":42,"last_note_at":"2026-10-20T13:45:12.123456Z"},` +
	`{"id":"t2","name":"Side \u001b[31mproject","role":"viewer","notes_count":0,"last_note_at":null}]`

func TestTenantsListPrintsATable(t *testing.T) {
	mu, fake := newAdminMu(myTenantsBody)
	run := mu.run(t, "tenants", "list")
	run.expectCode(t, 0)
	fake.expectCalls(t, "ListMyTenants")
	outputtest.Golden(t, "tenants_list", []byte(run.stdout))
}

func TestTenantsListJSONIsThePayloadUnchanged(t *testing.T) {
	mu, _ := newAdminMu(myTenantsBody)
	run := mu.run(t, "tenants", "list", "--json")
	run.expectCode(t, 0)
	if run.stdout != myTenantsBody+"\n" && run.stdout != myTenantsBody {
		t.Errorf("--json printed %q, want the body unchanged", run.stdout)
	}
}

func TestTenantsListNDJSONPrintsOneTenantPerLine(t *testing.T) {
	mu, _ := newAdminMu(myTenantsBody)
	run := mu.run(t, "tenants", "list", "--ndjson")
	run.expectCode(t, 0)
	outputtest.Golden(t, "tenants_list_ndjson", []byte(run.stdout))
}

func TestTenantsCreatePrintsTheID(t *testing.T) {
	mu, fake := newAdminMu(`{"id":"t9","name":"Acme","role":"owner"}`)
	run := mu.run(t, "tenants", "create", "Acme")
	run.expectCode(t, 0)
	fake.expectCalls(t, "CreateTenant Acme")
	if run.stdout != "t9\n" {
		t.Errorf("stdout = %q, want the tenant id", run.stdout)
	}
}

func TestTenantsCreateNeedsAName(t *testing.T) {
	mu, fake := newAdminMu("")
	run := mu.run(t, "tenants", "create")
	run.expectCode(t, 2)
	fake.expectCalls(t)
	if !strings.Contains(run.stderr, "NAME") {
		t.Errorf("stderr %q does not name the missing argument", run.stderr)
	}
}

func TestTenantsWithoutServerExplainsHowToConfigure(t *testing.T) {
	mu := newTestMu()
	run := mu.run(t, "tenants", "list")
	run.expectCode(t, 2)
	if !strings.Contains(run.stderr, "mu config set url") {
		t.Errorf("stderr %q does not explain how to configure a server", run.stderr)
	}
}

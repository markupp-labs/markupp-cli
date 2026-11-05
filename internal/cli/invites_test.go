package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

func TestInvitesListPrintsATable(t *testing.T) {
	mu, fake := newAdminMu(`[{"id":"7c9e6679-7425-40de-944b-e07fc1f90ae7","tenant_id":"5f0c2d1e-8a7b-4c3d-9e1f-2a3b4c5d6e7f","tenant_name":"Acme Research","role":"member","expires_at":"2026-11-02T10:00:00Z"}]`)
	run := mu.run(t, "invites", "list")
	run.expectCode(t, 0)
	fake.expectCalls(t, "ListMyInvites")
	outputtest.Golden(t, "invites_list", []byte(run.stdout))
}

func TestInvitesAcceptPrintsTheTenantID(t *testing.T) {
	mu, fake := newAdminMu(`{"id":"t1","name":"Acme","role":"member"}`)
	run := mu.run(t, "invites", "accept", "i1")
	run.expectCode(t, 0)
	fake.expectCalls(t, "AcceptInvite i1")
	if run.stdout != "t1\n" || !strings.Contains(run.stderr, "joined tenant Acme as member") {
		t.Errorf("stdout %q, stderr %q: want the tenant id and a confirmation", run.stdout, run.stderr)
	}
}

func TestInvitesAcceptJSONIsThePayload(t *testing.T) {
	const tenant = `{"id":"t1","name":"Acme","role":"member"}`
	mu, _ := newAdminMu(tenant)
	run := mu.run(t, "invites", "accept", "i1", "--json")
	run.expectCode(t, 0)
	if strings.TrimSpace(run.stdout) != tenant || run.stderr != "" {
		t.Errorf("stdout %q, stderr %q: want the payload alone", run.stdout, run.stderr)
	}
}

func TestInvitesAcceptConflictExitsSix(t *testing.T) {
	mu, fake := newAdminMu("")
	fake.err = failure.New(failure.Conflict, "POST /v1/me/invites/i1/accept", "already a member")
	mu.run(t, "invites", "accept", "i1").expectCode(t, 6)
}

func TestInvitesDeclineConfirmsOnStderr(t *testing.T) {
	mu, fake := newAdminMu("")
	run := mu.run(t, "invites", "decline", "i1")
	run.expectCode(t, 0)
	fake.expectCalls(t, "DeclineInvite i1")
	if run.stdout != "" || !strings.Contains(run.stderr, "declined invite i1") {
		t.Errorf("stdout %q, stderr %q", run.stdout, run.stderr)
	}
}

func TestInvitesNotFoundExitsFive(t *testing.T) {
	mu, fake := newAdminMu("")
	fake.err = failure.New(failure.NotFound, "POST /v1/me/invites/nope/decline", "invite not found")
	run := mu.run(t, "invites", "decline", "nope")
	run.expectCode(t, 5)
	if !strings.Contains(run.stderr, "nope") {
		t.Errorf("stderr %q does not name the invite", run.stderr)
	}
}

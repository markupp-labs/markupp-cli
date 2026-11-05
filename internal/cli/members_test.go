package cli

import (
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const membersBody = `[{"id":"0b7f3c1e-5d2a-4f6b-9c8e-1a2b3c4d5e6f","email":"ana@example.com","role":"owner","joined_at":"2026-01-02T03:04:05Z"},` +
	`{"id":"u2","email":"bruno@example.com","role":"viewer","joined_at":"2026-05-06T07:08:09.5-03:00"}]`

func TestMembersListPrintsATable(t *testing.T) {
	mu, fake := newAdminMu(membersBody)
	run := mu.run(t, "members", "list")
	run.expectCode(t, 0)
	fake.expectCalls(t, "ListMembers")
	outputtest.Golden(t, "members_list", []byte(run.stdout))
}

func TestMembersListUsesTheTenantFlag(t *testing.T) {
	mu, _ := newAdminMu(membersBody)
	mu.run(t, "members", "list", "--tenant", "t7").expectCode(t, 0)
	if got := mu.server.connections[0]; got.Resolved.Tenant.Value != "t7" || got.Public {
		t.Errorf("connection = %+v, want tenant t7 with credentials", got)
	}
}

func TestMembersRoleChangesTheRole(t *testing.T) {
	mu, fake := newAdminMu("")
	run := mu.run(t, "members", "role", "u1", "admin")
	run.expectCode(t, 0)
	fake.expectCalls(t, "ChangeMemberRole u1 admin")
	if run.stdout != "" || !strings.Contains(run.stderr, "u1") {
		t.Errorf("stdout %q should be empty and stderr %q should confirm", run.stdout, run.stderr)
	}
}

func TestMembersRoleRefusesAnUnknownRole(t *testing.T) {
	mu, fake := newAdminMu("")
	run := mu.run(t, "members", "role", "u1", "editor")
	run.expectCode(t, 2)
	fake.expectCalls(t)
	if !strings.Contains(run.stderr, `"editor"`) || !strings.Contains(run.stderr, "owner, admin, member, viewer") {
		t.Errorf("stderr %q should name the role and the expected ones", run.stderr)
	}
}

func TestMembersRemoveRequiresYes(t *testing.T) {
	mu, fake := newAdminMu("")
	run := mu.run(t, "members", "remove", "u1")
	run.expectCode(t, 2)
	fake.expectCalls(t)
	if !strings.Contains(run.stderr, "--yes") || !strings.Contains(run.stderr, `"u1"`) {
		t.Errorf("stderr %q should ask for --yes and name the member", run.stderr)
	}
}

func TestMembersRemoveWithYes(t *testing.T) {
	mu, fake := newAdminMu("")
	mu.run(t, "members", "remove", "u1", "--yes").expectCode(t, 0)
	fake.expectCalls(t, "RemoveMember u1")
}

func TestMembersRemoveHelpSaysWhatIsLost(t *testing.T) {
	run := newTestMu().run(t, "members", "remove", "--help")
	run.expectCode(t, 0)
	if !strings.Contains(run.stdout, "deleted with their notes") {
		t.Errorf("help %q does not say what is lost", run.stdout)
	}
}

func TestMembersInvitePrintsTheInviteAndWarnsAboutTheToken(t *testing.T) {
	mu, fake := newAdminMu(`{"id":"i1","email":"ana@example.com","role":"member","expires_at":"2026-11-02T10:00:00.25Z","created_at":"2026-10-26T10:00:00Z","token":"tok-123"}`)
	run := mu.run(t, "members", "invite", "ana@example.com", "--role", "member")
	run.expectCode(t, 0)
	fake.expectCalls(t, "InviteMember ana@example.com member")
	outputtest.Golden(t, "members_invite", []byte(run.stdout))
	if !strings.Contains(run.stderr, "only this once") {
		t.Errorf("stderr %q does not warn that the token is shown once", run.stderr)
	}
}

func TestMembersInviteWithoutTokenSaysNothingOnStderr(t *testing.T) {
	mu, _ := newAdminMu(`{"id":"i1","email":"ana@example.com","role":"member","expires_at":"2026-11-02T10:00:00Z"}`)
	run := mu.run(t, "members", "invite", "ana@example.com", "--role", "member")
	run.expectCode(t, 0)
	if run.stderr != "" || strings.Contains(run.stdout, "token") {
		t.Errorf("stdout %q, stderr %q: want no token line and no warning", run.stdout, run.stderr)
	}
}

func TestMembersInviteRequiresAValidRole(t *testing.T) {
	for _, args := range [][]string{{"--role", "boss"}, {}} {
		mu, fake := newAdminMu("")
		run := mu.run(t, append([]string{"members", "invite", "ana@example.com"}, args...)...)
		run.expectCode(t, 2)
		fake.expectCalls(t)
		if !strings.Contains(run.stderr, "--role") {
			t.Errorf("args %q: stderr %q does not name --role", args, run.stderr)
		}
	}
}

func TestMembersInvitesPrintsATable(t *testing.T) {
	mu, fake := newAdminMu(`[{"id":"i1","email":"ana@example.com","role":"admin","invited_by":"u1","expires_at":"2026-11-02T10:00:00Z","created_at":"2026-10-26T10:00:00Z"},` +
		`{"id":"i2","email":"bruno@example.com","role":"viewer","expires_at":"2026-10-30T10:00:00Z","accepted_at":"2026-10-25T09:00:00Z","created_at":"2026-10-23T10:00:00Z"}]`)
	run := mu.run(t, "members", "invites")
	run.expectCode(t, 0)
	fake.expectCalls(t, "ListTenantInvites")
	outputtest.Golden(t, "members_invites", []byte(run.stdout))
}

func TestMembersRevokeRequiresYes(t *testing.T) {
	mu, fake := newAdminMu("")
	mu.run(t, "members", "revoke", "i1").expectCode(t, 2)
	fake.expectCalls(t)
	mu.run(t, "members", "revoke", "i1", "--yes").expectCode(t, 0)
	fake.expectCalls(t, "RevokeInvite i1")
}

func TestMembersForbiddenExitsFourWithTheHint(t *testing.T) {
	mu, fake := newAdminMu("")
	fake.err = errFirstPartyOnly
	run := mu.run(t, "members", "list")
	run.expectCode(t, failure.Forbidden.ExitCode())
	if !strings.Contains(run.stderr, "first_party_only") || !strings.Contains(run.stderr, "Hint: this operation is only available in the web panel") {
		t.Errorf("stderr %q should carry the code and the hint", run.stderr)
	}
	if run.stdout != "" {
		t.Errorf("stdout = %q, want nothing on a failure", run.stdout)
	}
}

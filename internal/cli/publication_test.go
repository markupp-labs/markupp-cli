package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const publicationBody = `{"vault_id":"3b241101-e2bb-4255-8caf-4136c566a962","slug":"docs","title":"Markupp docs","published_by":"u1","published_at":"2026-05-01T12:00:00Z"}`

// fakePublisher is the VaultPublisher port.
type fakePublisher struct{ callLog }

func (f *fakePublisher) GetPublication(_ context.Context, id string) (io.ReadCloser, error) {
	return f.open("get", id)
}

func (f *fakePublisher) PublishVault(_ context.Context, id, slug, title string) (io.ReadCloser, error) {
	return f.open("publish", id, slug, title)
}

func (f *fakePublisher) UnpublishVault(_ context.Context, id string) error {
	return f.exec("unpublish", id)
}

func publicationMu() (*testMu, *fakePublisher) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	publisher := &fakePublisher{callLog{reply: publicationBody}}
	mu.server.services = Services{Publication: publisher}
	return mu, publisher
}

func TestVaultsPublicationPrintsThePublication(t *testing.T) {
	mu, publisher := publicationMu()
	run := mu.run(t, "vaults", "publication", "v1")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_publication", []byte(run.stdout))
	expectCallNames(t, publisher.calls, "get v1")
}

func TestVaultsPublishSendsSlugAndTitle(t *testing.T) {
	mu, publisher := publicationMu()
	run := mu.run(t, "vaults", "publish", "v1", "--slug", "docs", "--title", "Markupp docs")
	run.expectCode(t, 0)
	outputtest.Golden(t, "vaults_publication", []byte(run.stdout))
	expectCallNames(t, publisher.calls, "publish v1 docs Markupp docs")
}

func TestVaultsPublishRequiresASlug(t *testing.T) {
	mu, publisher := publicationMu()
	run := mu.run(t, "vaults", "publish", "v1")
	run.expectCode(t, 2)
	if len(publisher.calls) != 0 || !strings.Contains(run.stderr, "--slug") {
		t.Errorf("publish without --slug called %q, stderr %q", publisher.calls, run.stderr)
	}
}

func TestVaultsUnpublishRequiresYes(t *testing.T) {
	mu, publisher := publicationMu()
	mu.run(t, "vaults", "unpublish", "v1").expectCode(t, 2)
	run := mu.run(t, "vaults", "unpublish", "v1", "--yes")
	run.expectCode(t, 0)
	expectCallNames(t, publisher.calls, "unpublish v1")
	if run.stdout != "" || !strings.Contains(run.stderr, `"v1"`) {
		t.Errorf("unpublish stdout %q, stderr %q", run.stdout, run.stderr)
	}
}

func TestVaultsPublicationOfAnUnpublishedVaultIsNotFound(t *testing.T) {
	mu, publisher := publicationMu()
	publisher.err = &failure.Error{Kind: failure.NotFound, Code: "not_published", Message: "vault is not published"}
	run := mu.run(t, "vaults", "publication", "v1")
	run.expectCode(t, 5)
	if strings.Contains(run.stderr, "endpoint absent") || !strings.Contains(run.stderr, "not_published") {
		t.Errorf("stderr %q should carry the server error, not endpoint absent", run.stderr)
	}
}

func TestVaultsPublicationWithoutTheEndpointSaysWhatTheServerLacks(t *testing.T) {
	mu, publisher := publicationMu()
	publisher.err = errEndpointAbsent()
	run := mu.run(t, "vaults", "publish", "v1", "--slug", "docs")
	run.expectCode(t, 5)
	if !strings.Contains(run.stderr, "requires a server with vault publication") {
		t.Errorf("stderr %q should name the missing feature", run.stderr)
	}
}

func TestPublicationCommandsSayWhichServerTheyNeed(t *testing.T) {
	root := newRoot(newApp(newTestMu().deps(nil, nil)))
	for _, name := range []string{"publication", "publish", "unpublish"} {
		cmd, _, err := root.Find([]string{"vaults", name})
		if err != nil || cmd.Name() != name {
			t.Errorf("command vaults %s not found: %v", name, err)
			continue
		}
		if !strings.Contains(cmd.Long, "Requires a server with vault publication") {
			t.Errorf("help of vaults %s does not say which server it needs: %q", name, cmd.Long)
		}
	}
}

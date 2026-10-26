package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
)

const notificationsBody = `{"items":[{"kind":"invite","id":"8d1f4c2e-1b2a-4c3d-9e8f-0a1b2c3d4e5f","tenant_id":"t1","tenant_name":"Acme Research","role":"member","expires_at":"2026-06-01T00:00:00Z"},{"kind":"mention","id":"m1"}],"unread":1}`

// fakeNotifications is the NotificationReader port.
type fakeNotifications struct{ callLog }

func (f *fakeNotifications) Notifications(context.Context) (io.ReadCloser, error) {
	return f.open("feed")
}

func notificationsMu(reply string) (*testMu, *fakeNotifications) {
	mu := newTestMu()
	mu.env["MARKUPP_URL"] = "https://markupp.example.com"
	feed := &fakeNotifications{callLog{reply: reply}}
	mu.server.services = Services{Notifications: feed}
	return mu, feed
}

func TestNotificationsPrintsKnownItemsAndUnreadOnStderr(t *testing.T) {
	mu, feed := notificationsMu(notificationsBody)
	run := mu.run(t, "notifications")
	run.expectCode(t, 0)
	outputtest.Golden(t, "notifications", []byte(run.stdout))
	if run.stderr != "1 unread\n" {
		t.Errorf("stderr = %q, want the unread count", run.stderr)
	}
	expectCalls(t, feed.calls, "feed")
}

func TestNotificationsJSONKeepsTheBodyAndCountsOnStderr(t *testing.T) {
	mu, _ := notificationsMu(notificationsBody)
	run := mu.run(t, "notifications", "--json")
	run.expectCode(t, 0)
	if run.stdout != notificationsBody+"\n" || run.stderr != "1 unread\n" {
		t.Errorf("stdout %q, stderr %q", run.stdout, run.stderr)
	}
}

func TestNotificationsRejectsAMalformedFeed(t *testing.T) {
	mu, _ := notificationsMu(`{"items":"nope"}`)
	run := mu.run(t, "notifications")
	run.expectCode(t, 1)
	if run.stdout != "" || !strings.Contains(run.stderr, "malformed") {
		t.Errorf("stdout %q, stderr %q", run.stdout, run.stderr)
	}
}

func TestNotificationsWithoutTheEndpointSaysWhatTheServerLacks(t *testing.T) {
	mu, feed := notificationsMu("")
	feed.err = errEndpointAbsent()
	run := mu.run(t, "notifications")
	run.expectCode(t, 5)
	if !strings.Contains(run.stderr, "requires a server with notifications") {
		t.Errorf("stderr %q should name the missing feature", run.stderr)
	}
}

func TestNotificationsHelpSaysWhichServerItNeeds(t *testing.T) {
	root := newRoot(newApp(newTestMu().deps(nil, nil)))
	cmd, _, err := root.Find([]string{"notifications"})
	if err != nil || cmd.Name() != "notifications" {
		t.Fatalf("command notifications not found: %v", err)
	}
	if !strings.Contains(cmd.Long, "Requires a server with notifications") {
		t.Errorf("help %q does not say which server it needs", cmd.Long)
	}
}

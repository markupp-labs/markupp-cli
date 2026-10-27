package oauth

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func listen(t *testing.T, state string) *Callback {
	t.Helper()
	callback, err := Listen(context.Background(), state)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	t.Cleanup(func() { _ = callback.Close() })
	return callback
}

func callbackURL(callback *Callback, query url.Values) string {
	return callback.RedirectURI() + "?" + query.Encode()
}

func TestCallbackListensOnTheLoopbackAddress(t *testing.T) {
	callback := listen(t, "s1")
	parsed, err := url.Parse(callback.RedirectURI())
	if err != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.Port() == "" || parsed.Path != CallbackPath {
		t.Errorf("redirect URI = %q, want http://127.0.0.1:<port>/callback", callback.RedirectURI())
	}
}

func TestCallbackReturnsTheCode(t *testing.T) {
	callback := listen(t, "s1")
	if status := browserGet(t, callbackURL(callback, url.Values{"code": {"c1"}, "state": {"s1"}})); status != http.StatusOK {
		t.Errorf("browser status = %d, want 200", status)
	}
	code, err := callback.Wait(context.Background(), time.Second)
	if err != nil || code != "c1" {
		t.Errorf("Wait = %q, %v, want c1", code, err)
	}
}

func TestCallbackIgnoresAnotherState(t *testing.T) {
	callback := listen(t, "s1")
	if status := browserGet(t, callbackURL(callback, url.Values{"code": {"forged"}, "state": {"s2"}})); status != http.StatusBadRequest {
		t.Errorf("browser status for a foreign state = %d, want 400", status)
	}
	if status := browserGet(t, callback.RedirectURI()+"x?state=s1&code=c1"); status != http.StatusNotFound {
		t.Errorf("browser status for another path = %d, want 404", status)
	}
	browserGet(t, callbackURL(callback, url.Values{"code": {"c1"}, "state": {"s1"}}))
	code, err := callback.Wait(context.Background(), time.Second)
	if err != nil || code != "c1" {
		t.Errorf("Wait = %q, %v, want the code of the right state", code, err)
	}
}

func TestCallbackReportsARefusal(t *testing.T) {
	callback := listen(t, "s1")
	browserGet(t, callbackURL(callback, url.Values{"error": {"access_denied"}, "error_description": {"user said no"}, "state": {"s1"}}))
	_, err := callback.Wait(context.Background(), time.Second)
	expectKind(t, err, failure.Unauthenticated)
	if !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("error = %v, want the OAuth error code", err)
	}
}

func TestCallbackWithoutCodeIsAnError(t *testing.T) {
	callback := listen(t, "s1")
	browserGet(t, callbackURL(callback, url.Values{"state": {"s1"}}))
	if _, err := callback.Wait(context.Background(), time.Second); err == nil {
		t.Error("Wait without a code succeeded")
	}
}

func TestCallbackTimesOut(t *testing.T) {
	callback := listen(t, "s1")
	_, err := callback.Wait(context.Background(), 20*time.Millisecond)
	expectKind(t, err, failure.Unauthenticated)
	if !strings.Contains(err.Error(), callback.RedirectURI()) {
		t.Errorf("error = %v, want the redirect URI named", err)
	}
}

func TestCallbackStopsOnCancel(t *testing.T) {
	callback := listen(t, "s1")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := callback.Wait(ctx, time.Minute)
	expectKind(t, err, failure.Interrupted)
}

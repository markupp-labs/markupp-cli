package session

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestLogoutRevokesAndForgetsTheLogin(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, time.Hour)
	removed, err := newTestAuthenticator(store, &bytes.Buffer{}, nil).Logout(context.Background())
	if err != nil || !removed {
		t.Fatalf("Logout = %v, %v, want removed", removed, err)
	}
	if revoked := server.revokedTokens(); len(revoked) != 1 || revoked[0] != "refresh-0" {
		t.Errorf("revoked = %q, want the refresh token", revoked)
	}
	if settings := loadSettings(t, store); settings.Auth != nil || settings.URL != server.URL {
		t.Errorf("settings = %+v, want the login removed and the URL kept", settings)
	}
}

func TestLogoutForgetsTheLoginWhenRevokeFails(t *testing.T) {
	server := newTokenServer(t)
	server.revokeStatus = http.StatusServiceUnavailable
	store := newStore(t)
	storeLogin(t, store, server.URL, time.Hour)
	removed, err := newTestAuthenticator(store, &bytes.Buffer{}, nil).Logout(context.Background())
	expectKind(t, err, failure.Unavailable)
	if !removed || !strings.Contains(failure.HintOf(err), "web panel") {
		t.Errorf("Logout = %v, %v (hint %q), want removed with a hint about the panel", removed, err, failure.HintOf(err))
	}
	if settings := loadSettings(t, store); settings.Auth != nil {
		t.Errorf("settings = %+v, want the login removed", settings)
	}
}

func TestLogoutWithoutLoginDoesNothing(t *testing.T) {
	server := newTokenServer(t)
	removed, err := newTestAuthenticator(newStore(t), &bytes.Buffer{}, nil).Logout(context.Background())
	if err != nil || removed {
		t.Errorf("Logout = %v, %v, want nothing removed", removed, err)
	}
	if requests, _ := server.counts(); requests != 0 {
		t.Errorf("server saw %d requests, want none", requests)
	}
}

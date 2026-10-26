package session

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestTokenUsesTheStoredTokenWhileFresh(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, time.Minute)
	token, err := newTestSource(store, server.URL).Token(context.Background())
	if err != nil || token != "access-0" {
		t.Errorf("Token = %q, %v, want the stored access-0", token, err)
	}
	if requests, _ := server.counts(); requests != 0 {
		t.Errorf("server saw %d requests, want none", requests)
	}
}

func TestTokenRefreshesWithinTheMargin(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, ExpiryMargin-time.Second)
	token, err := newTestSource(store, server.URL+"/").Token(context.Background())
	if err != nil || token != "access-1" {
		t.Fatalf("Token = %q, %v, want the refreshed access-1", token, err)
	}
	auth := loadSettings(t, store).Auth
	if auth == nil || auth.AccessToken != "access-1" || auth.RefreshToken != "refresh-1" || !auth.ExpiresAt.Equal(testNow.Add(time.Hour)) {
		t.Errorf("stored credential = %v, want access-1, refresh-1, expiring in an hour", auth)
	}
	if auth.Issuer != server.URL || auth.ClientID != ClientID || auth.Scope != Scope {
		t.Errorf("stored credential = %v, want issuer, client and scope kept", auth)
	}
}

func TestConcurrentRefreshRotatesOnce(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, -time.Minute)
	path, _ := store.Path()
	tokens := make([]string, 8)
	errs := make([]error, 8)
	var wg sync.WaitGroup
	for i := range tokens {
		wg.Go(func() {
			// Each source has its own store, like separate mu processes.
			tokens[i], errs[i] = newTestSource(config.NewStore(path), server.URL).Token(context.Background())
		})
	}
	wg.Wait()
	for i := range tokens {
		if errs[i] != nil || tokens[i] != "access-1" {
			t.Errorf("caller %d got %q, %v, want access-1", i, tokens[i], errs[i])
		}
	}
	if _, refreshes := server.counts(); refreshes != 1 {
		t.Errorf("server rotated the refresh token %d times, want once", refreshes)
	}
}

func TestInvalidGrantClearsTheLogin(t *testing.T) {
	server := newTokenServer(t)
	server.current = "refresh-spent-elsewhere"
	store := newStore(t)
	storeLogin(t, store, server.URL, -time.Minute)
	_, err := newTestSource(store, server.URL).Token(context.Background())
	expectKind(t, err, failure.Unauthenticated)
	if hint := failure.HintOf(err); !strings.Contains(hint, "mu login "+server.URL) {
		t.Errorf("hint = %q, want mu login %s", hint, server.URL)
	}
	settings := loadSettings(t, store)
	if settings.Auth != nil || settings.URL != server.URL || settings.Vault != "v1" {
		t.Errorf("settings = %+v, want the login removed and the rest kept", settings)
	}
}

func TestRejectedRefreshesOnlyTheTokenOnDisk(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, time.Hour)
	source := newTestSource(store, server.URL)
	fresh, retry, err := source.Rejected(context.Background(), "not-the-stored-token")
	if err != nil || retry || fresh != "" {
		t.Errorf("Rejected(foreign token) = %q, %v, %v, want no retry", fresh, retry, err)
	}
	fresh, retry, err = source.Rejected(context.Background(), "access-0")
	if err != nil || !retry || fresh != "access-1" {
		t.Errorf("Rejected(stored token) = %q, %v, %v, want a retry with access-1", fresh, retry, err)
	}
	if _, refreshes := server.counts(); refreshes != 1 {
		t.Errorf("refreshes = %d, want 1", refreshes)
	}
}

func TestTokenRefusesALoginForAnotherServer(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	storeLogin(t, store, server.URL, time.Hour)
	_, err := newTestSource(store, "https://other.example.com").Token(context.Background())
	expectKind(t, err, failure.Unauthenticated)
	if !strings.Contains(err.Error(), server.URL) || !strings.Contains(failure.HintOf(err), "mu login https://other.example.com") {
		t.Errorf("error = %v (hint %q), want the stored issuer and the login command", err, failure.HintOf(err))
	}
	if requests, _ := server.counts(); requests != 0 {
		t.Errorf("server saw %d requests, want none: the token must not travel", requests)
	}
}

func TestTokenWithoutLoginIsUnauthenticated(t *testing.T) {
	_, err := newTestSource(newStore(t), "https://markupp.example.com").Token(context.Background())
	expectKind(t, err, failure.Unauthenticated)
	if !strings.Contains(failure.HintOf(err), "mu login https://markupp.example.com") {
		t.Errorf("hint = %q, want the login command", failure.HintOf(err))
	}
}

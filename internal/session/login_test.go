package session

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func newTestAuthenticator(store Store, notice *bytes.Buffer, open func(string) error) Authenticator {
	return Authenticator{
		OAuth: NewOAuth(nil, false), Store: store, Now: fixedClock,
		Notice: notice, Open: open, Timeout: 5 * time.Second,
	}
}

func TestLoginStoresTheCredentialAndTheServer(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	var notice bytes.Buffer
	var opened string
	open := func(authorize string) error {
		opened = authorize
		browserVisit(t, server, authorize, url.Values{"code": {"code-1"}})
		return nil
	}
	cred, err := newTestAuthenticator(store, &notice, open).Login(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if cred.Issuer != server.URL || cred.AccessToken != "access-1" || !cred.ExpiresAt.Equal(testNow.Add(time.Hour)) {
		t.Errorf("credential = %v, want access-1 from %s expiring in an hour", cred, server.URL)
	}
	settings := loadSettings(t, store)
	if settings.URL != server.URL || settings.Auth == nil || settings.Auth.RefreshToken != "refresh-1" || settings.Auth.ClientID != ClientID {
		t.Errorf("settings = %+v, want the server and the credential stored", settings)
	}
	if !strings.Contains(notice.String(), opened) || opened == "" {
		t.Errorf("notice = %q, want the authorize URL %q", notice.String(), opened)
	}
}

func TestLoginGoesOnWhenTheBrowserDoesNotOpen(t *testing.T) {
	server := newTokenServer(t)
	open := func(authorize string) error {
		browserVisit(t, server, authorize, url.Values{"code": {"code-1"}})
		return errors.New("xdg-open: not found")
	}
	var notice bytes.Buffer
	if _, err := newTestAuthenticator(newStore(t), &notice, open).Login(context.Background(), server.URL); err != nil {
		t.Fatalf("Login with a failed browser = %v, want ok", err)
	}
}

func TestLoginRefusedSavesNothing(t *testing.T) {
	server := newTokenServer(t)
	store := newStore(t)
	open := func(authorize string) error {
		browserVisit(t, server, authorize, url.Values{"error": {"access_denied"}})
		return nil
	}
	var notice bytes.Buffer
	_, err := newTestAuthenticator(store, &notice, open).Login(context.Background(), server.URL)
	expectKind(t, err, failure.Unauthenticated)
	if settings := loadSettings(t, store); settings.Auth != nil || settings.URL != "" {
		t.Errorf("settings = %+v, want nothing saved", settings)
	}
}

func TestLoginTimesOutWithoutTheBrowser(t *testing.T) {
	server := newTokenServer(t)
	var notice bytes.Buffer
	auth := newTestAuthenticator(newStore(t), &notice, func(string) error { return nil })
	auth.Timeout = 20 * time.Millisecond
	_, err := auth.Login(context.Background(), server.URL)
	expectKind(t, err, failure.Unauthenticated)
}

func TestLoginWithoutOpenerOnlyPrintsTheURL(t *testing.T) {
	server := newTokenServer(t)
	var notice bytes.Buffer
	auth := newTestAuthenticator(newStore(t), &notice, nil)
	auth.Timeout = 20 * time.Millisecond
	_, _ = auth.Login(context.Background(), server.URL)
	if !strings.Contains(notice.String(), server.URL+"/oauth/authorize?") {
		t.Errorf("notice = %q, want the authorize URL", notice.String())
	}
}

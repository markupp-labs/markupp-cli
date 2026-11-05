package session

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

var testNow = time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

func fixedClock() time.Time { return testNow }

// tokenServer is a fake Markupp authorization server with a single use, rotating
// refresh token: reusing a spent one is an invalid_grant, as on the real server.
type tokenServer struct {
	*httptest.Server
	mu           sync.Mutex
	current      string
	generation   int
	refreshes    int
	requests     int
	challenge    string
	revoked      []string
	revokeStatus int
}

func newTokenServer(t *testing.T) *tokenServer {
	t.Helper()
	fake := &tokenServer{current: "refresh-0", revokeStatus: http.StatusOK}
	fake.Server = httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(fake.Close)
	return fake
}

func (f *tokenServer) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests++
	_ = r.ParseForm()
	switch r.URL.Path {
	case oauth.MetadataPath:
		f.metadata(w)
	case "/oauth/token":
		f.token(w, r.PostForm)
	case "/oauth/revoke":
		f.revoked = append(f.revoked, r.PostForm.Get("token"))
		w.WriteHeader(f.revokeStatus)
	default:
		http.NotFound(w, r)
	}
}

func (f *tokenServer) metadata(w http.ResponseWriter) {
	origin := f.URL
	reply(w, http.StatusOK, fmt.Sprintf(`{"issuer":%q,"authorization_endpoint":"%s/oauth/authorize","token_endpoint":"%s/oauth/token","revocation_endpoint":"%s/oauth/revoke"}`,
		origin, origin, origin, origin))
}

func (f *tokenServer) token(w http.ResponseWriter, form url.Values) {
	if !f.grantValid(form) {
		reply(w, http.StatusBadRequest, `{"error":"invalid_grant","error_description":"grant refused"}`)
		return
	}
	f.generation++
	f.current = fmt.Sprintf("refresh-%d", f.generation)
	reply(w, http.StatusOK, fmt.Sprintf(`{"access_token":"access-%d","token_type":"Bearer","expires_in":3600,"refresh_token":%q,"scope":%q}`,
		f.generation, f.current, Scope))
}

func (f *tokenServer) grantValid(form url.Values) bool {
	if form.Get("client_id") != ClientID || form.Get("resource") != f.URL {
		return false
	}
	if form.Get("grant_type") == "authorization_code" {
		return form.Get("code") == "code-1" && oauth.S256(form.Get("code_verifier")) == f.challenge
	}
	if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != f.current {
		return false
	}
	f.refreshes++
	return true
}

func (f *tokenServer) expectChallenge(challenge string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.challenge = challenge
}

func (f *tokenServer) counts() (requests, refreshes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests, f.refreshes
}

func (f *tokenServer) revokedTokens() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.revoked...)
}

func reply(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// newStore returns a real config store in a temporary directory.
func newStore(t *testing.T) *config.Store {
	t.Helper()
	return config.NewStore(filepath.Join(t.TempDir(), "markupp", "cli.json"))
}

// storeLogin saves a login for server whose access token expires after ttl.
func storeLogin(t *testing.T, store *config.Store, server string, ttl time.Duration) {
	t.Helper()
	settings := config.Settings{URL: server, Vault: "v1", Auth: &config.Credential{
		Issuer: server, ClientID: ClientID, AccessToken: "access-0", RefreshToken: "refresh-0",
		ExpiresAt: testNow.Add(ttl), Scope: Scope,
	}}
	if err := store.Save(settings); err != nil {
		t.Fatalf("saving the login: %v", err)
	}
}

func loadSettings(t *testing.T, store *config.Store) config.Settings {
	t.Helper()
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("loading the config: %v", err)
	}
	return settings
}

func newTestSource(store Store, server string) *Source {
	return NewSource(store, NewOAuth(nil, false), server, fixedClock)
}

func expectKind(t *testing.T, err error, want failure.Kind) {
	t.Helper()
	if got := failure.KindOf(err); err == nil || got != want {
		t.Fatalf("error = %v (kind %v), want kind %v", err, got, want)
	}
}

// browserVisit plays the browser: it reads the authorize URL, tells the server which
// challenge to expect and comes back to the callback with answer.
func browserVisit(t *testing.T, server *tokenServer, authorize string, answer url.Values) {
	t.Helper()
	parsed, err := url.Parse(authorize)
	if err != nil {
		t.Errorf("authorize URL %q: %v", authorize, err)
		return
	}
	query := parsed.Query()
	server.expectChallenge(query.Get("code_challenge"))
	answer.Set("state", query.Get("state"))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, query.Get("redirect_uri")+"?"+answer.Encode(), nil)
	if err != nil {
		t.Errorf("building the callback request: %v", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("visiting the callback: %v", err)
		return
	}
	_ = resp.Body.Close()
}

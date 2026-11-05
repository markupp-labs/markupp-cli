package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// authServer is a fake authorization server: it serves its metadata with its own
// origin as the issuer, and records the forms posted to the token and revocation
// endpoints.
type authServer struct {
	*httptest.Server
	mu          sync.Mutex
	paths       []string
	tokenForms  []url.Values
	revokeForms []url.Values
	metadata    func(origin string) string
	tokenReply  http.HandlerFunc
	revokeReply http.HandlerFunc
}

func newAuthServer(t *testing.T) *authServer {
	t.Helper()
	fake := &authServer{
		metadata:    standardMetadata,
		tokenReply:  replyJSON(http.StatusOK, `{"access_token":"access-1","token_type":"Bearer","expires_in":3600,"refresh_token":"refresh-1","scope":"notas:leitura"}`),
		revokeReply: replyJSON(http.StatusOK, `{}`),
	}
	fake.Server = httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(fake.Close)
	return fake
}

func standardMetadata(origin string) string {
	return fmt.Sprintf(`{"issuer":%q,"authorization_endpoint":"%s/oauth/authorize","token_endpoint":"%s/oauth/token","revocation_endpoint":"%s/oauth/revoke","scopes_supported":["notas:leitura"]}`,
		origin, origin, origin, origin)
}

func (f *authServer) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.paths = append(f.paths, r.Method+" "+r.URL.Path)
	f.mu.Unlock()
	switch r.URL.Path {
	case MetadataPath:
		if f.metadata == nil {
			http.NotFound(w, r)
			return
		}
		replyJSON(http.StatusOK, f.metadata(f.URL))(w, r)
	case "/oauth/token":
		f.recordForm(r, &f.tokenForms)
		f.tokenReply(w, r)
	case "/oauth/revoke":
		f.recordForm(r, &f.revokeForms)
		f.revokeReply(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (f *authServer) recordForm(r *http.Request, into *[]url.Values) {
	_ = r.ParseForm()
	form := r.PostForm
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		form = url.Values{"unexpected_content_type": {r.Header.Get("Content-Type")}}
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	*into = append(*into, form)
}

func (f *authServer) forms(of *[]url.Values) []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), *of...)
}

func (f *authServer) metadataOf() Metadata {
	return Metadata{
		Issuer:                f.URL,
		AuthorizationEndpoint: f.URL + "/oauth/authorize",
		TokenEndpoint:         f.URL + "/oauth/token",
		RevocationEndpoint:    f.URL + "/oauth/revoke",
	}
}

func replyJSON(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func testClient() *Client {
	return New(Config{ClientID: "markupp-cli", Scope: "notas:leitura notas:escrita"})
}

func expectKind(t *testing.T, err error, want failure.Kind) {
	t.Helper()
	if got := failure.KindOf(err); err == nil || got != want {
		t.Fatalf("error = %v (kind %v), want kind %v", err, got, want)
	}
}

// browserGet is what the browser does when the server redirects it to the callback.
func browserGet(t *testing.T, target string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatalf("building request to %s: %v", target, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", target, err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

func (f *authServer) seenPaths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.paths...)
}

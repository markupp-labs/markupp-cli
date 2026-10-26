package oauth

import (
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizeURLCarriesTheFlowParameters(t *testing.T) {
	meta := Metadata{Issuer: "https://markupp.example.com", AuthorizationEndpoint: "https://markupp.example.com/oauth/authorize"}
	pkce := PKCE{Verifier: "verifier", Challenge: "challenge"}
	raw := testClient().AuthorizeURL(meta, "http://127.0.0.1:5555/callback", pkce, "state-1")
	if !strings.HasPrefix(raw, meta.AuthorizationEndpoint+"?") {
		t.Fatalf("authorize URL = %q, want it on %s", raw, meta.AuthorizationEndpoint)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing %q: %v", raw, err)
	}
	want := url.Values{
		"response_type": {"code"}, "client_id": {"markupp-cli"},
		"redirect_uri": {"http://127.0.0.1:5555/callback"}, "code_challenge": {"challenge"},
		"code_challenge_method": {"S256"}, "state": {"state-1"},
		"scope": {"notas:leitura notas:escrita"}, "resource": {"https://markupp.example.com"},
	}
	if got := parsed.Query(); got.Encode() != want.Encode() {
		t.Errorf("query = %v, want %v", got, want)
	}
}

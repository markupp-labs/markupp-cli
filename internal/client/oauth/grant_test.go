package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestExchangeCodePostsTheCodeGrant(t *testing.T) {
	server := newAuthServer(t)
	got, err := testClient().ExchangeCode(context.Background(), server.metadataOf(), "code-1", "http://127.0.0.1:5555/callback", "verifier-1")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	want := Tokens{AccessToken: "access-1", RefreshToken: "refresh-1", ExpiresIn: time.Hour, Scope: "notas:leitura"}
	if got != want {
		t.Errorf("tokens = %+v, want %+v", got, want)
	}
	form := url.Values{
		"grant_type": {"authorization_code"}, "code": {"code-1"},
		"redirect_uri": {"http://127.0.0.1:5555/callback"}, "client_id": {"markupp-cli"},
		"code_verifier": {"verifier-1"}, "resource": {server.URL},
	}
	if posted := server.forms(&server.tokenForms); len(posted) != 1 || posted[0].Encode() != form.Encode() {
		t.Errorf("posted forms = %v, want %v", posted, form)
	}
}

func TestRefreshPostsTheRefreshGrant(t *testing.T) {
	server := newAuthServer(t)
	if _, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0"); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	form := url.Values{
		"grant_type": {"refresh_token"}, "refresh_token": {"refresh-0"},
		"client_id": {"markupp-cli"}, "resource": {server.URL},
	}
	if posted := server.forms(&server.tokenForms); len(posted) != 1 || posted[0].Encode() != form.Encode() {
		t.Errorf("posted forms = %v, want %v", posted, form)
	}
}

func TestInvalidGrantIsUnauthenticated(t *testing.T) {
	server := newAuthServer(t)
	server.tokenReply = replyJSON(http.StatusBadRequest, `{"error":"invalid_grant","error_description":"refresh token reused"}`)
	_, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0")
	expectKind(t, err, failure.Unauthenticated)
	if !IsInvalidGrant(err) || !strings.Contains(err.Error(), "refresh token reused") {
		t.Errorf("error = %v, want invalid_grant with the description", err)
	}
	if IsInvalidGrant(failure.New(failure.Usage, "", "other")) {
		t.Error("IsInvalidGrant accepted an unrelated error")
	}
}

func TestGrantErrorsNeverQuoteTheBody(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"success without refresh token": replyJSON(http.StatusOK, `{"access_token":"secret-access","expires_in":60}`),
		"error body not OAuth":          replyJSON(http.StatusBadRequest, `{"secret-access":"leaked"}`),
		"plain text error":              func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "secret-access", http.StatusBadRequest) },
	}
	for name, reply := range cases {
		t.Run(name, func(t *testing.T) {
			server := newAuthServer(t)
			server.tokenReply = reply
			_, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0")
			if err == nil || strings.Contains(err.Error(), "secret-access") {
				t.Errorf("error = %v, want an error that does not quote the body", err)
			}
		})
	}
}

func TestTokenResponseNotJSONKeepsTheDecodeError(t *testing.T) {
	server := newAuthServer(t)
	server.tokenReply = replyJSON(http.StatusOK, `secret-access`)
	_, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0")
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("error = %v, want it to wrap the *json.SyntaxError", err)
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "'s'") {
		t.Errorf("error = %q quotes the body", err)
	}
}

func TestMalformedTokensNameTheMissingFields(t *testing.T) {
	server := newAuthServer(t)
	server.tokenReply = replyJSON(http.StatusOK, `{"access_token":"a","expires_in":0}`)
	_, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0")
	expectKind(t, err, failure.Internal)
	if !strings.Contains(err.Error(), "expires_in") || !strings.Contains(err.Error(), "refresh_token") {
		t.Errorf("error = %v, want the missing fields named", err)
	}
}

func TestServerErrorOnTheTokenEndpointIsUnavailable(t *testing.T) {
	server := newAuthServer(t)
	server.tokenReply = replyJSON(http.StatusServiceUnavailable, `{"error":"temporarily_unavailable"}`)
	_, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0")
	expectKind(t, err, failure.Unavailable)
}

func TestCanceledGrantIsInterrupted(t *testing.T) {
	server := newAuthServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := testClient().Refresh(ctx, server.metadataOf(), "refresh-0")
	expectKind(t, err, failure.Interrupted)
}

func TestRefreshRefusesARedirectToAnotherHost(t *testing.T) {
	leaked := make(chan string, 1)
	other := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		leaked <- string(raw)
	}))
	t.Cleanup(other.Close)
	server := newAuthServer(t)
	server.tokenReply = func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL+"/oauth/token", http.StatusTemporaryRedirect)
	}
	if _, err := testClient().Refresh(context.Background(), server.metadataOf(), "refresh-0"); err == nil {
		t.Fatal("Refresh followed a 307 to another host, want an error")
	}
	select {
	case form := <-leaked:
		t.Errorf("the other host received %q, want no request", form)
	default:
	}
}

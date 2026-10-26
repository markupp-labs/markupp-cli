package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/session"
)

func realConnection(url, token string) Connection {
	return Connection{Resolved: config.Resolved{
		URL:   config.Value{Value: url, Source: config.FromFlag},
		Token: config.Value{Value: token, Source: config.FromEnv},
	}}
}

func TestConnectServerRefusesPlainHTTPOutsideLocalhost(t *testing.T) {
	_, err := connectServer(context.Background(), realConnection("http://markupp.example.com", "tok"))
	if failure.KindOf(err) != failure.Usage {
		t.Errorf("plain http error = %v, want usage", err)
	}
	conn := realConnection("http://markupp.example.com", "tok")
	conn.AllowPlainHTTP = true
	if _, err := connectServer(context.Background(), conn); err != nil {
		t.Errorf("plain http with the opt in = %v, want ok", err)
	}
}

func TestConnectServerNeedsACredentialUnlessPublic(t *testing.T) {
	_, err := connectServer(context.Background(), realConnection("https://markupp.example.com", ""))
	if failure.KindOf(err) != failure.Unauthenticated {
		t.Errorf("no token error = %v, want unauthenticated", err)
	}
	public := realConnection("https://markupp.example.com", "")
	public.Public = true
	if _, err := connectServer(context.Background(), public); err != nil {
		t.Errorf("public connect = %v, want ok", err)
	}
}

func TestConnectServerUsesTheStoredLogin(t *testing.T) {
	conn := realConnection("https://markupp.example.com", "")
	conn.Settings.Auth = &config.Credential{AccessToken: "stored"}
	if _, err := connectServer(context.Background(), conn); err != nil {
		t.Errorf("connect with stored login = %v, want ok", err)
	}
}

func TestTokenSourceRefreshesTheStoredLogin(t *testing.T) {
	conn := realConnection("https://markupp.example.com", "")
	conn.Settings.Auth = &config.Credential{AccessToken: "stored"}
	tokens, err := tokenSource(context.Background(), conn)
	if err != nil {
		t.Fatalf("tokenSource: %v", err)
	}
	if _, ok := tokens.(*session.Source); !ok {
		t.Errorf("token source = %T, want the refreshing *session.Source", tokens)
	}
}

func TestMarkuppTokenWinsOverTheStoredLogin(t *testing.T) {
	conn := realConnection("https://markupp.example.com", "raw")
	conn.Settings.Auth = &config.Credential{AccessToken: "stored"}
	tokens, err := tokenSource(context.Background(), conn)
	if err != nil || tokens != client.StaticToken("raw") {
		t.Errorf("token source = %#v, %v, want the static MARKUPP_TOKEN", tokens, err)
	}
}

func TestNotLoggedInHintNamesTheServer(t *testing.T) {
	_, err := tokenSource(context.Background(), realConnection("https://markupp.example.com", ""))
	if hint := failure.HintOf(err); !strings.Contains(hint, "mu login https://markupp.example.com") {
		t.Errorf("hint = %q, want mu login with the server", hint)
	}
}

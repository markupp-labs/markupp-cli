package cli

import (
	"context"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
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

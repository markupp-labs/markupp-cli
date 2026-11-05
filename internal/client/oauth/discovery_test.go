package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestDiscoverReadsTheMetadataAtTheOrigin(t *testing.T) {
	server := newAuthServer(t)
	got, err := testClient().Discover(context.Background(), server.URL+"/markupp/")
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != server.metadataOf() {
		t.Errorf("metadata = %+v, want %+v", got, server.metadataOf())
	}
	if paths := server.seenPaths(); len(paths) != 1 || paths[0] != "GET "+MetadataPath {
		t.Errorf("requests = %q, want one GET %s at the origin", paths, MetadataPath)
	}
}

func TestDiscoverRefusesAnIssuerOtherThanTheOrigin(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = func(origin string) string {
		return strings.Replace(standardMetadata(origin), `"issuer":"`+origin, `"issuer":"https://evil.example`, 1)
	}
	_, err := testClient().Discover(context.Background(), server.URL)
	expectKind(t, err, failure.Internal)
	if !strings.Contains(err.Error(), "https://evil.example") || !strings.Contains(err.Error(), server.URL) {
		t.Errorf("error = %v, want both the issuer and the expected origin", err)
	}
}

func TestDiscoverRefusesAnIssuerWithAPath(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = func(origin string) string {
		return strings.Replace(standardMetadata(origin), `"issuer":"`+origin, `"issuer":"`+origin+`/tenant`, 1)
	}
	_, err := testClient().Discover(context.Background(), server.URL)
	expectKind(t, err, failure.Internal)
}

func TestDiscoverRefusesEndpointsOnAnotherOrigin(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = func(origin string) string {
		return strings.Replace(standardMetadata(origin), origin+"/oauth/token", "https://evil.example/oauth/token", 1)
	}
	_, err := testClient().Discover(context.Background(), server.URL)
	expectKind(t, err, failure.Internal)
	if !strings.Contains(err.Error(), "token_endpoint") || !strings.Contains(err.Error(), "https://evil.example/oauth/token") {
		t.Errorf("error = %v, want the field and the offending URL", err)
	}
}

func TestDiscoverWithoutMetadataIsNotFound(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = nil
	_, err := testClient().Discover(context.Background(), server.URL)
	expectKind(t, err, failure.NotFound)
	if failure.HintOf(err) == "" {
		t.Errorf("error = %v, want a hint about the server version", err)
	}
}

func TestDiscoverRefusesPlainHTTPOutsideLocalhost(t *testing.T) {
	_, err := testClient().Discover(context.Background(), "http://markupp.example.com")
	expectKind(t, err, failure.Usage)
}

func TestDiscoverWithMalformedMetadataNamesTheShape(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = func(string) string { return `{"issuer":42}` }
	_, err := testClient().Discover(context.Background(), server.URL)
	expectKind(t, err, failure.Internal)
	if !strings.Contains(err.Error(), "token_endpoint") {
		t.Errorf("error = %v, want the expected shape", err)
	}
}

func TestDiscoverWithMalformedMetadataKeepsTheDecodeError(t *testing.T) {
	server := newAuthServer(t)
	server.metadata = func(string) string { return `{"issuer":42}` }
	_, err := testClient().Discover(context.Background(), server.URL)
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Errorf("error = %v, want it to wrap the *json.UnmarshalTypeError", err)
	}
}

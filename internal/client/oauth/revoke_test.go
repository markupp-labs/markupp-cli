package oauth

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestRevokePostsTheRefreshToken(t *testing.T) {
	server := newAuthServer(t)
	if err := testClient().Revoke(context.Background(), server.metadataOf(), "refresh-1"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	form := url.Values{"token": {"refresh-1"}, "token_type_hint": {"refresh_token"}, "client_id": {"markupp-cli"}}
	if posted := server.forms(&server.revokeForms); len(posted) != 1 || posted[0].Encode() != form.Encode() {
		t.Errorf("posted forms = %v, want %v", posted, form)
	}
}

func TestRevokeFailureIsReported(t *testing.T) {
	server := newAuthServer(t)
	server.revokeReply = replyJSON(http.StatusServiceUnavailable, `{"error":"temporarily_unavailable"}`)
	err := testClient().Revoke(context.Background(), server.metadataOf(), "refresh-1")
	expectKind(t, err, failure.Unavailable)
}

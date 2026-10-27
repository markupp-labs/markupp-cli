package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

const (
	codeInvalidGrant = "invalid_grant"
	tokensShape      = "expected access_token, a positive integer expires_in and refresh_token"
)

// tokenResponse is the success body of the token endpoint. It holds tokens, so no error
// ever quotes it.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// errorResponse is the error body of the token and revocation endpoints (RFC 6749,
// section 5.2).
type errorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// ExchangeCode trades the authorization code from the callback for tokens, proving the
// PKCE verifier.
func (c *Client) ExchangeCode(ctx context.Context, meta Metadata, code, redirectURI, verifier string) (Tokens, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {c.cfg.ClientID},
		"code_verifier": {verifier},
		"resource":      {meta.Issuer},
	}
	return c.requestTokens(ctx, meta.TokenEndpoint, form)
}

// Refresh trades refreshToken for a new pair. The server rotates the refresh token:
// the one passed here is spent, and reusing it revokes the whole authorization.
func (c *Client) Refresh(ctx context.Context, meta Metadata, refreshToken string) (Tokens, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {c.cfg.ClientID},
		"resource":      {meta.Issuer},
	}
	return c.requestTokens(ctx, meta.TokenEndpoint, form)
}

// IsInvalidGrant reports the refusal of a code or refresh token, which only a new
// login can fix.
func IsInvalidGrant(err error) bool {
	var classified *failure.Error
	return errors.As(err, &classified) && classified.Code == codeInvalidGrant
}

func (c *Client) requestTokens(ctx context.Context, endpoint string, form url.Values) (Tokens, error) {
	resp, err := c.post(ctx, endpoint, form)
	if err != nil {
		return Tokens{}, err
	}
	defer discard(resp)
	return decodeTokens("POST "+endpoint, resp)
}

// post sends form to endpoint and returns the response when it is a 200.
func (c *Client) post(ctx context.Context, endpoint string, form url.Values) (*http.Response, error) {
	resp, err := c.send(ctx, http.MethodPost, endpoint, formBody(form))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer discard(resp)
		return nil, grantError("POST "+endpoint, resp)
	}
	return resp, nil
}

// grantError reads the OAuth error code and description. A body in another shape is
// never quoted: only its status is reported.
func grantError(subject string, resp *http.Response) error {
	var body errorResponse
	if err := json.NewDecoder(limitBody(resp)).Decode(&body); err != nil || body.Error == "" {
		message := fmt.Sprintf("server answered %d without an OAuth error body, expected 200", resp.StatusCode)
		return failure.New(kindForStatus(resp.StatusCode), subject, message)
	}
	kind := kindForStatus(resp.StatusCode)
	if body.Error == codeInvalidGrant {
		kind = failure.Unauthenticated
	}
	message := fmt.Sprintf("server refused the request with %d: %s", resp.StatusCode, body.Description)
	return &failure.Error{Kind: kind, Subject: subject, Code: body.Error, Message: message}
}

func decodeTokens(subject string, resp *http.Response) (Tokens, error) {
	var body tokenResponse
	if err := json.NewDecoder(limitBody(resp)).Decode(&body); err != nil {
		return Tokens{}, failure.New(failure.Internal, subject, "token response is not a JSON object, "+tokensShape).WithCause(err)
	}
	if missing := missingTokenFields(body); len(missing) > 0 {
		message := fmt.Sprintf("token response lacks [%s], %s", strings.Join(missing, ", "), tokensShape)
		return Tokens{}, failure.New(failure.Internal, subject, message)
	}
	return Tokens{
		AccessToken:  body.AccessToken,
		RefreshToken: body.RefreshToken,
		ExpiresIn:    time.Duration(body.ExpiresIn) * time.Second,
		Scope:        body.Scope,
	}, nil
}

func missingTokenFields(body tokenResponse) []string {
	var missing []string
	if body.AccessToken == "" {
		missing = append(missing, "access_token")
	}
	if body.ExpiresIn <= 0 {
		missing = append(missing, "expires_in")
	}
	if body.RefreshToken == "" {
		missing = append(missing, "refresh_token")
	}
	return missing
}

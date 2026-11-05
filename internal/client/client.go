// Package client is the only package that knows the Markupp server speaks HTTP. It
// builds every request under /v1, attaches the credentials, maps error answers to
// failure kinds and hands successful bodies back unread, so --json can pass them
// through unchanged.
package client

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// TokenSource supplies the bearer token of each request. Rejected is called once when
// the server answers 401 to token; it returns a fresh token to retry with, or
// retry=false to give up. The session package implements it with refresh under the
// config lock.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
	Rejected(ctx context.Context, token string) (fresh string, retry bool, err error)
}

// StaticToken is a TokenSource that never refreshes, for MARKUPP_TOKEN.
type StaticToken string

// Token returns the token itself.
func (t StaticToken) Token(context.Context) (string, error) { return string(t), nil }

// Rejected never offers a retry: a raw token cannot be refreshed.
func (StaticToken) Rejected(context.Context, string) (string, bool, error) {
	return "", false, nil
}

// Options configures a Client. Transport nil means the default transport with
// timeouts; Logger nil means no request log.
type Options struct {
	BaseURL        string
	Tenant         string
	Tokens         TokenSource
	AllowPlainHTTP bool
	Transport      http.RoundTripper
	Logger         *slog.Logger
	UserAgent      string
}

// Client talks to one Markupp server.
type Client struct {
	base      *url.URL
	tenant    string
	tokens    TokenSource
	http      *http.Client
	userAgent string
}

// New returns a Client that authenticates every request with opts.Tokens and sends
// opts.Tenant in X-Markupp-Tenant when it is set.
func New(opts Options) (*Client, error) {
	if opts.Tokens == nil {
		return nil, failure.New(failure.Unauthenticated, opts.BaseURL, "not logged in").
			WithHint("run mu login, or set MARKUPP_TOKEN")
	}
	return build(opts)
}

// NewPublic returns a Client for the public routes: it sends neither a token nor a
// tenant, whatever opts carries.
func NewPublic(opts Options) (*Client, error) {
	opts.Tokens = nil
	opts.Tenant = ""
	return build(opts)
}

func build(opts Options) (*Client, error) {
	base, err := CheckBaseURL(opts.BaseURL, opts.AllowPlainHTTP)
	if err != nil {
		return nil, err
	}
	return &Client{
		base:      base,
		tenant:    opts.Tenant,
		tokens:    opts.Tokens,
		http:      NewHTTPClient(opts.Transport, opts.Logger),
		userAgent: opts.UserAgent,
	}, nil
}

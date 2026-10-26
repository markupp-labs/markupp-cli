// Package client is the only package that knows the Markupp server speaks HTTP.
package client

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
)

type TokenSource interface {
	Token(ctx context.Context) (string, error)
	Rejected(ctx context.Context, token string) (fresh string, retry bool, err error)
}

type StaticToken string

func (t StaticToken) Token(context.Context) (string, error) { return string(t), nil }
func (StaticToken) Rejected(context.Context, string) (string, bool, error) {
	return "", false, nil
}

type Options struct {
	BaseURL        string
	Tenant         string
	Tokens         TokenSource
	AllowPlainHTTP bool
	Transport      http.RoundTripper
	Logger         *slog.Logger
	UserAgent      string
}

type Client struct{}

type request struct {
	method string
	path   string
	query  url.Values
	body   []byte
}

type Page struct {
	Limit  int
	Offset int
}

const (
	MaxLimit  = 100
	MaxOffset = 10000
)

func New(opts Options) (*Client, error)                              { return &Client{}, nil }
func NewPublic(opts Options) (*Client, error)                        { return &Client{}, nil }
func CheckBaseURL(raw string, allowPlainHTTP bool) (*url.URL, error) { return nil, nil }
func NewHTTPClient(transport http.RoundTripper, logger *slog.Logger) *http.Client {
	return &http.Client{}
}
func (c *Client) BaseURL() *url.URL                                          { return nil }
func endpoint(segments ...string) string                                     { return "" }
func crossesTrust(from, to *url.URL) bool                                    { return false }
func (c *Client) open(ctx context.Context, r request) (io.ReadCloser, error) { return nil, nil }
func (c *Client) exec(ctx context.Context, r request) error                  { return nil }
func Elements(body io.ReadCloser) iter.Seq2[json.RawMessage, error] {
	return func(func(json.RawMessage, error) bool) {}
}
func (p Page) apply(q url.Values)                 {}
func setNonEmpty(q url.Values, key, value string) {}

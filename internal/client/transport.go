package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

const (
	apiPrefix      = "v1"
	tenantHeader   = "X-Markupp-Tenant"
	maxRedirects   = 10
	headerTimeout  = 30 * time.Second
	connectTimeout = 10 * time.Second
)

// request is one API call. path is already escaped, built with endpoint, and is
// relative to /v1. body is a JSON document, nil for none.
type request struct {
	method string
	path   string
	query  url.Values
	body   []byte
}

func (r request) subject() string {
	return r.method + " /" + apiPrefix + r.path
}

// endpoint joins path segments, escaping each one so an id can never add a segment or
// climb out of the route.
func endpoint(segments ...string) string {
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = strings.ReplaceAll(url.PathEscape(segment), ".", "%2E")
	}
	return "/" + strings.Join(escaped, "/")
}

// CheckBaseURL parses the server URL and refuses plain http outside the loopback
// interface unless allowPlainHTTP, since the token would travel unencrypted.
func CheckBaseURL(raw string, allowPlainHTTP bool) (*url.URL, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Scheme != "http") {
		return nil, failure.New(failure.Usage, "", fmt.Sprintf("invalid server URL %q, expected http(s)://host[:port][/path]", raw))
	}
	if base.Scheme == "http" && !allowPlainHTTP && !isLoopback(base.Hostname()) {
		message := fmt.Sprintf("refusing plain http to %q: the token would travel unencrypted; use https, or pass --allow-plain-http to accept the risk", raw)
		return nil, failure.New(failure.Usage, "", message)
	}
	return base, nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// NewHTTPClient returns the http.Client every request of mu goes through: bounded
// redirects that drop the credentials when they leave the host or downgrade to http,
// and refuse to resend a request body there (307 and 308 replay it, and the body of an
// OAuth grant or a note is as secret as the token), and the optional request log. The oauth package uses it too.
func NewHTTPClient(transport http.RoundTripper, logger *slog.Logger) *http.Client {
	if transport == nil {
		transport = defaultTransport()
	}
	if logger != nil {
		transport = &loggingTransport{next: transport, logger: logger}
	}
	return &http.Client{Transport: transport, CheckRedirect: checkRedirect}
}

func defaultTransport() http.RoundTripper {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = headerTimeout
	transport.TLSHandshakeTimeout = connectTimeout
	return transport
}

func checkRedirect(next *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects at %s", maxRedirects, next.URL.Redacted())
	}
	if !crossesTrust(via[0].URL, next.URL) {
		return nil
	}
	if next.GetBody != nil {
		return fmt.Errorf("refusing to resend the body of %s %s to %s: the redirect leaves the host or drops https",
			via[0].Method, via[0].URL.Redacted(), next.URL.Redacted())
	}
	next.Header.Del("Authorization")
	next.Header.Del(tenantHeader)
	return nil
}

// crossesTrust reports a redirect to another host or port, or from https to http.
func crossesTrust(from, to *url.URL) bool {
	return from.Host != to.Host || (from.Scheme == "https" && to.Scheme != "https")
}

// send runs r and returns the successful response, with one retry when the server
// rejects the token and the TokenSource offers a fresh one. Any status from 400 up is
// mapped to a *failure.Error and the body is closed.
func (c *Client) send(ctx context.Context, r request) (*http.Response, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.attempt(ctx, r, token)
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		resp, err = c.retryRejected(ctx, r, token, resp)
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer discard(resp)
		return nil, responseError(resp, r.subject())
	}
	return resp, nil
}

func (c *Client) token(ctx context.Context) (string, error) {
	if c.tokens == nil {
		return "", nil
	}
	return c.tokens.Token(ctx)
}

func (c *Client) retryRejected(ctx context.Context, r request, token string, rejected *http.Response) (*http.Response, error) {
	if c.tokens == nil {
		return rejected, nil
	}
	fresh, retry, err := c.tokens.Rejected(ctx, token)
	if err != nil || !retry {
		if err != nil {
			discard(rejected)
		}
		return rejected, err
	}
	discard(rejected)
	return c.attempt(ctx, r, fresh)
}

func (c *Client) attempt(ctx context.Context, r request, token string) (*http.Response, error) {
	req, err := c.newHTTPRequest(ctx, r, token)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, transportError(ctx, r.subject(), err)
	}
	return resp, nil
}

func (c *Client) newHTTPRequest(ctx context.Context, r request, token string) (*http.Request, error) {
	target := c.base.JoinPath(apiPrefix + r.path)
	target.RawQuery = r.query.Encode()
	req, err := http.NewRequestWithContext(ctx, r.method, target.String(), bodyReader(r.body))
	if err != nil {
		return nil, failure.Wrap(failure.Internal, r.subject(), err)
	}
	c.setHeaders(req.Header, token, r.body != nil)
	return req, nil
}

func (c *Client) setHeaders(header http.Header, token string, hasBody bool) {
	header.Set("Accept", "application/json")
	if hasBody {
		header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		header.Set("User-Agent", c.userAgent)
	}
	if token != "" {
		header.Set("Authorization", "Bearer "+token)
	}
	if c.tenant != "" {
		header.Set(tenantHeader, c.tenant)
	}
}

// bodyReader returns a nil reader for no body, so the request has no Content-Length,
// and a bytes.Reader otherwise, which lets http replay the body on a redirect.
func bodyReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return bytes.NewReader(body)
}

// discard drains a little of the body, so the connection can be reused, and closes it.
func discard(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorBody))
	_ = resp.Body.Close()
}

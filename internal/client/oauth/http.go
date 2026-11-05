package oauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// send runs one request with the request timeout. form nil means no body; otherwise it
// is posted form encoded. The subject of every error is the method and URL, which
// never carry a token.
func (c *Client) send(ctx context.Context, method, target string, form io.Reader) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	req, err := http.NewRequestWithContext(ctx, method, target, form)
	if err != nil {
		cancel()
		return nil, failure.Wrap(failure.Internal, method+" "+target, err)
	}
	req.Header.Set("Accept", "application/json")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		cancel()
		return nil, transportError(ctx, method+" "+target, err)
	}
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

// cancelOnClose releases the timeout of a request when its body is closed.
type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelOnClose) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

func formBody(form url.Values) io.Reader {
	return strings.NewReader(form.Encode())
}

// transportError classifies a request that got no answer.
func transportError(ctx context.Context, subject string, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return failure.Wrap(failure.Interrupted, subject, context.Canceled)
	}
	return failure.Wrap(failure.Unavailable, subject, err).
		WithHint("check that the server is up and reachable")
}

func kindForStatus(status int) failure.Kind {
	switch {
	case status == http.StatusUnauthorized:
		return failure.Unauthenticated
	case status == http.StatusForbidden:
		return failure.Forbidden
	case status == http.StatusNotFound:
		return failure.NotFound
	case status == http.StatusTooManyRequests:
		return failure.RateLimited
	case status >= http.StatusInternalServerError:
		return failure.Unavailable
	}
	return failure.Usage
}

func limitBody(resp *http.Response) io.Reader {
	return io.LimitReader(resp.Body, maxBody)
}

// discard drains a little of the body, so the connection can be reused, and closes it.
func discard(resp *http.Response) {
	_, _ = io.Copy(io.Discard, limitBody(resp))
	_ = resp.Body.Close()
}

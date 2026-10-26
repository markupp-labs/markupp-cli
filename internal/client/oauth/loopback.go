package oauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

const callbackReadTimeout = 10 * time.Second

// Callback is the server on the loopback interface the browser comes back to after
// the consent (RFC 8252, section 7.3). It accepts the first answer that carries the
// expected state and ignores the others, so a stale tab or a forged request cannot end
// the login.
type Callback struct {
	server      *http.Server
	redirectURI string
	state       string
	results     chan callbackResult
}

type callbackResult struct {
	code string
	err  error
}

// callbackReply is what the browser sees, and the result when the answer is accepted.
type callbackReply struct {
	status  int
	message string
	result  *callbackResult
}

// Listen starts the callback server on 127.0.0.1 on a free port.
func Listen(ctx context.Context, state string) (*Callback, error) {
	var config net.ListenConfig
	listener, err := config.Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, failure.Wrap(failure.Internal, "listening on 127.0.0.1 for the OAuth callback", err)
	}
	callback := &Callback{
		redirectURI: "http://" + listener.Addr().String() + CallbackPath,
		state:       state,
		results:     make(chan callbackResult, 1),
	}
	callback.server = &http.Server{Handler: http.HandlerFunc(callback.serveHTTP), ReadHeaderTimeout: callbackReadTimeout}
	go func() { _ = callback.server.Serve(listener) }()
	return callback, nil
}

// RedirectURI is the redirect_uri of this login: http://127.0.0.1:<port>/callback.
func (c *Callback) RedirectURI() string { return c.redirectURI }

// Wait returns the authorization code of the first accepted answer, or fails when the
// server refused the login, when timeout passes or when ctx ends.
func (c *Callback) Wait(ctx context.Context, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-c.results:
		return result.code, result.err
	case <-timer.C:
		message := fmt.Sprintf("the browser did not come back within %s, expected a redirect after the consent", timeout)
		return "", failure.New(failure.Unauthenticated, c.redirectURI, message).WithHint("run mu login again")
	case <-ctx.Done():
		return "", failure.Wrap(failure.Interrupted, c.redirectURI, ctx.Err())
	}
}

// Close stops the callback server.
func (c *Callback) Close() error {
	if err := c.server.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("closing the OAuth callback on %s: %w", c.redirectURI, err)
	}
	return nil
}

func (c *Callback) serveHTTP(w http.ResponseWriter, r *http.Request) {
	reply := c.judge(r.URL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(reply.status)
	_, _ = w.Write([]byte(reply.message + "\n"))
	if reply.result == nil {
		return
	}
	select {
	case c.results <- *reply.result:
	default:
	}
}

func (c *Callback) judge(u *url.URL) callbackReply {
	if u.Path != CallbackPath {
		return callbackReply{status: http.StatusNotFound, message: "Unknown path."}
	}
	query := u.Query()
	if subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(c.state)) != 1 {
		return callbackReply{status: http.StatusBadRequest, message: "This answer does not belong to the login in progress."}
	}
	result := c.resultOf(query)
	if result.err != nil {
		return callbackReply{status: http.StatusOK, message: "Login refused. Go back to the terminal to see why.", result: &result}
	}
	return callbackReply{status: http.StatusOK, message: "Login complete. You can close this tab and go back to the terminal.", result: &result}
}

func (c *Callback) resultOf(query url.Values) callbackResult {
	if code := query.Get("error"); code != "" {
		message := "the server refused the login: " + query.Get("error_description")
		return callbackResult{err: &failure.Error{Kind: failure.Unauthenticated, Subject: c.redirectURI, Code: code, Message: message}}
	}
	code := query.Get("code")
	if code == "" {
		message := "the browser came back without code or error, expected ?code=...&state=..."
		return callbackResult{err: failure.New(failure.Internal, c.redirectURI, message)}
	}
	return callbackResult{code: code}
}

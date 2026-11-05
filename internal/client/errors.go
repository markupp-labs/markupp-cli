package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// maxErrorBody bounds how much of an error answer is read.
const maxErrorBody = 64 << 10

// Server error codes that change the hint.
const (
	codeEndpointAbsent    = "endpoint_absent"
	codeFirstPartyOnly    = "first_party_only"
	codeInsufficientScope = "insufficient_scope"
)

// responseError maps an error answer to a *failure.Error: the kind from the status,
// the code and message from the ErrorResponse body, and the hints the headers allow.
func responseError(resp *http.Response, subject string) *failure.Error {
	body, isJSON := readErrorBody(resp)
	err := &failure.Error{
		Kind:    kindForStatus(resp.StatusCode),
		Subject: subject,
		Code:    body.Error,
		Message: body.Message,
	}
	if err.Message == "" {
		err.Message = fmt.Sprintf("server answered %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if resp.StatusCode == http.StatusNotFound && !isJSON {
		markEndpointAbsent(err)
	}
	addHint(err, resp.Header)
	return err
}

func readErrorBody(resp *http.Response) (schema.ErrorResponse, bool) {
	var body schema.ErrorResponse
	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if !strings.HasSuffix(mediaType, "json") {
		return body, false
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
	if err != nil || json.Unmarshal(raw, &body) != nil {
		return schema.ErrorResponse{}, false
	}
	return body, body.Error != ""
}

func kindForStatus(status int) failure.Kind {
	switch {
	case status == http.StatusBadRequest, status == http.StatusUnprocessableEntity:
		return failure.Usage
	case status == http.StatusUnauthorized:
		return failure.Unauthenticated
	case status == http.StatusForbidden:
		return failure.Forbidden
	case status == http.StatusNotFound:
		return failure.NotFound
	case status == http.StatusConflict:
		return failure.Conflict
	case status == http.StatusTooManyRequests:
		return failure.RateLimited
	case status >= http.StatusInternalServerError:
		return failure.Unavailable
	}
	return failure.Internal
}

// markEndpointAbsent handles the plain text 404 of the router: the route does not
// exist, which means a server older than the command.
func markEndpointAbsent(err *failure.Error) {
	err.Code = codeEndpointAbsent
	err.Message = "the server has no such endpoint"
	err.Hint = "the server may be older than this command requires; check --url or upgrade the server"
}

func addHint(err *failure.Error, header http.Header) {
	switch {
	case err.Kind == failure.Unauthenticated:
		err.Hint = "run mu login, or check MARKUPP_TOKEN"
	case err.Code == codeFirstPartyOnly:
		err.Hint = "this operation is only available in the web panel"
	case err.Kind == failure.Forbidden:
		addScopeHint(err, header.Get("WWW-Authenticate"))
	case err.Kind == failure.RateLimited:
		addRetryHint(err, header.Get("Retry-After"))
	}
}

func addScopeHint(err *failure.Error, challenge string) {
	err.Scope = challengeParam(challenge, "scope")
	if err.Scope == "" && err.Code != codeInsufficientScope {
		return
	}
	err.Hint = fmt.Sprintf("the token lacks the scope %q; run mu login again to grant it", err.Scope)
}

// challengeParam reads name="value" from a WWW-Authenticate challenge.
func challengeParam(challenge, name string) string {
	marker := name + `="`
	start := strings.Index(challenge, marker)
	if start < 0 {
		return ""
	}
	rest := challenge[start+len(marker):]
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func addRetryHint(err *failure.Error, retryAfter string) {
	seconds, parseErr := strconv.Atoi(strings.TrimSpace(retryAfter))
	if parseErr != nil || seconds < 0 {
		err.Hint = "too many requests; wait and retry"
		return
	}
	err.RetryAfter = time.Duration(seconds) * time.Second
	err.Hint = fmt.Sprintf("retry after %d seconds", seconds)
}

// transportError classifies a request that got no answer: an interrupt when the
// context was canceled, the server unavailable otherwise.
func transportError(ctx context.Context, subject string, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return failure.Wrap(failure.Interrupted, subject, ctx.Err())
	}
	return failure.Wrap(failure.Unavailable, subject, err).
		WithHint("check that the server at --url is up and reachable")
}

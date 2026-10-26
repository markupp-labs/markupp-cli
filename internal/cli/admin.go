package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// adminScopeNote ends the help of the commands that need the administracao scope.
const adminScopeNote = `With an OAuth login the token needs the administracao scope; a 403 names
the missing scope, and an operation reserved to the web panel says so.`

// bodyFetch reaches the server through svc and returns the raw body of the answer.
type bodyFetch func(ctx context.Context, svc Services) (io.ReadCloser, error)

// serverCall reaches the server through svc for an answer without content.
type serverCall func(ctx context.Context, svc Services) error

// addNDJSONFlag adds --ndjson to a listing the server does not paginate.
func addNDJSONFlag(cmd *cobra.Command, ndjson *bool) {
	cmd.Flags().BoolVar(ndjson, "ndjson", false, "print one JSON object per line, as the results stream in")
}

// listFrom renders the array fetch returns as a table, JSON or NDJSON.
func (a *App) listFrom(cmd *cobra.Command, ndjson bool, fetch bodyFetch, columns []output.Column, row rowFunc) error {
	r, err := a.renderer(ndjson)
	if err != nil {
		return err
	}
	body, err := a.fetchBody(cmd.Context(), fetch)
	if err != nil {
		return err
	}
	return r.list(body, columns, row)
}

// objectFrom renders the object fetch returns through render, or as JSON.
func (a *App) objectFrom(cmd *cobra.Command, fetch bodyFetch, render objectFunc) error {
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	body, err := a.fetchBody(cmd.Context(), fetch)
	if err != nil {
		return err
	}
	return r.object(body, render)
}

func (a *App) fetchBody(ctx context.Context, fetch bodyFetch) (io.ReadCloser, error) {
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return nil, err
	}
	return fetch(ctx, svc)
}

// callServer runs call and, when it succeeds, prints confirmation on stderr.
func (a *App) callServer(cmd *cobra.Command, call serverCall, confirmation string) error {
	svc, err := a.connect(cmd.Context(), Target{})
	if err != nil {
		return err
	}
	if err := call(cmd.Context(), svc); err != nil {
		return err
	}
	a.confirm(confirmation)
	return nil
}

// confirm prints a sanitized confirmation line on stderr.
func (a *App) confirm(message string) {
	_, _ = fmt.Fprintln(a.deps.Stderr, output.SanitizeLine(message))
}

// checkRole refuses a role outside the Role schema before reaching the server;
// subject names where the role came from.
func checkRole(subject, role string) error {
	if schema.IsRole(role) {
		return nil
	}
	message := fmt.Sprintf("invalid role %q, expected one of %s", role, strings.Join(schema.Roles, ", "))
	return failure.New(failure.Usage, subject, message)
}

// shortTime prints a server timestamp in UTC to the second, so it fits its column. A
// value that is not RFC 3339 is printed as it came.
func shortTime(stamp string) string {
	parsed, err := time.Parse(time.RFC3339Nano, stamp)
	if err != nil {
		return stamp
	}
	return parsed.UTC().Format(time.RFC3339)
}

// timeColumn is the width of a timestamp printed by shortTime.
const timeColumn = len("2006-01-02T15:04:05Z")

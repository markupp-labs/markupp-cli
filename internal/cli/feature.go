package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// The server features that some commands need beyond the published contract. The
// help of those commands names the feature, and so does the error when the server
// answers that it has no such endpoint.
const (
	featureVaults        = "vault management"
	featurePublication   = "vault publication"
	featureNotifications = "notifications"
)

// requiresServer appends to the long help of a command the server feature it needs.
func requiresServer(feature, long string) string {
	return long + "\n\nRequires a server with " + feature + "."
}

// withFeature turns the failure of an endpoint the server lacks into one that names
// the feature, and returns any other error unchanged.
func withFeature(feature string, err error) error {
	var absent *failure.Error
	if !client.EndpointAbsent(err) || !errors.As(err, &absent) {
		return err
	}
	return &failure.Error{
		Kind: failure.NotFound, Code: client.CodeEndpointAbsent, Subject: absent.Subject,
		Message: "endpoint absent", Err: err,
		Hint: fmt.Sprintf("this command requires a server with %s; check --url or upgrade the server", feature),
	}
}

// featureRead returns a pass through for the answer of a read, that names feature
// when the server lacks the endpoint:
//
//	body, err := featureRead(featureVaults)(svc.Vaults.ListVaults(ctx))
func featureRead(feature string) func(io.ReadCloser, error) (io.ReadCloser, error) {
	return func(body io.ReadCloser, err error) (io.ReadCloser, error) {
		if err != nil {
			return nil, withFeature(feature, err)
		}
		return body, nil
	}
}

// openFeature connects with the account credentials and runs call, naming feature
// when the server lacks the endpoint.
func (a *App) openFeature(ctx context.Context, feature string, call func(Services) (io.ReadCloser, error)) (io.ReadCloser, error) {
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return nil, err
	}
	return featureRead(feature)(call(svc))
}

// execFeature is openFeature for a call that answers without content.
func (a *App) execFeature(ctx context.Context, feature string, call func(Services) error) error {
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return err
	}
	return withFeature(feature, call(svc))
}

// inform writes a message for the person on stderr: what a command without output
// did, or a summary that is not part of the data.
func (a *App) inform(message string) {
	_, _ = fmt.Fprintln(a.deps.Stderr, message)
}

// showFeature runs call through openFeature and prints the object it answers.
func (a *App) showFeature(ctx context.Context, feature string, render objectFunc, call func(Services) (io.ReadCloser, error)) error {
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	body, err := a.openFeature(ctx, feature, call)
	if err != nil {
		return err
	}
	return r.object(body, render)
}

package oauth

import (
	"context"
	"time"
)

// Callback receives the browser back on the loopback interface.
type Callback struct{}

// Listen starts the callback server.
func Listen(context.Context, string) (*Callback, error) { return &Callback{}, nil }

// RedirectURI is the redirect_uri of the callback.
func (c *Callback) RedirectURI() string { return "" }

// Wait returns the authorization code.
func (c *Callback) Wait(context.Context, time.Duration) (string, error) {
	return "", errNotImplemented
}

// Close stops the callback server.
func (c *Callback) Close() error { return nil }

package session

import (
	"context"
	"io"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
	"github.com/markupp-labs/markupp-cli/internal/config"
)

// Authenticator logs in and out.
type Authenticator struct {
	OAuth   *oauth.Client
	Store   Store
	Now     Clock
	Notice  io.Writer
	Open    func(url string) error
	Timeout time.Duration
}

// Login runs the browser flow.
func (a Authenticator) Login(context.Context, string) (config.Credential, error) {
	return config.Credential{}, errNotImplemented
}

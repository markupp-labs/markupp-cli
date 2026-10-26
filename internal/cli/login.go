package cli

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/config"
)

// Authenticator logs in and out of a server. session.Authenticator satisfies it.
type Authenticator interface {
	Login(ctx context.Context, server string) (config.Credential, error)
	Logout(ctx context.Context) (removed bool, err error)
}

// AuthSetup is what an AuthFactory needs to build an Authenticator.
type AuthSetup struct {
	Store          ConfigStore
	AllowPlainHTTP bool
	Logger         *slog.Logger
	Clock          func() time.Time
	Notice         io.Writer
	OpenURL        func(url string) error
}

// AuthFactory builds the Authenticator of one execution.
type AuthFactory func(setup AuthSetup) Authenticator

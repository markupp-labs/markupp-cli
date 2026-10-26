package cli

import (
	"context"
	"io"
)

// AccountReader reads who and where mu is. *client.Client satisfies it.
type AccountReader interface {
	GetHealth(ctx context.Context) (io.ReadCloser, error)
	GetAuthenticatedUser(ctx context.Context) (io.ReadCloser, error)
	ListMyTenants(ctx context.Context) (io.ReadCloser, error)
}

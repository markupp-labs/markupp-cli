package cli

import (
	"context"
	"io"
)

// TenantAdmin lists and creates the tenants of the account. *client.Client
// satisfies it.
type TenantAdmin interface {
	ListMyTenants(ctx context.Context) (io.ReadCloser, error)
	CreateTenant(ctx context.Context, name string) (io.ReadCloser, error)
}

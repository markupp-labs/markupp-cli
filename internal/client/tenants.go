package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// ListMyTenants returns the MyTenantResponse array of the tenants of the account.
func (c *Client) ListMyTenants(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("me", "tenants")})
}

// CreateTenant creates a tenant owned by the account and returns its TenantResponse.
func (c *Client) CreateTenant(ctx context.Context, name string) (io.ReadCloser, error) {
	body, err := encodeAdminBody(schema.TenantRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return c.open(ctx, request{method: http.MethodPost, path: endpoint("tenants"), body: body})
}

// adminRequest lists the request schemas of the administration routes.
type adminRequest interface {
	schema.TenantRequest | schema.RoleRequest | schema.InviteRequest
}

// encodeAdminBody marshals an administration request into the JSON body of a request.
func encodeAdminBody[T adminRequest](value T) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, failure.Wrap(failure.Internal, "encoding the request body", err)
	}
	return body, nil
}

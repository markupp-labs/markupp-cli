package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// requestBody lists the schema types a request can carry as its JSON body.
type requestBody interface {
	schema.VaultNameRequest | schema.VaultMemberRequest
}

// openJSON encodes payload as the body of r, runs it and returns the response body.
func openJSON[T requestBody](ctx context.Context, c *Client, r request, payload T) (io.ReadCloser, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, failure.Wrap(failure.Internal, fmt.Sprintf("encoding the body of %s as %T", r.subject(), payload), err)
	}
	r.body = raw
	return c.open(ctx, r)
}

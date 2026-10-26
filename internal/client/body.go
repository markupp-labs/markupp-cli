package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// requestBody lists the schema structs sent as a request body. A new body type joins
// the union.
type requestBody interface {
	schema.NoteCreateRequest | schema.NoteUpdateRequest |
		schema.VaultNameRequest | schema.VaultMemberRequest | schema.PublicationRequest
}

// marshalBody encodes a schema request struct as the JSON body of a request.
func marshalBody[T requestBody](value T) ([]byte, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return nil, failure.Wrap(failure.Internal, fmt.Sprintf("encoding request body %T", value), err)
	}
	return body, nil
}

// openJSON encodes payload as the body of r, runs it and returns the response body.
func openJSON[T requestBody](ctx context.Context, c *Client, r request, payload T) (io.ReadCloser, error) {
	raw, err := marshalBody(payload)
	if err != nil {
		return nil, err
	}
	r.body = raw
	return c.open(ctx, r)
}

package oauth

import (
	"context"
	"errors"
)

var errNotImplemented = errors.New("not implemented")

// Discover reads the metadata of the server at server.
func (c *Client) Discover(_ context.Context, _ string) (Metadata, error) {
	return Metadata{}, errNotImplemented
}

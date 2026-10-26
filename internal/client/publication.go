package client

import (
	"context"
	"errors"
	"io"
)

var errPublicationPending = errors.New("not implemented")

// GetPublication returns the publication of the vault id.
func (c *Client) GetPublication(context.Context, string) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// PublishVault publishes the vault id under slug, or changes its slug and title.
func (c *Client) PublishVault(context.Context, string, string, string) (io.ReadCloser, error) {
	return nil, errPublicationPending
}

// UnpublishVault takes the vault id out of the public reads.
func (c *Client) UnpublishVault(context.Context, string) error { return errPublicationPending }

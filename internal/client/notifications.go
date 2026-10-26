package client

import (
	"context"
	"errors"
	"io"
)

// Notifications returns the notification feed of the account.
func (c *Client) Notifications(context.Context) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

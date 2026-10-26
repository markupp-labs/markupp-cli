package cli

import (
	"context"
	"io"
)

// NotificationReader reads the notification feed of the account.
type NotificationReader interface {
	Notifications(ctx context.Context) (io.ReadCloser, error)
}

package cli

import (
	"context"
	"io"
)

// InviteInbox answers the invites addressed to the account. *client.Client
// satisfies it.
type InviteInbox interface {
	ListMyInvites(ctx context.Context) (io.ReadCloser, error)
	AcceptInvite(ctx context.Context, id string) (io.ReadCloser, error)
	DeclineInvite(ctx context.Context, id string) error
}

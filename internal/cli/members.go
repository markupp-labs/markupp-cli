package cli

import (
	"context"
	"io"
)

// MemberAdmin manages the members and the invites of the tenant. *client.Client
// satisfies it.
type MemberAdmin interface {
	ListMembers(ctx context.Context) (io.ReadCloser, error)
	ChangeMemberRole(ctx context.Context, id, role string) error
	RemoveMember(ctx context.Context, id string) error
	InviteMember(ctx context.Context, email, role string) (io.ReadCloser, error)
	ListTenantInvites(ctx context.Context) (io.ReadCloser, error)
	RevokeInvite(ctx context.Context, id string) error
}

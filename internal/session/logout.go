package session

import (
	"context"

	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// Logout revokes the stored refresh token on its server (RFC 7009) and removes the
// credential, under the config lock so it cannot cross a refresh of another process.
// The credential is removed even when the revocation fails; the error then says so and
// points at the web panel. removed is false when there was no login to remove.
func (a Authenticator) Logout(ctx context.Context) (removed bool, err error) {
	settings, err := a.Store.Load()
	if err != nil || settings.Auth == nil {
		return false, err
	}
	issuer, revokeErr, err := a.removeLogin(ctx)
	if err != nil || revokeErr == nil {
		return issuer != "", err
	}
	return true, a.revokeFailure(issuer, revokeErr)
}

// removeLogin revokes and removes the credential under the lock. issuer is empty when
// another process removed it first.
func (a Authenticator) removeLogin(ctx context.Context) (issuer string, revokeErr, err error) {
	err = withLock(ctx, a.Store, func(settings *config.Settings) error {
		if settings.Auth == nil {
			return nil
		}
		issuer = settings.Auth.Issuer
		revokeErr = a.revoke(ctx, *settings.Auth)
		settings.Auth = nil
		return a.Store.Save(*settings)
	})
	return issuer, revokeErr, err
}

func (a Authenticator) revoke(ctx context.Context, cred config.Credential) error {
	meta, err := a.OAuth.Discover(ctx, cred.Issuer)
	if err != nil {
		return err
	}
	return a.OAuth.Revoke(ctx, meta, cred.RefreshToken)
}

func (a Authenticator) revokeFailure(issuer string, cause error) error {
	path, _ := a.Store.Path()
	return &failure.Error{
		Kind:    failure.KindOf(cause),
		Subject: issuer,
		Message: "the login was removed from " + path + ", but revoking it on the server failed: " + cause.Error(),
		Hint:    "revoke the authorization of " + ClientID + " in the web panel",
		Err:     cause,
	}
}

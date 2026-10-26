package session

import "context"

// Logout revokes and removes the stored login.
func (a Authenticator) Logout(context.Context) (bool, error) { return false, errNotImplemented }

package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// Source is the client.TokenSource of the login stored in the config file. It hands
// out the stored access token while more than ExpiryMargin remains and refreshes it
// otherwise, under the config lock: it re-reads the file once it holds the lock, so
// when another process already rotated the refresh token it uses that result instead
// of spending the refresh token a second time.
type Source struct {
	store  Store
	oauth  *oauth.Client
	server string
	now    Clock
}

// NewSource returns the Source of the login stored for server. A login stored for
// another origin is refused, so its token never travels to server. now nil means
// time.Now.
func NewSource(store Store, auth *oauth.Client, server string, now Clock) *Source {
	if now == nil {
		now = time.Now
	}
	return &Source{store: store, oauth: auth, server: server, now: now}
}

// Token returns the access token to send, refreshing it when it is about to expire.
func (s *Source) Token(ctx context.Context) (string, error) {
	settings, err := s.store.Load()
	if err != nil {
		return "", err
	}
	cred, err := s.credential(settings)
	if err != nil {
		return "", err
	}
	if s.fresh(cred) {
		return cred.AccessToken, nil
	}
	return s.renew(ctx)
}

// Rejected refreshes after the server answered 401 to token, once, and only when token
// is still the one on disk: a token that is not the stored one is not this login's to
// refresh, so there is no retry.
func (s *Source) Rejected(ctx context.Context, token string) (string, bool, error) {
	var fresh string
	err := s.locked(ctx, func(settings *config.Settings) error {
		cred, err := s.credential(*settings)
		if err != nil || cred.AccessToken != token {
			return err
		}
		fresh, err = s.rotate(ctx, settings, cred)
		return err
	})
	return fresh, fresh != "", err
}

// renew refreshes under the lock, unless another process did while this one waited.
func (s *Source) renew(ctx context.Context) (string, error) {
	var token string
	err := s.locked(ctx, func(settings *config.Settings) error {
		cred, err := s.credential(*settings)
		if err != nil {
			return err
		}
		if s.fresh(cred) {
			token = cred.AccessToken
			return nil
		}
		token, err = s.rotate(ctx, settings, cred)
		return err
	})
	return token, err
}

// locked runs fn on the settings read while holding the config lock.
func (s *Source) locked(ctx context.Context, fn func(*config.Settings) error) error {
	return withLock(ctx, s.store, fn)
}

// rotate spends the refresh token and saves the new pair before returning the new
// access token, so a crash never loses the only valid refresh token.
func (s *Source) rotate(ctx context.Context, settings *config.Settings, cred config.Credential) (string, error) {
	next, err := s.refresh(ctx, cred)
	if oauth.IsInvalidGrant(err) {
		return "", s.forget(settings, cred, err)
	}
	if err != nil {
		return "", err
	}
	settings.Auth = &next
	if err := s.store.Save(*settings); err != nil {
		return "", err
	}
	return next.AccessToken, nil
}

func (s *Source) refresh(ctx context.Context, cred config.Credential) (config.Credential, error) {
	meta, err := s.oauth.Discover(ctx, cred.Issuer)
	if err != nil {
		return config.Credential{}, err
	}
	tokens, err := s.oauth.Refresh(ctx, meta, cred.RefreshToken)
	if err != nil {
		return config.Credential{}, err
	}
	next := newCredential(cred.Issuer, tokens, s.now())
	if next.Scope == "" {
		next.Scope = cred.Scope
	}
	return next, nil
}

// forget removes a login the server refused: only a new login can replace it.
func (s *Source) forget(settings *config.Settings, cred config.Credential, cause error) error {
	settings.Auth = nil
	refused := failure.Wrap(failure.Unauthenticated, cred.Issuer, cause).
		WithHint("the stored login was removed; run mu login " + cred.Issuer)
	return errors.Join(refused, s.store.Save(*settings))
}

func (s *Source) credential(settings config.Settings) (config.Credential, error) {
	cred := settings.Auth
	if cred == nil || cred.AccessToken == "" {
		return config.Credential{}, failure.New(failure.Unauthenticated, s.server, "not logged in").
			WithHint("run mu login " + s.server + ", or set " + config.EnvToken)
	}
	if !oauth.SameOrigin(cred.Issuer, s.server) {
		message := fmt.Sprintf("the stored login is for %q, not for this server", cred.Issuer)
		return config.Credential{}, failure.New(failure.Unauthenticated, s.server, message).
			WithHint("run mu login " + s.server)
	}
	return *cred, nil
}

func (s *Source) fresh(cred config.Credential) bool {
	return s.now().Add(ExpiryMargin).Before(cred.ExpiresAt)
}

// newCredential stores tokens issued by issuer at now.
func newCredential(issuer string, tokens oauth.Tokens, now time.Time) config.Credential {
	return config.Credential{
		Issuer:       issuer,
		ClientID:     ClientID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    now.Add(tokens.ExpiresIn).UTC(),
		Scope:        tokens.Scope,
	}
}

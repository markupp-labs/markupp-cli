package session

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
	"github.com/markupp-labs/markupp-cli/internal/config"
)

// Authenticator logs mu in through the browser and out again. Open is the best effort
// browser opener, nil to only print the URL; Notice receives the URL and progress,
// stderr in mu; Timeout zero means oauth.CallbackTimeout.
type Authenticator struct {
	OAuth   *oauth.Client
	Store   Store
	Now     Clock
	Notice  io.Writer
	Open    func(url string) error
	Timeout time.Duration
}

// flow is what only one login knows: its PKCE pair and its state.
type flow struct {
	pkce  oauth.PKCE
	state string
}

// Login runs the authorization code flow with PKCE through the browser, against the
// built in client of the server, and stores the credential and server in the config
// file. Nothing is stored unless the whole flow succeeds.
func (a Authenticator) Login(ctx context.Context, server string) (config.Credential, error) {
	server, err := config.NormalizeURL(server)
	if err != nil {
		return config.Credential{}, err
	}
	meta, err := a.OAuth.Discover(ctx, server)
	if err != nil {
		return config.Credential{}, err
	}
	cred, err := a.authorize(ctx, meta)
	if err != nil {
		return config.Credential{}, err
	}
	return cred, a.save(ctx, server, cred)
}

// authorize sends the browser to the consent and trades the code it brings back.
func (a Authenticator) authorize(ctx context.Context, meta oauth.Metadata) (config.Credential, error) {
	login, err := newFlow()
	if err != nil {
		return config.Credential{}, err
	}
	callback, err := oauth.Listen(ctx, login.state)
	if err != nil {
		return config.Credential{}, err
	}
	defer func() { _ = callback.Close() }()
	code, err := a.awaitBrowser(ctx, meta, callback, login)
	if err != nil {
		return config.Credential{}, err
	}
	tokens, err := a.OAuth.ExchangeCode(ctx, meta, code, callback.RedirectURI(), login.pkce.Verifier)
	if err != nil {
		return config.Credential{}, err
	}
	return newCredential(meta.Issuer, tokens, a.now()), nil
}

func newFlow() (flow, error) {
	pkce, err := oauth.NewPKCE()
	if err != nil {
		return flow{}, err
	}
	state, err := oauth.RandomToken()
	if err != nil {
		return flow{}, err
	}
	return flow{pkce: pkce, state: state}, nil
}

// awaitBrowser prints the authorize URL, tries to open it and waits for the callback.
// A browser that does not open is not an error: the URL is on stderr.
func (a Authenticator) awaitBrowser(ctx context.Context, meta oauth.Metadata, callback *oauth.Callback, login flow) (string, error) {
	target := a.OAuth.AuthorizeURL(meta, callback.RedirectURI(), login.pkce, login.state)
	_, _ = fmt.Fprintf(a.notice(), "Opening the browser to log in to %s. If it does not open, visit:\n%s\n", meta.Issuer, target)
	if a.Open != nil {
		_ = a.Open(target)
	}
	return callback.Wait(ctx, a.timeout())
}

func (a Authenticator) save(ctx context.Context, server string, cred config.Credential) error {
	return withLock(ctx, a.Store, func(_ context.Context, settings *config.Settings) error {
		settings.URL = server
		settings.Auth = &cred
		return a.Store.Save(*settings)
	})
}

func (a Authenticator) now() time.Time {
	if a.Now == nil {
		return time.Now()
	}
	return a.Now()
}

func (a Authenticator) notice() io.Writer {
	if a.Notice == nil {
		return io.Discard
	}
	return a.Notice
}

func (a Authenticator) timeout() time.Duration {
	if a.Timeout <= 0 {
		return oauth.CallbackTimeout
	}
	return a.Timeout
}

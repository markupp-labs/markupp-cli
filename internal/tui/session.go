package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// startRefresh reloads the current listing every cfg.Refresh until ctx ends.
func (s *shell) startRefresh() {
	if s.cfg.Refresh <= 0 {
		return
	}
	ticker := time.NewTicker(s.cfg.Refresh)
	go func() {
		defer ticker.Stop()
		for s.waitTick(ticker) {
			s.app.QueueUpdateDraw(s.tick)
		}
	}()
}

func (s *shell) waitTick(ticker *time.Ticker) bool {
	select {
	case <-s.ctx.Done():
		return false
	case <-ticker.C:
		return true
	}
}

// tick reloads the top view unless the refresh is paused or a dialog is open.
func (s *shell) tick() {
	s.session = s.cfg.session()
	s.refreshChrome()
	if s.paused || s.overlay {
		return
	}
	if r, ok := s.stack.top().(Refreshable); ok && r.Refreshable() {
		s.reload()
	}
}

// loadAccount reads the email of the account for the header.
func (s *shell) loadAccount() {
	if s.ports.Account == nil || s.scope.PublicSlug != "" {
		return
	}
	s.async(s.readAccount)
}

func (s *shell) readAccount(ctx context.Context) (func(), error) {
	body, err := s.ports.Account.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil, err
	}
	user, err := decodeBody[schema.User](body)
	return func() { s.account = user.Email; s.refreshChrome() }, err
}

// switchScope connects as scope, then opens the view next returns.
func (s *shell) switchScope(scope Scope, name string, next func() View) {
	s.async(func(ctx context.Context) (func(), error) {
		ports, err := s.cfg.Connect(ctx, scope)
		return func() { s.enterScope(scope, name, ports, next) }, err
	})
}

func (s *shell) enterScope(scope Scope, name string, ports Ports, next func() View) {
	s.scope, s.tenantName, s.ports, s.vault = scope, name, ports, ""
	s.flashInfo(fmt.Sprintf("switched to %s", s.scopeLabel()))
	s.loadAccount()
	s.reset(next())
}

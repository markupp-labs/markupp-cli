package tui

import (
	"context"
	"time"
)

// startRefresh reloads the current listing every cfg.Refresh until ctx ends.
func (s *shell) startRefresh() {}

func (s *shell) waitTick(ticker *time.Ticker) bool { return *new(bool) }

// tick reloads the top view unless the refresh is paused or a dialog is open.
func (s *shell) tick() {}

// loadAccount reads the email of the account for the header.
func (s *shell) loadAccount() {}

func (s *shell) readAccount(ctx context.Context) (func(), error) { return *new(func()), *new(error) }

// switchScope connects as scope, then opens the view next returns.
func (s *shell) switchScope(scope Scope, name string, next func() View) {}

func (s *shell) enterScope(scope Scope, name string, ports Ports, next func() View) {}

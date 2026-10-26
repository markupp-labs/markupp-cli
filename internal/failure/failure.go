// Package failure classifies every error the CLI can return and maps it to the exit
// code that is part of the public contract of mu.
package failure

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// Kind is the class of a failure. Each Kind has one exit code.
type Kind int

// The kinds, in the order of the exit code table.
const (
	Internal Kind = iota
	Usage
	Unauthenticated
	Forbidden
	NotFound
	Conflict
	RateLimited
	Unavailable
	Interrupted
)

var kindNames = [...]string{
	"internal", "usage", "unauthenticated", "forbidden", "not_found",
	"conflict", "rate_limited", "unavailable", "interrupted",
}

// Error is a classified failure. Subject names the offending value (an argument, a
// flag, a request), Code is the server error code when there is one, Scope the OAuth
// scope the server asked for and RetryAfter the wait a rate limit asked for.
type Error struct {
	Kind       Kind
	Code       string
	Message    string
	Subject    string
	Hint       string
	Scope      string
	RetryAfter time.Duration
	Err        error
}

// Error renders "subject: message (code)", leaving out the empty parts.
func (e *Error) Error() string {
	text := e.Message
	if text == "" && e.Err != nil {
		text = e.Err.Error()
	}
	if e.Subject != "" {
		text = e.Subject + ": " + text
	}
	if e.Code != "" {
		text += " (" + e.Code + ")"
	}
	return text
}

// Unwrap exposes the cause to errors.Is and errors.As.
func (e *Error) Unwrap() error { return e.Err }

// ExitCode returns the process exit code for the kind: 0 is success, 130 is an
// interrupt, and the kinds in between count up from 1.
func (k Kind) ExitCode() int {
	if k == Interrupted {
		return 130
	}
	return int(k) + 1
}

// String returns the kind name used in logs and tests.
func (k Kind) String() string {
	if k < 0 || int(k) >= len(kindNames) {
		return "kind(" + strconv.Itoa(int(k)) + ")"
	}
	return kindNames[k]
}

// New builds an Error of kind about subject.
func New(kind Kind, subject, message string) *Error {
	return &Error{Kind: kind, Subject: subject, Message: message}
}

// Wrap builds an Error of kind about subject that keeps err as the cause.
func Wrap(kind Kind, subject string, err error) *Error {
	return &Error{Kind: kind, Subject: subject, Err: err}
}

// WithHint returns a copy of e carrying hint, the next step offered to the user.
func (e *Error) WithHint(hint string) *Error {
	hinted := *e
	hinted.Hint = hint
	return &hinted
}

// KindOf classifies any error: a wrapped *Error keeps its kind, a canceled context is
// an interrupt and everything else is internal.
func KindOf(err error) Kind {
	var classified *Error
	if errors.As(err, &classified) {
		return classified.Kind
	}
	if errors.Is(err, context.Canceled) {
		return Interrupted
	}
	return Internal
}

// ExitCode returns the exit code for err, 0 for nil.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	return KindOf(err).ExitCode()
}

// HintOf returns the hint carried by err, empty when there is none.
func HintOf(err error) string {
	var classified *Error
	if errors.As(err, &classified) {
		return classified.Hint
	}
	return ""
}

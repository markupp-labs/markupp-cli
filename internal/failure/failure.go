// Package failure classifies every error the CLI can return and maps it to the exit
// code that is part of the public contract of mu.
package failure

import "time"

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

// Error is a classified failure. Subject names the offending value (an argument, a
// flag, a request), Code is the server error code when there is one.
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

func (e *Error) Error() string { return "" }

func (e *Error) Unwrap() error { return nil }

// ExitCode returns the process exit code for the kind.
func (k Kind) ExitCode() int { return 0 }

// String returns the kind name used in logs and tests.
func (k Kind) String() string { return "" }

// New builds an Error of kind about subject.
func New(kind Kind, subject, message string) *Error { return nil }

// Wrap builds an Error of kind about subject that keeps err as the cause.
func Wrap(kind Kind, subject string, err error) *Error { return nil }

// WithHint returns a copy of e carrying hint.
func (e *Error) WithHint(hint string) *Error { return e }

// KindOf classifies any error.
func KindOf(err error) Kind { return Internal }

// ExitCode returns the exit code for err, 0 for nil.
func ExitCode(err error) int { return 0 }

// HintOf returns the hint carried by err, empty when there is none.
func HintOf(err error) string { return "" }

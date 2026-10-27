package failure

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestKindExitCodeTable(t *testing.T) {
	table := map[Kind]int{
		Internal: 1, Usage: 2, Unauthenticated: 3, Forbidden: 4, NotFound: 5,
		Conflict: 6, RateLimited: 7, Unavailable: 8, Interrupted: 130,
	}
	for kind, want := range table {
		if got := kind.ExitCode(); got != want {
			t.Errorf("%s.ExitCode() = %d, want %d", kind, got, want)
		}
	}
}

func TestKindStringNamesEveryKind(t *testing.T) {
	names := map[Kind]string{
		Internal: "internal", Usage: "usage", Unauthenticated: "unauthenticated",
		Forbidden: "forbidden", NotFound: "not_found", Conflict: "conflict",
		RateLimited: "rate_limited", Unavailable: "unavailable", Interrupted: "interrupted",
	}
	for kind, want := range names {
		if got := kind.String(); got != want {
			t.Errorf("Kind(%d).String() = %q, want %q", int(kind), got, want)
		}
	}
	if got := Kind(99).String(); got != "kind(99)" {
		t.Errorf("unknown kind String() = %q, want kind(99)", got)
	}
}

func TestErrorMessageNamesSubjectAndCode(t *testing.T) {
	err := New(Conflict, "PUT /v1/notes/n1", "note changed on the server")
	err.Code = "conflict"
	want := "PUT /v1/notes/n1: note changed on the server (conflict)"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorMessageWithoutSubjectOrCode(t *testing.T) {
	err := New(Usage, "", "missing argument")
	if got := err.Error(); got != "missing argument" {
		t.Errorf("Error() = %q, want %q", got, "missing argument")
	}
}

func TestWrapKeepsCauseAndUsesItAsMessage(t *testing.T) {
	cause := errors.New("connection refused")
	err := Wrap(Unavailable, "GET /v1/vaults", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(Wrap(...), cause) = false, want true")
	}
	if got := err.Error(); got != "GET /v1/vaults: connection refused" {
		t.Errorf("Error() = %q", got)
	}
}

func TestKindOfClassifiesWrappedErrors(t *testing.T) {
	wrapped := fmt.Errorf("loading: %w", New(NotFound, "note n1", "not found"))
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"wrapped failure", wrapped, NotFound},
		{"plain error", errors.New("boom"), Internal},
		{"canceled context", fmt.Errorf("get: %w", context.Canceled), Interrupted},
	}
	for _, tc := range cases {
		if got := KindOf(tc.err); got != tc.want {
			t.Errorf("%s: KindOf = %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestExitCodeOfNilIsZero(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Errorf("ExitCode(nil) = %d, want 0", got)
	}
	if got := ExitCode(New(Forbidden, "x", "y")); got != 4 {
		t.Errorf("ExitCode(forbidden) = %d, want 4", got)
	}
}

func TestWithHintDoesNotMutateTheOriginal(t *testing.T) {
	base := New(Unauthenticated, "token", "expired")
	hinted := base.WithHint("run mu login")
	if base.Hint != "" {
		t.Errorf("original hint = %q, want empty", base.Hint)
	}
	if got := HintOf(fmt.Errorf("x: %w", hinted)); got != "run mu login" {
		t.Errorf("HintOf = %q, want %q", got, "run mu login")
	}
	if got := HintOf(errors.New("plain")); got != "" {
		t.Errorf("HintOf(plain) = %q, want empty", got)
	}
}

func TestWrapReadTurnsACanceledReadIntoAnInterrupt(t *testing.T) {
	canceled := WrapRead(Unavailable, "reading the server response", fmt.Errorf("read tcp: %w", context.Canceled))
	if canceled.Kind != Interrupted || !errors.Is(canceled, context.Canceled) {
		t.Errorf("canceled read = %v (kind %s), want an interrupt keeping the cause", canceled, canceled.Kind)
	}
	broken := WrapRead(Unavailable, "reading the server response", errors.New("connection reset"))
	if broken.Kind != Unavailable || broken.Subject != "reading the server response" {
		t.Errorf("broken read = %v (kind %s), want unavailable", broken, broken.Kind)
	}
}

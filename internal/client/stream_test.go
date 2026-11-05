package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// trackedBody records whether it was closed.
type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func collect(t *testing.T, body string) ([]string, error) {
	t.Helper()
	tracked := &trackedBody{Reader: strings.NewReader(body)}
	var got []string
	var lastErr error
	for element, err := range Elements(tracked) {
		if err != nil {
			lastErr = err
			break
		}
		got = append(got, string(element))
	}
	if !tracked.closed {
		t.Errorf("Elements(%q) did not close the body", body)
	}
	return got, lastErr
}

func TestElementsYieldsEachArrayElement(t *testing.T) {
	got, err := collect(t, `[{"id":"n1"}, {"id":"n2","x":[1,2]}]`)
	if err != nil {
		t.Fatalf("Elements: %v", err)
	}
	if strings.Join(got, "|") != `{"id":"n1"}|{"id":"n2","x":[1,2]}` {
		t.Errorf("elements = %q", got)
	}
}

func TestElementsOfNullOrEmptyIsEmpty(t *testing.T) {
	for _, body := range []string{"null", "[]", " [ ] "} {
		got, err := collect(t, body)
		if err != nil || len(got) != 0 {
			t.Errorf("Elements(%q) = %q, %v, want nothing", body, got, err)
		}
	}
}

func TestElementsRejectsNonArrays(t *testing.T) {
	for _, body := range []string{`{"id":"n1"}`, `"x"`, `[{"id":`, ``} {
		_, err := collect(t, body)
		if failure.KindOf(err) != failure.Internal || err == nil {
			t.Errorf("Elements(%q) error = %v, want an internal malformed response", body, err)
		}
	}
}

func TestElementsStopsAndClosesWhenTheConsumerBreaks(t *testing.T) {
	tracked := &trackedBody{Reader: strings.NewReader(`[1,2,3]`)}
	for element := range Elements(tracked) {
		if string(element) != "1" {
			t.Fatalf("first element = %s", element)
		}
		break
	}
	if !tracked.closed {
		t.Error("body not closed after break")
	}
}

func TestPageAppliesOnlyGivenValues(t *testing.T) {
	cases := map[Page]string{
		{}:                      "",
		{Limit: 5}:              "limit=5",
		{Offset: 20}:            "offset=20",
		{Limit: 100, Offset: 1}: "limit=100&offset=1",
	}
	for page, want := range cases {
		query := url.Values{}
		page.apply(query)
		if got := query.Encode(); got != want {
			t.Errorf("%+v.apply = %q, want %q", page, got, want)
		}
	}
}

func TestSetNonEmptySkipsBlankValues(t *testing.T) {
	query := url.Values{}
	setNonEmpty(query, "vault", "")
	setNonEmpty(query, "q", "kubernetes")
	if got := query.Encode(); got != "q=kubernetes" {
		t.Errorf("query = %q, want q=kubernetes", got)
	}
}

// interruptedBody serves prefix, then fails as a body read does after its context
// was canceled.
type interruptedBody struct {
	prefix *strings.Reader
}

func (b *interruptedBody) Read(p []byte) (int, error) {
	if b.prefix.Len() > 0 {
		return b.prefix.Read(p)
	}
	return 0, fmt.Errorf("read tcp 127.0.0.1:80: %w", context.Canceled)
}

func (b *interruptedBody) Close() error { return nil }

func TestElementsInterruptedMidStreamIsAnInterrupt(t *testing.T) {
	body := &interruptedBody{prefix: strings.NewReader(`[{"id":"n1"},{"id":`)}
	var last error
	for _, err := range Elements(body) {
		last = err
	}
	if got := failure.KindOf(last); got != failure.Interrupted {
		t.Errorf("interrupted stream error = %v (kind %s), want interrupted", last, got)
	}
}

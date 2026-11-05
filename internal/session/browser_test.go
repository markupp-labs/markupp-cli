package session

import (
	"slices"
	"testing"
)

func TestBrowserCommandPerPlatform(t *testing.T) {
	target := "https://markupp.example.com/oauth/authorize?a=1&b=2"
	cases := map[string][]string{
		"linux":   {"xdg-open", target},
		"freebsd": {"xdg-open", target},
		"darwin":  {"open", target},
		"windows": {"cmd", "/c", "start", "", "https://markupp.example.com/oauth/authorize?a=1^&b=2"},
	}
	for goos, want := range cases {
		name, args := browserCommand(goos, target)
		if got := append([]string{name}, args...); !slices.Equal(got, want) {
			t.Errorf("browserCommand(%s) = %q, want %q", goos, got, want)
		}
	}
}

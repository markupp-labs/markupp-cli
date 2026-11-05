package cli

import (
	"errors"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestShortTimeFitsTheColumn(t *testing.T) {
	cases := map[string]string{
		"2026-10-20T13:45:12.123456Z": "2026-10-20T13:45:12Z",
		"2026-05-06T07:08:09.5-03:00": "2026-05-06T10:08:09Z",
		"2026-01-02T03:04:05Z":        "2026-01-02T03:04:05Z",
		"":                            "",
		"yesterday":                   "yesterday",
	}
	for in, want := range cases {
		got := shortTime(in)
		if got != want {
			t.Errorf("shortTime(%q) = %q, want %q", in, got, want)
		}
		if in != "yesterday" && len(got) > timeColumn {
			t.Errorf("shortTime(%q) = %q, wider than %d", in, got, timeColumn)
		}
	}
}

func TestCheckRoleNamesTheSubjectAndTheValue(t *testing.T) {
	if err := checkRole("ROLE", "viewer"); err != nil {
		t.Errorf("checkRole(viewer) = %v, want nil", err)
	}
	err := checkRole("--role", "")
	var failed *failure.Error
	if failure.KindOf(err) != failure.Usage || !errors.As(err, &failed) || failed.Subject != "--role" {
		t.Errorf("checkRole(\"\") = %#v, want a usage error about --role", err)
	}
}

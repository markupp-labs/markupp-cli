package cli

import (
	"strings"
	"testing"
)

func TestVaultCommandsWithoutAServerExplainHowToConfigureOne(t *testing.T) {
	for _, args := range [][]string{{"notes", "list"}, {"graph", "show"}} {
		run := newTestMu().run(t, args...)
		run.expectCode(t, 2)
		if !strings.Contains(run.stderr, "no Markupp server configured") {
			t.Errorf("mu %s: stderr = %q, want the server hint", strings.Join(args, " "), run.stderr)
		}
	}
}

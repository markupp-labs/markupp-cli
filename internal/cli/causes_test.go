package cli

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestInvalidSinceKeepsTheParseError(t *testing.T) {
	_, err := (&notesEditCmd{since: "yesterday"}).version()
	var parseErr *time.ParseError
	if !errors.As(err, &parseErr) {
		t.Errorf("--since error = %v, want it to wrap the *time.ParseError", err)
	}
}

func TestInvalidCommunityIDKeepsTheParseError(t *testing.T) {
	_, err := parseCommunityID(&cobra.Command{Use: "communities"}, "seven")
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Errorf("CID error = %v, want it to wrap strconv.ErrSyntax", err)
	}
}

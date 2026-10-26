// Package outputtest holds the golden file helper shared by the tests of every package
// that renders output. Run the tests with -update to rewrite the golden files:
//
//	go test ./internal/cli -update
package outputtest

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files under testdata")

// Golden compares got with testdata/<name>.golden of the calling package, or writes
// that file when the tests run with -update.
func Golden(t testing.TB, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		writeGolden(t, path, got)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading golden file %s (run with -update to create it): %v", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output differs from %s (run with -update to accept)\n--- got\n%s\n--- want\n%s", path, got, want)
	}
}

func writeGolden(t testing.TB, path string, got []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, got, 0o600); err != nil {
		t.Fatalf("writing golden file %s: %v", path, err)
	}
}

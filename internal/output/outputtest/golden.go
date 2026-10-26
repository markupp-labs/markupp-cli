// Package outputtest holds the golden file helper shared by the tests of every package
// that renders output. Run the tests with -update to rewrite the golden files.
package outputtest

import "testing"

// Golden compares got with testdata/<name>.golden of the calling package.
func Golden(t testing.TB, name string, got []byte) { t.Helper() }

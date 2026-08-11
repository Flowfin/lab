//go:build integration_hardware

package one

import "testing"

// A fixture. Nothing compiles it: the go tool ignores every directory named
// testdata, and this file is behind the harness constraint as well. What it
// exists for is its name, which is how a record's declaration is checked
// against what the directory registers.
func TestSequentialWriteOnASpinningDisk(t *testing.T) {
	t.Skip("a fixture")
}

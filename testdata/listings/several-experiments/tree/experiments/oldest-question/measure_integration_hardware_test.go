//go:build integration_hardware

package oldestquestion

import "testing"

// A fixture. It exists so that this tree reads the same way under both verbs:
// the record declares a spinning disk and this is the file that registers a
// test with the harness, so a walk of this tree refuses nothing either.
func TestSequentialWriteOnASpinningDisk(t *testing.T) {
	t.Skip("a fixture")
}

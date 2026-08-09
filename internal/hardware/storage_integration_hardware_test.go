//go:build integration_hardware

package hardware

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestSequentialWriteOnRealStorage is the first test in this harness and the
// shape every later one follows: it names the hardware it needs in its first
// line, it measures, and it reports the measurement rather than asserting a
// number.
//
// It asserts nothing about how fast the storage is. A threshold would be a
// claim about one machine written into a test that runs on another, which is
// exactly the mistake this harness exists to keep out of the default run. What
// it fails on is the write not working at all.
//
// The measurement is about the machine it ran on. It is not this suite's
// result and it is not a property of this repository.
func TestSequentialWriteOnRealStorage(t *testing.T) {
	Needs(t, "the storage device the temporary directory is on, with room for 32 MiB")

	const size = 32 << 20

	dir := t.TempDir()
	path := filepath.Join(dir, "block")

	block := make([]byte, 1<<20)
	for i := range block {
		block[i] = byte(i)
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("cannot create %s: %v", path, err)
	}

	start := time.Now()
	for written := 0; written < size; written += len(block) {
		if _, err := file.Write(block); err != nil {
			file.Close()
			t.Fatalf("cannot write to %s: %v", path, err)
		}
	}
	if err := file.Sync(); err != nil {
		file.Close()
		t.Fatalf("cannot flush %s to the device: %v", path, err)
	}
	elapsed := time.Since(start)
	if err := file.Close(); err != nil {
		t.Fatalf("cannot close %s: %v", path, err)
	}

	t.Logf("wrote and flushed %d MiB in %s, which is %.1f MiB per second on this machine",
		size>>20, elapsed, float64(size>>20)/elapsed.Seconds())
}

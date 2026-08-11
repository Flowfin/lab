// Command main builds a tree of experiment records in a temporary directory and
// times two passes over it: one that only walks the directories, and one that
// also opens every record and reads its bytes.
//
// It is an experiment's prototype rather than part of the runner. Nothing in
// cmd/ or internal/ imports it, it is not run by any check, and it writes only
// inside the temporary directory it creates and removes.
//
// Run it from a checkout:
//
//	go run ./experiments/reading-a-tree-of-records
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// The shape of the tree. A thousand is the number the question names, and the
// record is the size the records in this repository are: the longest of them is
// under four kilobytes and most are near one.
const (
	experiments = 1000
	rounds      = 7
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "measure: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := os.MkdirTemp("", "reading-a-tree-of-records-")
	if err != nil {
		return fmt.Errorf("cannot make a temporary directory: %w", err)
	}
	defer os.RemoveAll(root)

	bytes, err := build(root)
	if err != nil {
		return err
	}

	fmt.Printf("%s/%s, %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
	fmt.Printf("%d experiments, %d bytes of records, %d rounds\n", experiments, bytes, rounds)

	walking, err := fastest(func() (int, error) { return walk(root, false) })
	if err != nil {
		return err
	}
	reading, err := fastest(func() (int, error) { return walk(root, true) })
	if err != nil {
		return err
	}

	fmt.Printf("walking the directories:        %v\n", walking)
	fmt.Printf("walking and reading every file: %v\n", reading)
	fmt.Printf("reading costs %.2f times the walk\n", float64(reading)/float64(walking))
	return nil
}

// build writes the tree and returns how many bytes of records it holds.
func build(root string) (int, error) {
	record := recordOfAboutAKilobyte()
	total := 0

	for i := 0; i < experiments; i++ {
		dir := filepath.Join(root, "experiments", fmt.Sprintf("experiment-%04d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, fmt.Errorf("cannot make %s: %w", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "EXPERIMENT.md"), []byte(record), 0o644); err != nil {
			return 0, fmt.Errorf("cannot write into %s: %w", dir, err)
		}
		total += len(record)
	}
	return total, nil
}

// walk makes one pass over the tree, reading the files or not, and returns how
// many it visited. The count is returned and printed nowhere on purpose: it is
// there so that the compiler cannot decide the reading was pointless and remove
// it, which is the way a measurement like this reports a time for work that
// never happened.
func walk(root string, read bool) (int, error) {
	seen := 0
	dir := filepath.Join(root, "experiments")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("cannot read %s: %w", dir, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record := filepath.Join(dir, entry.Name(), "EXPERIMENT.md")
		if !read {
			info, err := os.Stat(record)
			if err != nil {
				return 0, fmt.Errorf("cannot stat %s: %w", record, err)
			}
			seen += int(info.Size() % 2)
			continue
		}
		data, err := os.ReadFile(record)
		if err != nil {
			return 0, fmt.Errorf("cannot read %s: %w", record, err)
		}
		seen += len(data) % 2
	}
	return seen, nil
}

// fastest runs a pass several times and returns the shortest time it took. The
// shortest is the one least contaminated by whatever else the machine was
// doing, and a mean over a machine that is doing other things measures the
// other things.
func fastest(pass func() (int, error)) (time.Duration, error) {
	best := time.Duration(0)

	for i := 0; i < rounds; i++ {
		start := time.Now()
		if _, err := pass(); err != nil {
			return 0, err
		}
		took := time.Since(start)
		if best == 0 || took < best {
			best = took
		}
	}
	return best, nil
}

// recordOfAboutAKilobyte is a record the size of the ones in this repository. It
// is prose rather than a repeated character because a filesystem that
// compresses would otherwise be measured doing something no real tree asks of
// it.
func recordOfAboutAKilobyte() string {
	var out strings.Builder

	out.WriteString("Slug: experiment\nState: answered\nQuestion-Written: 2026-01-01\nAnswer-Written: 2026-01-02\n\n")
	out.WriteString("## Question\n\nDoes this shape of record cost anything to read?\n\n")
	out.WriteString("## Method\n\n")
	for i := 0; out.Len() < 1024; i++ {
		fmt.Fprintf(&out, "Line %d of a method somebody wrote out in full, at about the width the prose in this repository is written at.\n", i)
	}
	out.WriteString("\n## Answer\n\nNo, and this file exists to find out by how little.\n")
	return out.String()
}

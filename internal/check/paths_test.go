package check

import (
	"strings"
	"testing"
)

// TestWhatCountsAsALinkToAFileBesideTheDocument exercises the reader directly,
// because the invariants fixtures cannot reach most of it. A fixture is a tree
// and a verdict, so it proves that a link to a missing file is refused and a
// sentence naming the same file is not, and it says nothing about a target
// carrying a scheme, a fragment or a directory. Those arms all produce the same
// outcome as an arm that stopped working, which is silence, so this is what
// notices one going quiet.
func TestWhatCountsAsALinkToAFileBesideTheDocument(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "an ordinary link",
			text: "See [the notice](NOTICE.md) before you start.",
			want: []string{"NOTICE.md"},
		},
		{
			name: "a name with no extension",
			text: "The text is in [DCO](DCO) at the root.",
			want: []string{"DCO"},
		},
		{
			name: "a fragment on the end",
			text: "See [the exit codes](README.md#exit-codes).",
			want: []string{"README.md"},
		},
		{
			name: "the same target twice",
			text: "[one](NOTICE.md) and [again](NOTICE.md)",
			want: []string{"NOTICE.md"},
		},
		{
			name: "two targets keep the order they were written in",
			text: "[first](NOTICE.md) then [second](LICENSE)",
			want: []string{"NOTICE.md", "LICENSE"},
		},

		{
			name: "the same name in a sentence rather than in a link",
			text: "There is no LICENSE in this tree yet.",
		},
		{
			name: "a target under a directory is the other reader's",
			text: "See [the record](docs/decisions/0002-repository-layout.md).",
		},
		{
			name: "a target on somebody else's server",
			text: "See [the page](https://example.invalid/NOTICE.md).",
		},
		{
			name: "an address rather than a file",
			text: "Write to [the list](mailto:nobody@example.invalid).",
		},
		{
			name: "a place inside the same page",
			text: "See [further down](#what-was-run).",
		},
		{
			name: "a windows separator is a directory too",
			text: `See [the record](docs\notes.md).`,
		},
		{
			name: "an empty target",
			text: "See [nothing]().",
		},
		{
			name: "a parenthetical in a sentence after a bracket",
			text: "The answer was no [and it stayed no] (which was the useful part).",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LinkTargetsWithoutADirectory(tc.text)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("read %v, want %v", got, tc.want)
			}
		})
	}
}

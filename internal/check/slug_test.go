package check

import (
	"strings"
	"testing"
)

// TestWhatCountsAsALegalSlug exercises the shape directly, because the fixtures
// cannot. Every arm below produces the same property, and the harness compares
// which properties a tree refused rather than which line refused them, so a case
// for one arm satisfies the standing requirement for all of them. This is what
// notices an arm that stopped biting.
func TestWhatCountsAsALegalSlug(t *testing.T) {
	tests := []struct {
		name   string
		slug   string
		refuse bool
	}{
		{name: "an ordinary slug", slug: "sequential-write-on-spinning-disk"},
		{name: "one word", slug: "timing"},
		{name: "digits", slug: "raid-10-rebuild"},
		{name: "digits only", slug: "2026"},
		{name: "exactly the limit", slug: strings.Repeat("a", slugLimit)},

		{name: "empty", slug: "", refuse: true},
		{name: "one character over the limit", slug: strings.Repeat("a", slugLimit+1), refuse: true},
		{name: "upper case", slug: "Timing", refuse: true},
		{name: "a space", slug: "timing test", refuse: true},
		{name: "brackets and a comma", slug: "timing-test-(final,-v2)", refuse: true},
		{name: "two hyphens in a row", slug: "timing--test", refuse: true},
		{name: "a leading hyphen", slug: "-timing", refuse: true},
		{name: "a trailing hyphen", slug: "timing-", refuse: true},
		{name: "an underscore", slug: "timing_test", refuse: true},
		{name: "a path separator", slug: "timing/test", refuse: true},
		{name: "a dot", slug: "timing.test", refuse: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reason := refuseSlug(tc.slug)
			if tc.refuse && reason == "" {
				t.Fatalf("%q is accepted as a slug", tc.slug)
			}
			if !tc.refuse && reason != "" {
				t.Fatalf("%q is refused as a slug: %s", tc.slug, reason)
			}
		})
	}
}

// TestALongSlugIsRefusedForItsLengthAndNotForItsShape holds the one arm whose
// message a reader is most likely to be sent the wrong way by. A slug that is
// otherwise perfectly formed and too long needs to hear about the length, not
// about letters and hyphens.
func TestALongSlugIsRefusedForItsLengthAndNotForItsShape(t *testing.T) {
	reason := refuseSlug(strings.Repeat("a", slugLimit+1))
	if !strings.Contains(reason, "characters") {
		t.Fatalf("a slug one character too long is refused with %q, which does not name the length", reason)
	}
}

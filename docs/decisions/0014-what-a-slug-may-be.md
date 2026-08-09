# 0014. What an experiment slug may be

## What was decided

This record supersedes 0008. Everything that record fixes about an experiment
record stands exactly as it is written there, and this adds the one thing it
left open: what a slug may be. Record 0008 says so itself, at the field, and
names the issue that would come back for it.

A slug is lower case letters, digits and single hyphens. It begins and ends with
a letter or a digit, it never carries two hyphens in a row, and it is at most 64
characters long.

Written as the check reads it:

    ^[a-z0-9]+(-[a-z0-9]+)*$

That governs both places a slug appears. The directory under `experiments/` is
named with one, and the `Slug` field in the record repeats it.

Two costs decided this, and the second is the one that bites.

A reader loses the walk. A slug is quoted in a listing, in a promotion section
read on another board, and in a sentence somebody writes about a result. A
string with spaces and brackets survives none of those cleanly, and the record
pointing at it stops resolving for a reason that is invisible in the text.

A checkout stops agreeing with itself across machines. The suite runs on every
platform the release ships a binary for, because the runner's whole job is
reading a checkout and that is the part that differs. Two experiments whose
slugs differ only in case are two directories on one filesystem and one
directory on another. The second machine either loses a record or reports a
mismatch that is the correct answer there and the wrong answer everywhere else,
and once both directories exist in history no check can repair it. Lower case
only is what removes that case, and the machine that would have folded them is
the machine nobody would have heard from.

64 characters is a number rather than a derivation. It is longer than any
question anybody has needed to name and short enough that a path built from it
survives every filesystem the release targets. It is written here so that a
reader meets a number somebody chose rather than a limit that appeared.

What this does not decide is whether a slug describes its experiment. `test-2`
is a legal slug and a poor name, and no reading of the tree separates those. The
record's question is what a reader goes by, and the review is where a poor name
is caught.

## What it applies to

Every directory under `experiments/` and every `Slug` field, from the commit
this record lands on.

It applies to the checks that read either of them, which is issue #59, and the
shape above is written at those checks as well as here so that a check does not
read its rule out of a document.

It does not apply to anything else the tree names. A decision record's filename
is record 0000's, a package directory is Go's, and a branch name is nobody's
here.

## What else was considered

Allowing upper case, and comparing slugs case-insensitively wherever they are
read.

Allowing underscores as well as hyphens.

Fixing no shape at all and leaving the directory name to whoever creates it,
which is the state this record ends.

Deriving the slug from the question rather than letting anybody write one.

## What each rejected option would have cost

Upper case with a case-insensitive comparison costs a rule that has to be right
in every place a slug is compared, and there are more of those every milestone:
the walk, the listing, the header check, the promotion section, and whatever
reads a slug next. One of them will compare with the wrong one of the two
methods, and the machine where it matters is the machine the author is not
sitting at. Forbidding the case that folds is one rule in one place.

Underscores buy nothing a hyphen does not, and cost a second separator that has
to be chosen between every time a slug is written. Two spellings of the same
name is exactly the ambiguity a fixed shape exists to remove.

No shape at all costs what this record was written for. `Timing test (final,
v2)` satisfies every other rule in this plan, quotes badly everywhere a slug is
quoted, and cannot be repaired later without renaming a directory the
permanence rule keeps.

Deriving the slug from the question costs the author the one naming choice they
should have. A derived slug is long, it changes if the question is reworded, and
the question is allowed to be reworded before the work starts. It also puts a
transformation between what somebody typed and what the tree holds, which is one
more thing to get right on a filesystem that folds case.

# 0000. How decisions are recorded

## What was decided

Decisions that shape this repository live in `docs/decisions/`, one decision per
file. A file is named `NNNN-short-slug.md`, where `NNNN` is a four-digit number
that is never reused and never renumbered. This record is `0000` and it is the
one every later record is written against.

Each record carries four sections, in this order and under these headings:

1. `## What was decided`
2. `## What it applies to`
3. `## What else was considered`
4. `## What each rejected option would have cost`

Section 3 lists the options that were not taken. Section 4 says, for each of
them, what taking it would have cost. A record whose section 3 is empty is not a
decision, it is an announcement, and the format is built so that a reader sees
the difference without having to know the history. Where an option was rejected
for a reason that is not a cost, section 4 still names the option and says so.

A record is not edited once something depends on it. Fixing a typo before
anything references the record is ordinary. Changing what it decided, after work
has been done against it, is not, because a reader who followed the old text has
no way to find out that the ground moved.

Decisions change by superseding. A later record names the earlier one by number
in its first section, says what moved and why, and the earlier record stays where
it is with its text intact. The earlier record gains one line at the top naming
the record that superseded it, and that line is the only edit a superseded record
takes. So the reasoning that was true at the time is still readable, next to the
reasoning that replaced it.

Numbers come from the plan rather than from the order the files happen to arrive
in. Each decision has an issue, and the issue names the number and the filename
its record will carry, so two records written at the same time do not choose the
same number.

## What it applies to

Every decision that shapes this repository and that later work depends on: the
layout, the language the runner is built in, what an experiment record is, what
may never be committed, what happens to code once an experiment has an answer,
and anything else a person arriving later would otherwise have to reconstruct
from a diff.

It does not apply to the ordinary judgements inside one change. Which name a
function gets, or how a check is written once its rule is decided, is argued in
the change and read in review. A record for every such choice would bury the
records that matter.

It does not apply to the tracker either. Issues are where a decision is argued
and where the evidence is collected. The record is what survives the argument, so
a decision that exists only in an issue is not yet recorded.

## What else was considered

Keeping decisions in issue comments and linking to them from the code.

Keeping one long document listing every decision in sequence.

Keeping no separate record at all, and treating the commit message of the change
that first depends on a decision as the record.

Requiring a status field on each record, with values such as proposed, accepted
and superseded.

## What each rejected option would have cost

Issue comments have no current state. An issue is a conversation, so a reader
arriving later has to read the whole thread and work out which sentence won,
and a decision that was revised three times reads as three decisions. The
tracker can also be closed, exported or moved, and then nothing that depended on
those comments has a source any more.

One long document produces a merge conflict on every decision, because every
change touches the same file, and it grows into something nobody reads to the
end. It also makes superseding invisible, since the natural way to change such a
document is to edit the paragraph that is wrong, which destroys the reasoning
that was true before.

Commit messages are attached to a change rather than to a subject. A decision
that shapes five later changes lives in the message of whichever one happened
first, which is not where anybody would look, and `git log` is a poor index for
somebody who does not already know what they are searching for. History is also
rewritten more often than a tracked file is.

A status field costs an accuracy problem that has no owner. A record marked
accepted stays marked accepted after it has been superseded unless somebody
remembers to go back and change it, and an edit to a landed record is exactly
what this decision refuses. Superseding by writing a later record needs no field
to be kept true, because the later record's own existence is the status.

# 0026. Evidence in a landed record that has stopped reproducing

## What was decided

A landed record's evidence is never brought up to date. A paste in a record is
a measurement of the day it was taken rather than a claim about today, and
editing it into agreement with a later reading forges a run nobody made. Record
`0000` gives a landed record exactly one legal edit, the line naming what
superseded it, and that edit stays exactly as narrow as it is.

When a paste in a landed record stops reproducing, one of two things is true,
and which one it is decides everything that follows.

Either the decision still stands on today's evidence, and then nothing happens.
A stale paste under a standing decision is history, not error. It says what was
in front of whoever took the decision, which is what a later reader needs in
order to judge the reasoning rather than only the outcome. No edit, no
annotation, no successor record.

Or the ground moved far enough that the decision would fall differently today,
and the answer there is not an edit either. It is a new record superseding the
old one, carrying the fresh measurement and naming what it replaces, with the
earlier record left word-stable as the reason things were once otherwise. That
route is already record `0000`'s, and what this record adds is that a fallen
premise is one of the things that sends a decision down it.

So the question a paste that no longer reproduces raises is about the decision
and not about the paste. Does this still hold is the whole test. Does this
still print what it printed is not the test, and answering the second one is
how five documents in this tree came to describe an answered question as open.

**What a reader is owed here and does not get.** Nothing marks a paste that has
stopped reproducing, and under this decision nothing ever will, so a reader who
needs to know whether a quoted command still answers that way runs it. Carrying
the command is what makes that possible and is why these records carry commands
at all. The cost is real and lands on the reader in a hurry, who takes a paste
for a current reading: the two records this route was decided against are both
cases where a reader following the paste is told something that was true on the
day it was taken and is not true now. That price is named here rather than met
later, and it is what a directory whose records can be trusted not to have
moved under anybody costs.

**Nothing refuses either half of this.** No check reads whether a paste still
reproduces, and none could without running commands out of a document, which is
the opposite of the property record `0009` fixes about what an automatic run
executes. Nothing notices a decision whose premise has fallen and whose
successor nobody wrote either. Both halves rest on whoever reads a record and
whoever reads the change, and that is written here so that this record is not
mistaken for a gate.

## What it applies to

Every record in `docs/decisions/`, the ones already landed included, because
what it governs is what may be done to a record after it lands rather than how
one is written.

It applies to a paste, a quoted command output, a linked issue state and any
other reading a record carries as evidence. It does not apply to what a record
decided, which is record `0000`'s subject, and it opens no second route to
changing that.

It does not apply to a typo fixed before anything depends on the record, which
record `0000` already calls ordinary.

It does not apply to documents outside `docs/decisions/`. A document describing
how this repository works is meant to describe it today, so a claim in one that
has stopped reproducing is a defect to repair in place. That is the difference
this record rests on: `docs/promotion.md` was repaired and record `0012` is not
going to be.

Two records were read against the test above on 2026-08-29, and both are
standing decisions on today's evidence, so neither is edited and neither gains
a successor.

`docs/decisions/0012-the-supported-platforms.md` says that whether this board
publishes downloadable release artefacts is open, and makes itself superseded
if the answer is source only. The answer is not source only, so the conditional
never fires and the six platforms stand:

    gh issue view 46 --repo Flowfin/lab --json state --jq '.state'
    CLOSED

`docs/decisions/0018-the-licence-of-this-board.md` says the file on the default
branch says something else, that this repository declares nothing, and that a
reader should trust neither the record nor the sidebar until issue #47 lands.
That issue landed, and the three readings now agree with what the record
decided:

    gh api repos/Flowfin/lab --jq '.license.spdx_id'
    GPL-3.0
    gh issue view 47 --repo Flowfin/lab --json state --jq '.state'
    CLOSED
    git grep -n 'const DeclaredLicence' origin/main \
      -- internal/invariants/invariants.go
    origin/main:internal/invariants/invariants.go:92:const DeclaredLicence = "GNU GENERAL PUBLIC LICENSE"

Those pastes are measurements of 2026-08-29 and this record is subject to its
own rule about them.

## What else was considered

Editing the stale paste in place and leaving the decision alone.

Widening record `0000`'s single legal edit to admit an evidence-only amendment
appended to a landed record.

Superseding whenever a paste stops reproducing, whether or not the decision
moved.

Marking a stale paste with a line saying it no longer reproduces.

Deciding nothing, and leaving each case to whoever meets it.

## What each rejected option would have cost

Editing the paste in place costs the record the only thing that makes it worth
quoting. A paste is presented as the output of a command run at a moment, so an
edited one asserts a run that never happened, in the file a later reader trusts
most. It also destroys the evidence that the ground moved at all: after the
edit, a record whose premise fell and one whose premise held read identically.
It is the edit record `0000` refuses, arriving under a name that sounds like
maintenance.

An evidence-only amendment costs the one-legal-edit rule its edge, and it costs
it at once rather than gradually. An amendment to the evidence and an amendment
to the argument are the same shape in a diff, so the rule stops being readable
off the file and becomes a judgement about what the editor meant. It also
creates an obligation with no owner: an amendment is itself a reading of a day,
so a record accumulating them accumulates stale ones, and the option solves its
own problem once and then reproduces it.

Superseding on every stale paste costs the directory its signal. A superseding
record means the ground moved, and a chain of them over records whose decisions
never changed leaves a reader unable to tell which supersession changed
anything. It also grows without limit and for reasons outside this board: a
closed issue, an edited ruleset or a renamed constant elsewhere would each owe
a record here, and the number of them would say nothing about how often this
repository changed its mind.

Marking the paste is the status field record `0000` already rejected, in a
different costume. It is a line that has to be kept true by somebody
remembering, on a file the rules say may not be edited, so it needs the
amendment option above before it can exist at all and then inherits every cost
of it. A marker nobody maintains is worse than no marker, because a record
carrying one unmarked stale paste and one marked one reads as though the
unmarked one is current.

Deciding nothing costs what was already measured. Five documents on this board
described an answered question as open, one of them long enough that a
checklist refused a hand-over the answer permits, and the whole set was found
by somebody running the commands rather than by anything noticing. Without a
written route each of those meets a different reader who resolves it a
different way, and the ones who resolve it by editing leave nothing behind
saying they did.

# 0022. How long a held-back record waits

## What was decided

A held-back record waits 90 days from the report. The window is published in the
security policy. There is exactly one written extension, granted on a reasoned
request.

**This record names record `0010` and says what moved.** That record decides
that an experiment whose question is whether shipped software can be made to
misbehave does not start here in public, and that the record is written once the
flaw is fixed and the affected project has said what it wants said. It says in
its own words that how long such a record waits when neither of those arrives is
not decided there, and that what the listing shows meanwhile is part of the same
question and is also not decided there. Both halves are decided here. Nothing
else in `0010` moves: which experiments do not start in public, where the
question goes instead, and what has to be true before the record is written are
unchanged, and this record does not supersede it.

**Why a window rather than waiting.** Waiting indefinitely is safest for the
people running the affected software, and it hands the schedule to whoever is
slowest to reply. What it costs is a board carrying an unwritten record that
nobody outside knows exists, which is the invisible half-finished state this
board was opened to make impossible. A date is what turns a silence into
something both sides can plan against.

**Why one written extension rather than a bare window.** A date chosen in
advance is sometimes wrong for the flaw in front of it, and publishing on
schedule against a flaw still live in software this organisation ships is a
thing to choose deliberately rather than to arrive at by default. The extension
is written and reasoned so that the choice is visible afterwards, and it is one
rather than any number, because a window that can be extended repeatedly is
waiting indefinitely with extra paperwork.

**What the extension costs, stated rather than presented as free.** It needs
somebody here to answer the request inside the window. That is the same failure
the window exists to handle, arriving one level up: a request that goes
unanswered leaves the record at its original date, which is the better of the
two outcomes and is not a decision anybody took in that moment.

**What the listing shows while a record waits.** A field beside the state rather
than a fourth state. The record stays in `asking`, and it carries `Held-back:`
with the date of the report, which is what the 90 days are counted from. The
listing prints that as its own dated line, saying that an experiment is held
back and when the clock started, and saying nothing about what it is about. A
date and the fact of a hold disclose that something exists. They disclose
nothing a reader could act on, which is the whole point of holding it back.

**What that costs record `0003` and the state refusal, which is where this
answer is not free.** Record `0003` fixes three states and no others, and
`record-state-is-not-one-of-the-three` refuses a fourth:

    git grep -n -o 'record-state-is-not-one-of-the-three' origin/main \
      -- internal/check/check.go
    origin/main:internal/check/check.go:176:record-state-is-not-one-of-the-three

Both stay exactly as they are, and that is the cost rather than the saving. The
refusal guarantees that the state a record carries is one of three. It has never
guaranteed what the listing prints, and after this record the listing prints a
line that no refusal governs. A reader who took a green run for a guarantee that
the listing shows one of three things was already wrong, and this is the change
that makes them visibly wrong.

The second half of the cost is the field itself. It is a field added to the
record format after record `0013` landed, so it is optional and a check over it
refuses only what is present. A record that is being held back and does not
carry the field is not refused: it sits in `asking` with a question that says
nothing, which is exactly the misreport this half of the answer exists to
prevent. **Nothing in this repository refuses that today, and nothing can under
record `0013`.** What stands behind it is the template, the review, and whoever
writes the record.

**The 90 days has to reach `SECURITY.md`, and it is not there today.** That file
is where the window is published, and this record is not that publication. It
also carries a sentence that a reader will meet first and could take for a
contradiction:

    git show origin/main:SECURITY.md | grep -n 'no response deadline'
    139:There is no response deadline and there will not be one. Nothing holds me to a

The two run in opposite directions and both stand. That sentence is about a
report arriving here and how long a reporter waits for an answer from this
board. This record is about a report leaving here and how long this board waits
before publishing what it found in somebody else's software. A window on what
this board owes others is not a promise about what others may expect from it.

## What it applies to

Every experiment record held back under record `0010`, from the commit this
record lands on: the 90 days, the single extension, and the `Held-back:` line
that makes the hold visible in the listing.

It applies to `SECURITY.md`, which is where the window is published and where a
reporter and an affected project read it.

It applies to the record format, which gains the field under record `0013`, and
to whatever check reads that field, in the direction of what is present only.

It does not apply to record `0003`. The three states stay three, and a held-back
experiment is in `asking` like any other experiment whose work has not finished.

It does not apply to the questions record `0010` answers. Which experiments do
not start here in public, where the question goes instead, and what has to be
true before a record is written are that record's and are unchanged by this one.

It does not apply where an affected project publishes its own disclosure policy
and this board has reported into it. The two windows are separate promises, and
the earlier date is the one that binds this board.

## What else was considered

Waiting indefinitely, so the record is written when the fix and the statement
arrive and never before.

A fixed window with no extension, counted from the report and published in the
security policy.

## What each rejected option would have cost

Waiting indefinitely costs this board the thing it was opened for. An experiment
that found something real would sit in a state nobody outside can see, for as
long as whoever is slowest to reply takes, and the schedule would belong to them
rather than to anybody who took a decision. It is also the option that decays
quietly: nothing marks the day it became unreasonable, so the record is not
published late, it is simply never published, and the board reads as though the
experiment never happened. That is the invisible half-finished state every other
rule here exists to prevent, arriving through the one door left open for a good
reason.

A fixed window with no extension costs the case the window is worst at. A date
chosen 90 days in advance is a guess about a flaw nobody had looked at yet, and
the case where it is wrong is the case where publishing on schedule puts a live
flaw in software this organisation ships in front of everybody who might use it.
Under that option the only ways out are publishing anyway or breaking the rule,
and a rule broken once for a good reason is a rule everybody afterwards knows
can be broken. The extension costs an answer inside the window, which is a real
cost and is written above. What it buys is that the exception is taken in
writing rather than taken silently.

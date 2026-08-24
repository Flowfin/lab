# 0019. Code under another licence

## What was decided

An experiment may bring in code under a licence that is not this board's, and
only inside a quarantined directory that carries its own licence file.

The directory is `experiments/<slug>/borrowed/`, and the licence the borrowed
code arrives under is `experiments/<slug>/borrowed/LICENSE`. One such directory
per experiment. It sits inside the experiment rather than at the root of the
tree, which is not a stylistic choice: record `0002` refuses a root directory it
does not name, so a quarantine at the root would be a change to that record,
while the layout inside an experiment is loose by that record's own words and
`experiments/<slug>/` already holds one experiment and everything it needs.

Everything under that directory is under the licence its own file names.
Everything outside it is under this board's own licence, which is decided as
entry one of the plan and recorded under its own issue rather than here.

The experiment record declares it. `Borrowed:` in the header names the source and
the licence, and its absence means the experiment borrowed nothing. That is a
field the record format does not carry today, so adding it is a change to the
format and goes through
[0013-how-the-record-format-changes.md](0013-how-the-record-format-changes.md)
rather than being written into a header and discovered by a reader.

Why this rather than the two ends. Allowing borrowed code anywhere with nothing
but a line in the record puts a directory that is not under this board's licence
somewhere in a tree that reads as uniform, and somebody promoting the work later
has to notice it at the exact moment they are least careful. That is a footgun
placed where attention is lowest. Forbidding it outright makes a real class of
question unaskable here, and the questions that begin with existing code are a
class this board wants, because most interesting questions about software are
questions about software that already exists.

What the quarantine buys is that the boundary is in the layout rather than in
somebody's attention. A person promoting a result walks a directory tree, and a
directory named `borrowed` with a licence file in it is visible to a walk. A
declaration in a header is visible only to a reader who opened the header.

What it costs, stated rather than left to be discovered. It is a layout rule that
has to be taught, so it belongs in the contributing guide and not only here. It
needs a check, which is ordinary gate work in this tree rather than a new
apparatus: the runner already walks `experiments/` and already reads the header.

**Nothing in this repository refuses a violation of this rule today.** No check
reads `borrowed/`, no check reads a `Borrowed:` field, and the field is not part
of the format yet. What stands behind the rule is this record, the guide, and
whoever reads the change. That is the state on the day this lands, and it is
written here so the rule is not mistaken for a gate.

## What it applies to

Every experiment on this board that starts from code somebody else wrote, from
the commit this record lands on.

It applies to the contributing guide, which has to carry the rule where a person
starting an experiment meets it, and to the record format, which has to gain the
field before the declaration means anything.

It does not apply to the runner. `cmd/lab/` and `internal/` borrow nothing under
this rule; a dependency of the runner is a module requirement and is covered by
the notices and the bill of materials the release carries.

It does not decide whether a particular licence may be borrowed from at all.
Some licences are incompatible with promotion into a GPL plugin board whatever
directory the code sits in, and this record fixes where borrowed code lives
rather than which licences are acceptable.

## What else was considered

Allowing borrowed code anywhere in an experiment, with a declaration in the
experiment record naming the source and its licence.

Forbidding borrowed code entirely, so every experiment starts from nothing.

Putting the quarantine at the root of the tree, in one directory shared by every
experiment.

## What each rejected option would have cost

Allowing it anywhere costs the promotion path its safety. The tree then reads as
uniform and is not, and the only thing separating a file under this board's
licence from a file under somebody else's is a line in a header that the person
copying the directory did not open. The failure lands on whoever promotes the
result, which is later, elsewhere, and after the person who knew has stopped
thinking about it.

Forbidding it entirely costs a class of question this board exists to take. An
experiment asking whether an existing implementation behaves a particular way
cannot be run without that implementation, and the answer would be to run it
somewhere unrecorded, which is worse than running it here: the board loses the
record and gains nothing.

A shared quarantine at the root costs a change to record `0002`, which refuses a
root directory it does not name, and it costs the property that makes the
quarantine work. Borrowed code would sit away from the experiment that borrowed
it, so deleting an abandoned experiment leaves its borrowed code behind, and two
experiments borrowing different versions of one thing collide in a directory
neither of them owns.

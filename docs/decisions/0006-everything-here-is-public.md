# 0006. Everything here is public, and what that means may never be committed

## What was decided

This decision has two halves and they belong together.

### The first half, what being public buys

Everything in this repository is public from the first commit, including the
experiments that failed. That is deliberate, and it is worth saying out loud
because it changes what people are willing to try.

A failure that is visible is a failure somebody else does not have to repeat. An
experiment that answered no is, on this board, a finished piece of work and it is
presented as one. Its record sits next to the records of the experiments that
answered yes, in the same format, with the same weight.

Nothing here is hidden because it did not work. Nothing is quietly deleted
because it is embarrassing. Where an experiment's code is removed, that happens
under record `0004` and leaves a line saying so, which is the opposite of quiet.

### The second half, what may never be committed

Precisely because everything here is public, some things may never enter the
tree. Git history does not forget, so this list is about what is committed at
all, not about what is in the current checkout.

- No credential, token, key or secret of any kind, including ones that have
  expired or been rotated. An expired credential still reveals its format, its
  issuer and often the account it belonged to.
- No personal data of any kind, whether it belongs to the person running the
  experiment or to anybody else. Names, addresses, e-mail addresses, device
  identifiers and account identifiers are all covered.
- No copy of a real media library, real account data, or real logs from a running
  server, whoever runs it.
- No file whose licence forbids it being here.

Where an experiment genuinely needs real data to answer its question, the data
stays on the host it is already on. It does not enter the tree, not as a fixture,
not as a sample, not as an attachment to a record. The record carries what was
measured and the command that measured it, and it carries no copy of the thing it
was measured against.

Whether an experiment on this board may work with real data at all, even on the
host and never in the tree, is a wider question than this record settles. This
record fixes only the boundary at the tree, and that boundary holds whichever way
the wider question is answered.

The reason the two halves are one decision is that a public board with no stated
exclusions produces one bad commit and then a scramble. The exclusions are the
price of the openness, and stating them apart from it invites somebody to read
the first half without the second.

Nothing in this repository can refuse a violation of the second half today. There
is no scanner in the tree that reads a commit for a credential or for personal
data, and this record does not create one. What stands behind the list is a
person reading before they commit and a person reading the change, which is
stated here rather than implied so that nobody mistakes the list for a gate.

## What it applies to

Every commit on every branch of this repository, and every file a change adds,
including test fixtures, sample configuration, screenshots and anything attached
to a record.

It applies to the second half retroactively in the only sense that matters. A
thing that reaches history is not removed by a later commit, so a mistake here is
handled as an exposure and not as a revert, which means rotating what leaked and
saying so.

It does not apply to the tracker, which follows the same spirit but has its own
route for taking something down.

## What else was considered

Making the board private, or private until an experiment succeeds.

Publishing only the experiments that answered yes.

Stating the openness and leaving the exclusions to a separate document written
later.

Listing the exclusions as guidance rather than as a rule, on the reasoning that
nothing here refuses a violation anyway.

## What each rejected option would have cost

A private board costs the whole reason the board exists. A failure nobody can
read saves nobody any work, and an experiment that becomes public only once it
succeeds teaches the same lesson as every other board in this organisation, which
is that only finished things are shown. It would also mean the interesting half
of the corpus, the negative results, never reaches anybody.

Publishing only the successes costs the record its meaning. The listing would
show a board where everything worked, which is not a description of any research
that has ever happened, and a reader would have no way to tell an experiment that
answered no from one that was never run.

Leaving the exclusions for later costs exactly one commit. The gap between
opening a public board and writing down what may not go on it is the window in
which the mistake happens, and the mistake is permanent. The two halves are one
record so that the window does not exist.

Guidance rather than a rule costs the ability to say a thing was violated. If the
list is advice, then a commit carrying a credential broke nothing, and there is
no clear statement to point at when deciding how to respond. The absence of a
machine that refuses something is a reason to write the rule more plainly, not a
reason to write it more weakly.

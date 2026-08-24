# 0023. A commit on the default branch carries a verified signature

## What was decided

A commit on the default branch of this board has to carry a verified signature.
It becomes effective as the account keys operations#1609 sets up for the working
accounts land, and not before, because a rule that refuses every merge before
anybody holds a key is a rule that gets turned off rather than followed.

The reason is that two claims are otherwise collapsed into one. That an account
pushed a commit and that a key signed it are different statements, and only the
second survives somebody else holding the account. Not requiring a signature
costs nothing to operate and rests authorship entirely on the account, which is
one credential and one recovery route away from being somebody else's.

The costs are real, they arrive later than the decision does, and each of them
lands on somebody who did not take it. They are written here rather than
discovered.

A key has to be held somewhere and rotated eventually. That is the same key
story record 0021 needs for signed release artefacts, which is why the two
answers were taken together: one custody story, written once, in
operations#1609.

An unsigned commit anywhere in a branch's history refuses the merge rather than
refusing the commit. So the repair is rebuilding the branch rather than adding
one more commit to it, and it lands at the end of the work rather than at the
start, which is the moment it is most expensive and least expected.

Edits made through the web interface and anything an automation authors are
signed by the platform's key or not at all. That decides what those routes can
still be used for, and it is a consequence rather than a side effect.

**Earlier history does not become signed.** This covers what arrives after the
day it takes effect, whatever was intended by setting it, so a reader must not
take a signed default branch for a branch that is signed end to end. Stating
that here is the point: the rule buys a property from a date, and a property
from a date read as a property of the whole is worse than no property at all.

The requirement is a setting on the branch protection rather than anything in
this tree, so nothing here refuses an unsigned commit and this record does not
claim otherwise. Read the live state rather than trusting this paragraph:

    gh api repos/Flowfin/lab/rules/branches/main --jq '.[].type'

Neither this board nor the board the quality-parity work targets required a
signature when this was decided, so parity settles nothing here and the ruleset
walk points at this record instead of deciding it.

## What it applies to

Every commit reaching the default branch of this board from the day the keys
land, whoever authors it, with no bypass for whoever administers the repository.
A gate with an exception for the person most likely to be in a hurry is not a
gate.

It applies to the ruleset work that configures it and to the quality-parity
document that walks the ruleset.

It does not apply to a feature branch, which no ruleset covers, and it does not
apply to anything already in history.

It decides that a signature is required and does not decide what happens to a
contributor who has no key. This board takes experiments from anybody, so that
question is real rather than hypothetical, and the record answering who may run
an experiment here is where it belongs.

## What else was considered

Not requiring a signature at all, and resting authorship on the account.

Requiring one from the first release onward rather than from the day the keys
land.

## What each rejected option would have cost

Resting on the account costs the distinction the whole decision is about. It is
free to operate and it is exactly as strong as the weakest recovery route on one
account, which is not a property this board controls or can measure. The cost is
invisible until the day it is total.

Tying it to the first release costs the same thing the effective date already
costs, and more of it. A history carrying unsigned commits does not become
signed afterwards, so the later the rule starts the more of the branch it does
not cover, and the release is later than the keys. It also splits the key story
from record 0021 by a stretch of calendar for no gain, since the keys that sign
the release artefacts are the keys that sign the commits, and having them
without using them is a cost already paid and not collected.

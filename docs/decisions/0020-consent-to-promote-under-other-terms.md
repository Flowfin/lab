# 0020. Consent to promote a result under another board's terms

## What was decided

A result may be promoted into a board under different terms from this one only
with the contributor's explicit consent, given at promotion time.

No contributor agreement in advance, and no restriction to work somebody
authored themselves. Both of those are on the list below with what they would
have cost.

**This record names record `0005` and says what moved.** That record fixes
promotion as a hand-over recorded on both sides, and it says the promotion
section names four things and that nothing about it is optional. It also says
that who is entitled to place a contributor's work under another board's terms
is not decided there. This is that question answered, and the answer reaches the
section: it names a fifth thing.

It does not supersede record `0005`. Everything else that record fixes stands
unchanged, and marking it superseded would tell a reader that the whole of it
had been replaced when one paragraph of it has been added to. What that costs is
real and is not softened here: read on its own, `0005`'s sentence that the
section names four things is now wrong, and nothing in `0005` says so. What
carries the correction is this record and a reader who finds it.

**The consent is a line in the record rather than a conversation somebody
remembers.** `Consent:` in the promotion section, naming who consented and the
terms consented to. A hand-over that happened in a conversation is a
hand-over nobody can check afterwards, and the question it gets asked in is the
one nobody wants to be answering: whether a contributor agreed to their work
going somewhere, months after the person who asked them has stopped thinking
about it.

**What reads it, stated as it stands rather than as it is meant to end up.**
`promotion-section-is-missing-something` is the property that reads the
promotion section and refuses one that does not name what record `0005` fixes:

    git grep -n -o 'promotion-section-is-missing-something' origin/main \
      -- internal/check/promotion.go
    origin/main:internal/check/promotion.go:16:promotion-section-is-missing-something

It names four things today and the consent line is not one of them, so **nothing
in this repository refuses a promotion section that names no consent on the day
this record lands.** What stands behind the rule until that changes is this
record and whoever reads the change. Adding the line to the set that property
reads is what turns the rule into a refusal, and it is a change to the record
format, which goes through
[0013-how-the-record-format-changes.md](0013-how-the-record-format-changes.md).

That route decides the shape of the refusal, and it is worth reading before
somebody writes one. Record `0013` makes an added field optional and
refuses only what is present, so a refusal on an absent `Consent:` line is a
refusal on an absence, which that record forbids. What is available instead is
the shape the four fields already have: the promotion section is the optional
thing, an experiment that was never handed over carries none and is not refused,
and a section that is present is held to everything it has to name. No record on
this board carries one today, so nothing turns red when the fifth name is added:

    git grep -c '^## Promotion' origin/main -- 'experiments/*/EXPERIMENT.md'
    exit=1

That count is what makes the change free at the moment it lands, and it is a
fact about today rather than a property of the design. A promotion section that
lands before the refusal does and carries no consent line is what record `0013`
says must not be reddened afterwards, and the repair then is a reader's rather
than a check's.

**The sign-off gate does not cover this, and nobody should read it as though it
did.** Every non-merge commit here carries a `Signed-off-by` trailer matching
its author, checked by the DCO workflow. What a contributor certifies there is
their right to contribute the work under this project's licence. It says
nothing about a different licence on a different board, it was not written to,
and a hand-over that leaned on it would be resting a licensing question on a
certificate about another one.

**What the answer costs.** A step at promotion time is a step that will
sometimes be skipped under time pressure, and promotion time is exactly when
somebody is trying to get a result into the place that wants it. That is the
known cost, it is the reason the consent is a written line rather than an
understanding, and the line is not a control until something reads it. Both
halves belong together: a residual stated with what holds it is honest, and a
residual stated alone reads as an excuse.

## What it applies to

Every promotion out of this board into a repository whose terms differ from the
licence record `0018` names, from the commit this record lands on.

It applies to the promotion section record `0005` fixes, which gains the
`Consent:` line, and to the record format, which gains it under record `0013`.

It applies to whoever does the promoting rather than to the receiving board. A
board takes work by its own rules, and nothing here obliges it to check what
this record asks for.

It does not apply to a promotion into a board under the same terms. There is
nothing for a contributor to consent to when the code leaves under the licence
it was written under, and asking anyway turns a real question into a formality
that gets ticked.

It does not apply to somebody copying an idea out of a record with no code
moving, which record `0005` already says needs no section at all.

## What else was considered

A contributor agreement, signed in advance, covering promotion under other
terms for everything that contributor ever writes here.

Taking only work the promoter authored themselves, so consent is never needed
from anybody else.

## What each rejected option would have cost

A contributor agreement puts a legal document in front of somebody who wanted to
try an idea. That is a real deterrent, and it is heaviest on a board whose whole
purpose is lowering the cost of trying: the person it turns away is the one who
had a question and half an hour, which is the contributor this board is for. It
also buys certainty at the wrong moment. The agreement is signed before anybody
knows which board the work might go to or what terms it would go under, so what
it collects is a permission granted in the abstract, and the cost of getting it
is paid by every contributor including the ones whose work never leaves.

Taking only self-authored work costs the board the thing the hand-over path
exists for. The results worth promoting are not reliably the ones the promoter
wrote, and under this option an experiment somebody else ran well enough to be
wanted elsewhere is stuck here. It also has a failure mode worse than the
restriction: the cheapest way around it is for the promoter to rewrite the work
and hand over their own version, which produces a worse result, loses the
original author's contribution, and does it quietly.

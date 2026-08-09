# 0013. How the record format changes once records are permanent

## What was decided

A field added to the experiment record format after this record lands is
optional, and a check over it refuses only what is present. An absent field is
never a refusal.

That means a record written before a field existed stays legal on the day the
field arrives. It also means the format can never require anything again, which
is the price and is paid deliberately.

What happens to records already on the default branch when the format changes:
nothing. They are not edited, not migrated and not marked. A record already
there is already correct, and it stays correct, because the rule that arrived
after it does not reach back.

A permanent record is not edited to satisfy a new rule. This is the sentence
somebody will be looking for while a board is red, so it is written plainly
here rather than implied by the rule above. If a new check makes older records
red, the check is wrong, not the records, and the repair is in the check.

The collision this avoids is concrete and it arrives the first time a field is
required rather than optional. The field lands, every record written before
that day lacks it, the checker refuses all of them at once, and the default
branch is red with two repairs available: edit records that are supposed to be
permanent, or weaken the check that just landed. Both destroy something, and
the one taken under a red board is whichever is quicker. Deciding now costs
nothing; deciding then costs whichever of the two somebody reached for first.

Two things this does not give up.

A check over a field that is present can be as strict as it likes. Optional is
about absence and about nothing else. A record declaring a field and
contradicting itself, or declaring it empty, is refused exactly as hard as it
would have been under a required field, and both refusals already planned in
this repository are of that shape rather than of the absence shape.

A field can still become effectively expected without becoming required. A
template that carries it, a checklist that asks for it and a review that
notices it missing are all available, and none of them turns an old record red.
What is given up is the machine refusing an omission, which is real: the
checker cannot tell a record that omitted a field on purpose from one that
forgot it, and that difference is a reader's judgement from the day this lands.

## What it applies to

The experiment record format, from the commit this record lands on, and every
change that adds a field to it. Both issues in the plan that add a field answer
to this record, which is why it is numbered ahead of them.

It applies to the checks that read those fields, in the direction of absence
only. What a check does with a field it can see is that check's own argument.

It does not apply to the decision records in `docs/decisions/`. Their shape is
record `0000`, they are not experiment records, and they change by superseding.

It does not apply to a field being removed from the format, which is a
different question, has not arisen, and would be a later record if it did.

## What else was considered

A version number in the header, with the checker holding the rules for every
version it has ever known.

A required field that applies from a date, refusing records whose question date
falls after it.

Migrating existing records when the format changes, by editing them.

Requiring the field and letting the board be red until somebody chooses.

## What each rejected option would have cost

A version in the header costs a field in every record forever, most of them
carrying the same number, and it costs the checker a branch per version that it
can never remove. It also puts the machinery in the wrong place: the thing that
varies is which rules apply, and a version number makes every author of every
record carry a piece of that machinery in a file they wanted to write prose in.
It buys the ability to require fields, which is the thing being given up here,
and it charges every record ever written for it.

A required field applying from a date costs three things at once. The checker
carries a date, which is a constant that means nothing to anybody reading the
source. A reader has to know that date to understand why two records in one
directory are held to different rules, and nothing in either record says so.
And the date is read from the record being judged, so the field is required of
exactly the records that declare themselves recent, which is the wrong way
round for a rule meant to bind what comes next.

Migrating existing records destroys the thing this board's central claim rests
on. A record is what somebody wrote when they wrote it, and a migration is a
later hand editing it so that a check written afterwards passes. Even a
faithful migration removes the ability to tell an original record from a
corrected one, and the first unfaithful one is undetectable.

Requiring the field and letting the board go red is the option that looks like
rigour and is the one that produces the worst outcome, because it is not
actually a decision. It defers the choice to the moment when both repairs are
destructive and somebody is under pressure, which is the collision this record
exists to prevent rather than a way of handling it.

## What this owes elsewhere

The compatibility position, wherever it comes to live, points at this record
rather than restating what it says. A second copy of this rule drifts against
this one, and the copy is the one a reader will find first.

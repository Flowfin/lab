# Code of conduct

## Who this covers, and where

Everybody who takes part here: an issue, a pull request, a review, a commit
message, an experiment record, and a private message that started from one of
those. It covers me. I hold this repository and I am inside these rules rather
than above them, and if that sentence were not here the rest of this document
would be a set of rules for other people.

The door is open on purpose, which is
[docs/decisions/0024-who-may-run-an-experiment-here.md](docs/decisions/0024-who-may-run-an-experiment-here.md),
and this file is part of what that costs. A board anybody may propose an
experiment on owes anybody who shows up a written answer to what happens when
somebody behaves badly.

## What is expected

Argue with the work. This board already runs that way for evidence: a claim
carries the command that produced it, and a disagreement is settled by executing
something rather than by who said it. The same rule pointed at people means the
thing under discussion is a record, a measurement or a change, never the person
who wrote it.

Let an experiment fail. Everything here is public from the first commit, a
failed experiment included, and the whole arrangement depends on it being cheap
to write down an answer of no. [CONTRIBUTING.md](CONTRIBUTING.md) says that
nothing on this board treats a no as a lesser outcome, and that anybody who feels
a check pushing them towards a more flattering answer should report it as a
defect in the check. Treating somebody's failed experiment as a failure of theirs
is the behaviour that would make those sentences false, and it is the one this
board can least afford.

Keep measured and assumed apart in the sentence. Being wrong in public is the
ordinary cost of working this way, and a correction is a repair rather than a
verdict on whoever needed it. Somebody who writes down that they did not measure
something has done the harder thing.

Take a refusal with its reason. A change sent back here comes back with what was
wrong, and disagreeing with that reason is fair. Reopening it unchanged somewhere
else is not.

## What is not acceptable

- An attack on a person rather than on their work, including one dressed as a
  question about their competence.
- Demeaning somebody for who they are, or for a group they belong to.
- Harassment, in the open or in messages that started here and continued
  somewhere else.
- Publishing somebody's private details - a legal name, an address, an employer,
  anything they did not put here themselves - without their consent.
- Unwanted attention of a sexual kind, and sexual material anywhere in this
  repository or its tracker.
- Sustained disruption: a settled argument reopened without new evidence, a
  thread derailed on purpose, a review answered with volume instead of
  substance.
- Threatening any of the above, whether or not it is carried out.

That list holds what has to be named and it is a floor rather than a boundary.
Behaviour nobody wrote down is judged by the section above it and is not
permitted by its absence from this one.

One entry on that list is also a rule about the tree, and the overlap is worth
seeing rather than tripping over. Somebody else's personal data is something this
board may never hold, which is
[docs/decisions/0006-everything-here-is-public.md](docs/decisions/0006-everything-here-is-public.md),
and git history does not forget. So publishing it here is a conduct problem and
an exposure at once, and it is handled as both: what was published is removed and
said out loud rather than quietly reverted.

## Reporting

Two routes, in this order.

**A private report on this repository.** The form under Security is the only
private channel this repository has:

    https://github.com/Flowfin/lab/security/advisories/new

It is labelled for a vulnerability, because that is what it was built for. A
conduct report sent through it is not misfiled: it reaches the same person, it
stays private, and submitting it publishes nothing. Whether the form answers at
all is a setting on the repository rather than a promise in a file, so it is read
rather than asserted:

    gh api repos/Flowfin/lab/private-vulnerability-reporting
    {"enabled":true}

Run 2026-09-05. That reading needs administrative access to this repository, so
it is not one a reader of this board can make. What a reader can do is open the
address above and see whether the form is there.

[SECURITY.md](SECURITY.md) names the same form for a vulnerability report. The
two documents point at one door on purpose: a second private channel would be a
second thing to keep alive, and a spare channel is unused right up to the moment
it is needed.

**A message through GitHub to the account that holds this repository**, which is
https://github.com/iderex. Use this where the first route does not fit - where
the report is about the form, or about me, or where you would rather it did not
sit in a security queue.

Do not open an issue about a conduct problem. An issue here is public from the
moment it is submitted, and it names the person it is about to everybody before
anybody has read it. That is the same reason this board has a security policy at
all, one file over.

No mailbox is published, here or anywhere else on this board. An address in a
document outlives whoever was reading it and stays in the history after it is
changed, and an address nobody watches is worse than none, because the document
promises a reply. Both routes above are accounts on a service that already
authenticates whoever sends the message, and neither needs anything kept alive to
go on working.

## What happens after a report arrives

I read it and I answer it. If something is missing I ask for it, and if I decide
it is not a problem you get the reason rather than silence.

There is no deadline and there will not be one, for the reason
[SECURITY.md](SECURITY.md) gives about the other route: a window this board
cannot hold is worse than none, because a reporter left past it cannot tell a
busy week from a report that never arrived.

What I can do is limited, and stating it exactly is better than implying more:
ask somebody to stop, edit or delete content on this repository, close or lock a
thread, block an account from this repository, and report an account to GitHub.
None of that reaches anybody outside this repository and none of it is a sanction
anywhere else.

Your name does not appear in whatever follows unless you ask for it to. Where
something has to be said in public it is said about the behaviour and the
outcome.

## Where this document is weak

**One person holds it.** A report about my own behaviour comes to me, and there
is no second reader, no appeal and no independent body. That is the honest state
of a board with one holder rather than an arrangement I am recommending, and it
is why the route that does not go through me is written here rather than left to
be found:

    https://github.com/contact/report-abuse

**Nothing enforces this.** No check in this repository reads this file and none
could. Every leg of the run judges bytes in the tree, and how somebody spoke to
somebody else is not a byte in the tree. What stands behind this document is that
I do what it says, and the way to find out that I did not is to report it. A
document that looked enforced would be worse than this sentence.

**It is not the Contributor Covenant**, and that is deliberate. That document's
enforcement section describes a body of people, a ladder of consequences and a
review of appeals. None of those exists here, and publishing a ladder nobody
climbs would describe an apparatus this board does not have, which is the failure
the paragraph above is written against.

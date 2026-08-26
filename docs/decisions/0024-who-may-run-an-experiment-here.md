# 0024. Who may run an experiment here

## What was decided

This board takes experiments from anybody.

That is what a public board usually means, and it is the only way this one
becomes useful to anybody who is not already here. The whole apparatus was built
for it before it was decided: a sign-off gate that asks somebody to certify
their right to contribute, an issue template that invites a proposal, and a
contributing guide written for a person arriving with an idea. None of those
settled the question, and each of them reads as a promise until it is settled.

**The residual is the part of this answer that has to survive into the record.**
The cost is not review effort. Review effort is ordinary, it scales with what
arrives, and nobody was ever going to decide this question on it. The cost is
that three rules on this board depend on a person having read them before they
start, and somebody arriving for the first time is exactly the person least
likely to have read any of them.

**The three rules, named rather than gestured at.**

What may never be committed. The list is in
[0006-everything-here-is-public.md](0006-everything-here-is-public.md) and again
under `## What may never be committed` in `CONTRIBUTING.md`: no credential, no
personal data, no copy of a real library or real logs, no file whose licence
forbids it being here. Git history does not forget, so a mistake here is handled
as an exposure rather than as a revert.

What happens when an experiment finds a flaw in shipped software. That is
[0010-a-flaw-in-shipped-software.md](0010-a-flaw-in-shipped-software.md), and
what it asks of somebody is that they recognise the moment before they write
anything down in public. It is the rule most likely to be met by accident, by
somebody asking a different question who finds something.

What an experiment may do with real data. `docs/privacy.md` carries what is
settled and says plainly that whether an experiment here may touch real personal
data at all is not settled there. So the third rule a newcomer must have read is
one this board has not finished writing, which is a worse position than the
other two rather than a lesser one: a person cannot follow a rule that does not
exist, and the part of it that does exist is the part in the list above.

**Nothing here refuses a violation of any of the three, and this record does not
soften that.** The guide says so at the list itself:

    git show origin/main:CONTRIBUTING.md | grep -n 'refuses a violation'
    141:Nothing in this repository refuses a violation of that list today. What stands

The nearest thing to a mechanism is the credential-shape scan over tracked text,
and its own source says what it is worth before anybody quotes it as cover:

    git grep -n -o 'FLOOR AND NOT A GUARANTEE' origin/main \
      -- internal/invariants/vocabulary.go
    origin/main:internal/invariants/vocabulary.go:22:FLOOR AND NOT A GUARANTEE

It matches formats that have actually leaked. It does not read entropy, it does
not see anything base64-wrapped or split across lines, and a password and a name
have no shape at all. Personal data, a copy of a real library and a licence that
forbids a file being here are outside it entirely, and the first two are what
this rule is mostly about. For the other two rules there is nothing at all:
whether a question is about a flaw in shipped software, and what an experiment
did on a host that is not this tree, are not properties of a checkout, and no
reading of one decides them.

So the residual sits on whoever reads the change, and the first two rules are
the ones where a mistake is public and permanent. That is what taking
experiments from anybody costs, and it is written here rather than left in the
issue that is closed when this record lands.

**The answer belongs in the first paragraph of `CONTRIBUTING.md`, with the three
rules linked from it.** The first paragraph, because the alternative is somebody
finding out after doing the work, and linked from that paragraph, because the
person least likely to have read the three rules should meet them where they
arrive rather than a hundred lines further down. That guide is on the default
branch already and carries none of this today, so what this record fixes is
where the sentence goes and the edit that puts it there is owed:

    git show origin/main:CONTRIBUTING.md | sed -n '3,8p' | head -1
    This board is for questions that will probably fail. Everything here is public

## What it applies to

Who may open an issue, propose an experiment and send a change here, from the
commit this record lands on. The answer is anybody.

It applies to `CONTRIBUTING.md`, whose first paragraph carries the answer and
links the three rules.

It applies to the templates under `.github/ISSUE_TEMPLATE/`, which invite a
proposal and are written for the audience this answer says exists.

It applies to the sign-off gate, which asks a contributor to certify their right
to contribute under this project's licence, and which now has contributors to
ask.

It applies to record `0020`. Consent to promote a result under another board's
terms only becomes a question once somebody other than the promoter has written
something here, and this is the answer that makes it one.

It does not decide what this board does with what arrives. A proposal can be
declined, an experiment can be asked to state its question better, and a change
can be rejected. Taking experiments from anybody is about who may offer, and
nothing here obliges this board to take any particular thing.

It does not settle the third rule. What an experiment may do with real data is
open, `docs/privacy.md` says where the answer comes from, and this record does
not answer it by needing it.

## What else was considered

Taking no outside experiments, and saying so.

Taking proposals as issues rather than as changes, so a question is discussed
before anybody writes code against it.

## What each rejected option would have cost

Taking none costs the board its reach, which is most of what a public board is
for, and it buys a much smaller surface: the three rules above would then be
read only by people who already know them. The cost that decides it is what
happens to everything already built. The contributing guide, the issue templates
and the sign-off gate are all written for a person arriving from outside, and
under this option every one of them is furniture for an audience that does not
exist. A repository that reads as open while accepting nothing wastes the time
of whoever believes it, and it wastes it after they have done the work.

Taking proposals as issues rather than as changes looks like the careful middle
and costs the cheapest thing this board offers. Trying something is the whole
proposition here, and under this option trying something first becomes a request
that somebody has to answer. The person with a question and an afternoon is the
contributor this board is for, and a queue between them and the work is a queue
they will not join. It also moves the residual rather than reducing it: the
three rules still have to be read before anybody runs anything, and a discussion
in an issue is not a mechanism that refuses a violation any more than the guide
is.

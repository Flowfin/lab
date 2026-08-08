# 0010. When an experiment finds a flaw in shipped software

## What was decided

Everything on this board is public from the first commit, and the questions this
board is for sit next to software people run. Sooner or later an experiment asks
whether something can be bypassed, overloaded or read that should not be, and the
answer is yes. At that moment this board's central rule and responsible
disclosure point in opposite directions. The rule says write the question before
the work and the answer when it stops, in public. Disclosure says the affected
project hears about it first.

### Which experiments do not start here in public

An experiment whose question is whether shipped software can be made to misbehave
in a way that harms the people running it does not start here in public. That
covers a question about bypassing an authentication or authorisation boundary,
about reading data the software is meant to keep from the reader, about crashing
or exhausting a service somebody is running, and anything else whose answer would
be a working way to hurt an operator or their users.

The question goes to the affected project through whatever private route that
project publishes, and the work happens where that project asks for it. This
board is not where such a flaw is first written down. No directory, no record, no
issue and no branch here carries it before the affected project has it.

An experiment that finds something unintended while asking a different question
takes the same route, and that is the common case rather than the exception. The
work stops, nothing sensitive is committed, and the report goes to the affected
project. A gap in the tree while that happens is cheaper than a public record of
a live flaw.

### What is written afterwards, and what has to be true first

What comes back here is the record. Two things have to be true before it is
written. The flaw has to be fixed, and the affected project has to have said what
it wants said.

Once both hold, an experiment record is written in the ordinary shape, with the
question, the answer and a pointer to the published advisory. That record is worth
having, because how a thing broke is exactly the sort of result this board exists
to keep, and by the time both conditions hold, publishing it costs nobody
anything.

How long such a record waits when the fix or the advisory never arrives is a
judgement rather than an engineering answer, and it is not decided here. What the
listing shows while a record is held back is part of the same question and is
also not decided here.

### Software this organisation does not ship

Software this organisation does not ship is outside this decision. The same
private route usually exists, this board does not set it, and whoever found the
thing follows that project's own policy. Where that project publishes no policy at
all, that is a fact about the project and not permission to publish here.

### What enforces this

Nothing. No check reads intent, and nothing in a checkout separates a question
about performance from a question about a bypass. There is no route here that
could refuse an experiment for the question it is asking, and this record does not
create one.

This is a rule a person follows. The only machinery behind it is that the security
policy and the contributing guide both carry it where somebody reads before
starting, so the person most likely to be about to break it is the person most
likely to have just read it. The mechanism is a reader. Calling it anything else
would be the assurance this repository refuses to make.

## What it applies to

Every experiment on this board, from the moment its question is chosen, and every
person working on one.

It applies to the question rather than to the finding. An experiment that was
never going to ask about a flaw and then trips over one is inside this decision,
which is why the paragraph about the common case is in the first section rather
than in a footnote.

It does not apply to a flaw in this repository's own runner, which has its own
route in the security policy.

## What else was considered

Publishing everything immediately, on the reasoning that the board's rule about
writing the question first admits no exception.

Writing the question here in public but holding back the detail, so the listing
shows that something is being asked without saying what.

Keeping a private area inside this repository for this class of work.

## What each rejected option would have cost

Publishing immediately costs the people running the software. A public record of
a live flaw is a working recipe handed to anybody who reads the board, before the
project that could fix it has heard about it, and the cost lands on operators who
had no part in the decision. It is also one commit and it cannot be taken back,
which is why the rule is set before it happens rather than during.

A public question with the detail held back costs more than it saves. A question
naming the software and saying that its handling of something is being examined
is most of a disclosure already, and it points a reader at exactly where to look
while the fix does not exist yet. It also produces a record that says nothing,
which is the shape this board most wants to avoid, and it would sit in the
listing looking like an ordinary experiment.

A private area inside this repository costs the one property the board is built
on. It would mean a tree where some things are visible and some are not, so a
reader could no longer take the listing as the whole of what is happening here,
and the exception would be available for anything anybody preferred not to show.
The private route the affected project publishes already exists and is already
the right place, so this would be a second one with weaker reasons.

Slug: the-slug-of-this-experiment
State: asking
Question-Written: 2026-01-01
Needs-Hardware: none

Copy this file to `experiments/<slug>/EXPERIMENT.md` and replace every value in
the header above. The slug is the name of the directory this record sits in. The
state is `asking` on the day that directory is created. The question date is the
day the question below was written, as `YYYY-MM-DD`, and it does not move
afterwards, because the listing sorts by it. Add `Answer-Written` in the same
change that writes the answer.

Add `Measurement-Commit` in that same change where the answer quotes a
measurement, with the object name of the commit the measurement was produced at,
written in full. It is neither here nor in `Answer-Written`'s position because a
template that ships a field filled in teaches every new record to declare a
value it does not have yet. The code the measurement ran against may be removed
later, and without this the answer keeps its numbers and loses the thing that
produced them while still reading as complete.

`Needs-Hardware` is what this experiment needs beyond the runner, in words
somebody deciding whether to reproduce it can act on, or `none`. It starts at
`none` here because that is the right answer for almost every experiment. A test
that genuinely needs a device is registered in the integration-hardware harness
under `internal/hardware`, in a file whose name ends
`_integration_hardware_test.go`, and a record that says one thing while the
directory says the other is refused.

Add `Borrowed` where the experiment starts from code somebody else wrote, naming
where that code came from and the licence it arrives under. It is not in the
header above for the reason `Measurement-Commit` is not: a template that ships a
field filled in teaches every new record to declare a value it does not have, and
almost every experiment borrows nothing. The code itself goes in
`experiments/<slug>/borrowed/`, which carries its own `LICENSE` naming those
terms, and a record declaring the field with no such directory is refused, as is
a borrowed directory with no licence file in it.

That directory is the only place a quarantine may be, and one is all record
`0019` allows, so a directory named `borrowed` anywhere else in the experiment
is refused too. One inside the quarantine is not: the code in there is laid out
by whoever wrote it and this board does not rearrange it.

What that refusal does not do is worth knowing before you rely on it. Nothing
reads the licence file, so a green run says the layout and the declaration agree
and says nothing about which licence the code is actually under or whether the
result may be promoted anywhere. A borrowed directory in an experiment whose
record declares nothing passes, because an absent field is never refused.

Add `Held-back` where the experiment is held back under
`docs/decisions/0010-a-flaw-in-shipped-software.md`, with the date of the report
the window is counted from, written as `YYYY-MM-DD`. It is not in the header
above for the reason `Measurement-Commit` is not, and here a value shipped
filled in would be worse than misleading: it would declare a hold on every
record copied from this file. The record stays in `asking` while it waits, the
listing prints that it is held back and when the clock started and nothing about
what it is about, and a value that is not a date is refused. How long the wait
is, what ends it and what the single extension costs are
`docs/decisions/0022-how-long-a-held-back-record-waits.md`, and `SECURITY.md` is
where that window is published for the project the report went to.

What that refusal does not reach is the case the field exists for. A record
being held back that declares no `Held-back` is not refused and cannot be, since
an absent field is never a refusal. It sits in `asking` with a question that
says nothing and appears in the listing as ordinary unanswered work, which is
the misreport the field is written against and which only the person writing the
record can prevent.

Add `Real-Data` where the experiment reads real personal data, naming what
category of data, on whose host, and what will be written down about it, or
write `none`. It is not in the header above for the reason `Measurement-Commit`
is not: a template that ships a field filled in teaches every new record to
declare a value it does not have, and `none` shipped filled in would be a claim
made by whoever copied the file rather than by whoever ran the experiment. It
goes in the commit that writes the question, because the point of it is that the
measurement was named while the answer was still unknown. Whose data may be read
and what has to be agreed in advance are
`docs/decisions/0025-real-data-in-an-experiment.md`, and `docs/privacy.md` is
where the same rule is written for whoever is running the thing.

What that refusal does not reach is the case the field exists for. A record
declaring the field with nothing after the colon is refused, and an experiment
that reads real data and declares nothing at all is not, since an absent field is
never a refusal. Nothing here reads whose machine the data was on, whether the
measurement was the agreed one, or whether the record carries the data rather
than a measurement of it.

The format is `docs/decisions/0008-the-experiment-record.md`, as added to by
`docs/decisions/0015-an-experiment-declares-the-harness-it-needs.md`, by
`docs/decisions/0016-an-answer-names-the-commit-it-measured.md`, by
`docs/decisions/0019-code-under-another-licence.md`, by
`docs/decisions/0022-how-long-a-held-back-record-waits.md` and by
`docs/decisions/0025-real-data-in-an-experiment.md`. This file is a
convenience and those records are the authority.

## Question

One question, written as a question, so that a reader who was not there can tell
whether it has been answered. A topic is not a question: an experiment whose
record says "media transcoding" can never be shown to have missed its answer.

## Method

What was done, in enough detail that somebody else could do it again or say why
they cannot. The commands, the machine where it matters, and what was measured.

What may never be committed alongside this record is
`docs/decisions/0006-everything-here-is-public.md`. Whether a question may be
answered against real data at all, and on what conditions, is
`docs/decisions/0025-real-data-in-an-experiment.md`.

## Answer

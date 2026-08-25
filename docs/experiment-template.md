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

What that refusal does not do is worth knowing before you rely on it. Nothing
reads the licence file, so a green run says the layout and the declaration agree
and says nothing about which licence the code is actually under or whether the
result may be promoted anywhere. A borrowed directory in an experiment whose
record declares nothing passes, because an absent field is never refused.

The format is `docs/decisions/0008-the-experiment-record.md`, as added to by
`docs/decisions/0015-an-experiment-declares-the-harness-it-needs.md`, by
`docs/decisions/0016-an-answer-names-the-commit-it-measured.md` and by
`docs/decisions/0019-code-under-another-licence.md`. This file is a
convenience and those records are the authority.

## Question

One question, written as a question, so that a reader who was not there can tell
whether it has been answered. A topic is not a question: an experiment whose
record says "media transcoding" can never be shown to have missed its answer.

## Method

What was done, in enough detail that somebody else could do it again or say why
they cannot. The commands, the machine where it matters, and what was measured.

What may never be committed alongside this record, and what happens where a
question can only be answered against real data, is
`docs/decisions/0006-everything-here-is-public.md`.

## Answer

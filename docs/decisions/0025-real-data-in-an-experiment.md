# 0025. Real personal data in an experiment

## What was decided

An experiment here may touch real personal data, under two conditions, and it
declares that it does before the work starts.

The data belongs to the person running the experiment. Their own library, their
own accounts, their own logs, on their own host. Not a colleague's, not an
employer's, not a user's, and not a copy of any of those that somebody was
allowed to hold for a different purpose.

The measurements are agreed in the record before the work starts. What is
measured is written into the record in the commit that writes the question, so
the reading that gets published was named while the answer was still unknown.
A measurement thought of afterwards is a new question and takes a new record.

The rule already written in [../privacy.md](../privacy.md) stands on top of
both, unchanged and not restated here: the data never enters the tree, in any
form, and what may be written down is the measurement.

This record supersedes 0008. Everything that record fixes about an experiment
record stands exactly as it is written there, and this adds one field to the
header.

`Real-Data` names what category of data, on whose host, and what will be
written down about it, or the single word `none`.

    Real-Data: my own media library, on my own machine, and what gets written down is the scan time and the item count
    Real-Data: none

The field exists because the two conditions above are conditions on something
nobody can see afterwards. An answer that says a library of a hundred thousand
items took four minutes to scan reads identically whether the library was the
author's own or an employer's, and whether the timing was the agreed
measurement or the one that came out best. The declaration is what makes the
difference visible at the moment it is still cheap to ask about, which is the
review of the commit that writes the question.

The field is optional, which record 0013 fixes for every field added after it,
so a record written before this one stays legal and an absent field is never
refused. What that costs is stated at the bottom of this section rather than
left for somebody to discover.

`none` is what a record says when the question is answered against generated or
synthetic data, and it is the answer for almost every experiment. Where
synthetic data can answer the question it does, and a record that needed real
data says why in its method. That sentence makes the easy path the default
without banning the hard one, and it is a sentence for a reader rather than a
rule for a machine.

One refusal follows, and it reads a declaration rather than an absence. A
record that declares `Real-Data` and writes nothing after the colon is refused
under `record-real-data-declaration-is-empty`. It names no data, no host and
nothing that will be written down, and it does not say `none`, so it carries
the appearance of a declaration and none of the content of one. That is the
mistake the shape invites: the field is typed while the question is being
written, the value is left for the moment the work starts, and the moment the
work starts is when nobody reopens the header.

WHAT NOTHING REFUSES, and this is most of the rule. An experiment that touches
real personal data and declares nothing at all is refused by nothing, because
record 0013 makes an absent field legal and the runner cannot tell a record
that omitted the field from one written before the field existed. Whose data it
was, whether the measurement was agreed in advance, and whether the record
carries the data rather than a measurement of it are all outside every reading
of a checkout, and the last of those is stated in `docs/privacy.md` as having
no mechanism and keeps not having one. What stands behind this rule is the
template, the person writing the record, and whoever reads the change.

## What it applies to

Every experiment on this board that touches real personal data, from the commit
this record lands on.

The experiment record format, which gains one optional field. Records already
on the default branch are not edited, not migrated and not marked, which record
0013 fixes and which this record does not vary.

It applies to `docs/privacy.md`, which said this question was open and now
carries the answer, and to `docs/experiment-template.md`, which carries the
field for records written from here on.

It does not apply to what may never be committed, which is record 0006 and is
unchanged by this. A declaration permits an experiment to read real data on the
host it is already on; nothing here permits a byte of it into the tree.

It does not apply to data that is not personal. A generated library, a public
corpus and a synthetic account set are outside this record, and an experiment
using them says `none` or says nothing.

## What else was considered

Refusing real personal data on this board outright.

Allowing it with no declaration at all, leaving the host rule in
`docs/privacy.md` to carry the whole of it.

Allowing data belonging to somebody else where that person consented.

Requiring the field of every record rather than making it optional.

A `## Real data` section carrying the three parts as named lines, in the shape
record 0005 gives the promotion section, rather than one header field.

## What each rejected option would have cost

Refusing it outright costs the board the questions it is most useful for. How a
library of a hundred thousand items behaves, how a real account set looks after
a migration and what a real log actually contains are answerable against real
data and against nothing else, and a board that cannot ask them is less useful
than it was opened to be. The ban would also not hold: the questions arrive
anyway, and the first one would be answered somewhere with no record of what
was touched, which is worse than the thing the ban was for.

No declaration costs the reviewer the only moment the question is cheap. The
host rule says what happens to real data if an experiment uses it, and says
nothing about whether this experiment did. A reviewer reading a finished answer
cannot tell, and asking afterwards asks somebody to remember what they did
rather than to confirm what they wrote.

Consent from a third party costs a promise this board cannot keep. Consent is a
thing with a scope, a date and a way of being withdrawn, none of which this
tree can hold or check, and a record asserting that somebody agreed is an
assertion nobody here can test. Restricting the rule to the runner's own data
gives up real questions and gives up a class of failure this board would have
no way to detect or repair.

Requiring the field turns every record written before today red on the day this
lands, and record 0013 exists to refuse exactly that trade. Its own words are
that the repair available under a red board is either editing permanent records
or weakening the check that just landed, and both destroy something.

The section costs a second syntax for a job the header already does, and it
would put the declaration below the fold, where the header is what a reader
scans and what every other field of this format lives in. The promotion section
earns its shape because it carries four values that each need their own line
and a commit range that has to be read; this carries one sentence. What the
section would buy is a refusal that could ask for the third part by name, and
that is not the residual worth paying for: whether words name a real category,
a real host and a real measurement is a judgement no reading of a checkout
makes either way.

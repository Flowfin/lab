# 0008. What an experiment record looks like

Superseded by 0014, which fixes what a slug may be, and by 0015, which adds the
field naming what an experiment needs beyond the runner. Everything below stands
as it was written.

## What was decided

Record `0003` says what an experiment record has to mean. This record says what
it has to look like, because a rule a machine cannot read is an explanation of a
rule rather than a rule.

`EXPERIMENT.md` opens with a small header of fixed fields and continues as
prose. The header is what the runner reads. The prose is what a person reads,
and nothing tries to parse it.

### The header

The header is the run of lines from the first byte of the file to the first
blank line. Every line in it is a field, written as a name, a colon, a space and
a value, with the name at column zero:

    Slug: sequential-write-on-spinning-disk
    State: asking
    Question-Written: 2026-08-09

Four field names are fixed here.

`Slug` repeats the name of the directory the record sits in, so a reader holding
a quoted slug can walk back to the experiment and a reader holding the record
can tell which directory it belongs in. What a slug may be in the first place is
not decided here; issue #59 fixes that shape and supersedes this record when it
does.

`State` is one of the three record `0003` names, written exactly as that record
writes them: `asking`, `answered`, `abandoned`.

`Question-Written` is the date the question was written, as `YYYY-MM-DD`. It is
written in the commit that creates the directory and it does not move
afterwards, because it is what the listing sorts by and a date that can be
edited sorts a stalled experiment wherever its author prefers.

`Answer-Written` is the date the answer was written, in the same shape. It is
absent until there is an answer and it is added in the same change as the
answer.

Order does not matter and a name appears at most once. A name this record does
not fix is allowed to appear: a field added by a later record is unknown to
every checker built before it, and a format that refused an unrecognised name
could only ever grow by breaking the checkers already in the tree.

### The prose

Headings at level two, in this order:

    ## Question
    ## Method
    ## Answer

Text may sit between the header and the first heading, and nothing reads it. A
title, a note to whoever picks the work up, or nothing at all. It is named here
so that a checker built against this record treats it as prose rather than as a
malformed section.

`## Question` carries one question, written as a question. One question and not
a topic, because a directory holding three questions has no state that is true
of all of them, and a record whose question is a topic can never be shown to
have missed its answer.

`## Method` says what was done, in enough detail that somebody else could do it
again or say why they cannot.

`## Answer` exists in every record from the day it is created, and it is empty
until there is an answer. That is the shape rather than an accident of the
template. Writing an answer is then filling in something that is visibly
missing, rather than remembering to add a heading whose absence nobody sees.

A promoted experiment also carries `## Promotion`, which record `0005` fixes the
contents of. It is added by the change that hands the work over and it is absent
from every other record.

### What the format is not

It is not a database. It is text somebody has to want to read, and a header of
twenty fields would be a form to fill in rather than a record to write. The
machine-readable part is kept to what the checks genuinely need, and a thing
that can live in the prose lives there.

A template lives at `docs/experiment-template.md` and carries the shape above
with nothing filled in. The template is a convenience and not the authority.
This record is the authority, and the runner parses the template in its own
suite so that the two cannot quietly disagree.

Nothing refuses a record over any of this at the commit this record lands on.
The runner can read the header and the sections, and that is the whole of it.
The refusals arrive with the issues that argue them: a record that never wrote
its question, a record claiming an answer it does not carry, a header that
disagrees with itself, and a slug that is not a legal slug. Writing that gap
down here is cheaper than a reader inferring from a green run that the format is
enforced.

## What it applies to

Every `EXPERIMENT.md` under `experiments/`, from the commit this record lands
on, and every check that reads one.

It applies to the template, which carries this shape and is checked against the
runner rather than trusted.

How this format changes afterwards is record `0013` and not this record. A field
added later is optional and a check over it refuses only what is present, which
is why the four names above are the whole of what a later checker may assume is
there.

It does not apply to the decision records in `docs/decisions/`. Those carry the
four sections record `0000` fixes, they hold no state, and they change by
superseding.

It does not apply to anything else an experiment directory holds. Code, data
files and notes beside the record are the experiment's own business.

## What else was considered

A header fenced as YAML front matter between `---` lines.

A larger header, carrying the question, the method and the answer as fields
rather than as prose.

No header at all, with the checker reading the level-two headings and inferring
the state from which of them carry text.

A separate machine-readable file beside the record, holding the fields, with the
record left as pure prose.

Refusing a field name this record does not fix.

## What each rejected option would have cost

YAML front matter costs a parser and a dependency, on a runner whose language
was chosen partly for carrying almost none. It also buys nothing the format
needs: the values here are a slug, a word from a set of three and two dates, and
none of them wants nesting, lists, quoting rules or the several ways YAML has of
writing a date. The cost that would land last is the worst of it, because a YAML
parser accepts documents this format has no meaning for, so the checker would
have to refuse valid YAML and explain why.

A larger header costs the record its readability, which is the thing it exists
for. A question written into a field is a question written to fit on one line,
and the method is where the detail that makes an experiment reproducible lives.
It would also put the answer inside the part a machine reads, which invites a
check over what the answer says, and no reading of a tree can judge that.

No header costs the state. A record read as answered because its answer section
has text can be answered by typing anything under the heading, and there would
be nothing for a check to compare the claim against, since the claim and the
evidence would be the same bytes. Record `0003` makes exactly that point: naming
the state separately from the answer is what makes "claims answered, carries no
answer" a thing a machine can see.

A separate fields file costs the one-file rule and buys nothing. An experiment
would carry two files that have to agree, a reader would have to open both to
know what they are looking at, and every rule about the record would have to say
which of the two it meant. It also puts the state somewhere a person editing the
prose does not see, which is how a record ends up claiming `asking` under a
finished answer.

Refusing an unfixed field name costs the format its ability to grow. Record
`0013` makes a new field optional so that records written before it stay legal;
refusing unknown names would break the same rule from the other side, because
every checker built before the new field would refuse every record written after
it. The cost of allowing them is that a misspelled field name is read as a new
field and passes, and that is a real cost, paid because the alternative makes
the format un-growable.

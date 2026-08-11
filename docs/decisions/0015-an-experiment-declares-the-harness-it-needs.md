# 0015. An experiment declares the harness it needs

## What was decided

This record supersedes 0008. Everything that record fixes about an experiment
record stands exactly as it is written there, and this adds one field to the
header.

`Needs-Hardware` names what an experiment needs beyond the runner, in words
somebody can act on, or the single word `none`.

    Needs-Hardware: a spinning disk with no other load on it
    Needs-Hardware: none

Record 0007 keeps the default run headless and unelevated, and a test that
cannot meet both halves goes to the integration-hardware harness under
`internal/hardware/` instead of skipping itself inside the default run. That
boundary already exists. What did not exist until this record is a way for a
reader to see which side of it an experiment sits on without opening its source,
and a way for a machine to notice when the record and the source disagree.

The field is optional, which record 0013 fixes for every field added after it,
so a record written before this one stays legal and an absent field is never
refused. What that costs is stated at the bottom of this section rather than
left for somebody to discover.

`none` is the reason the field can be checked in both directions. An absent
field and a field saying `none` are the same sentence to a reader and different
statements to a checker: the first says nothing, and the second is a claim the
tree can contradict. Without a word for it, an experiment carrying hardware
tests and declaring nothing would be indistinguishable from every record written
before the field existed, and half of the rule would be unenforceable by
construction.

Two refusals follow, and both read a declaration rather than an absence.

A record declaring hardware whose directory holds no test registered in that
harness is refused. The failure is ordinary: the tests were moved into the
default run once they stopped needing the device, and the record still tells a
reader they need a machine they do not need.

A record declaring `none` whose directory holds a test registered in that
harness is refused. This is the direction that costs somebody an afternoon. The
declaration says the result can be reproduced anywhere, the tests say otherwise,
and the person who finds out is whoever cloned the repository to check the
answer.

A record declaring the field and writing nothing after the colon is refused as
well, and it is a separate rule with a separate repair. It names no hardware and
it does not say `none`, so there is nothing for either direction above to
compare against, and reading it as one of the two would be a guess written into
a checker.

What a test registered in that harness is, exactly. The harness's files are
named `*_integration_hardware_test.go` and sit behind the
`integration_hardware` build constraint, which is the convention
`internal/hardware/` already holds itself to and reads its own source for. This
record fixes the same name as the marker inside an experiment directory, so a
file that registers a test with the harness is a file whose name says so, and
nothing has to parse Go to find one.

What this does not judge, and it is most of what the field says. Whether the
words name the hardware the tests actually need is a reader's judgement.
`Needs-Hardware: a laptop` satisfies every rule here and tells nobody anything.
The check holds the two claims to not contradicting each other, and the review
is where a declaration that says nothing useful is caught.

What the optional field costs, in one sentence, because it is the part somebody
will meet later: a record that omits the field entirely is not held to either
direction, so an experiment can carry hardware tests and stay silent about them,
and no run refuses it.

## What it applies to

Every `EXPERIMENT.md` under `experiments/`, from the commit this record lands
on, and every check that reads one.

It applies to the template at `docs/experiment-template.md`, which carries the
field with `none` filled in, because the template is what most records are
copied from and the value that is right for almost every experiment is the one
that should already be there.

It applies to the listing, which prints what each record declares. A reader
scanning for a result they could reproduce themselves is asking exactly this
question, and the answer is one column rather than a file to open.

It does not apply to the runner's own tests. `internal/hardware/` is the harness
itself and is not an experiment, and the rules about a record reach nothing
outside `experiments/`.

It does not apply retroactively. Record 0013 fixes that, and the absence of the
field in a record written earlier is not a defect in that record.

## What else was considered

A required field, so that every record says which side of the boundary it is on.

Deriving the answer from the tree alone, with no field at all, by reporting
which experiments hold a file registered in the harness.

A boolean field rather than a field naming the hardware in words.

Reading the build constraint inside each Go file rather than the file's name.

## What each rejected option would have cost

A required field costs what record 0013 spent a whole record refusing to spend.
Every record written before the field existed would be refused on the day it
landed, the default branch would be red, and the two repairs available would be
editing records that are supposed to be permanent or weakening the check that
had just landed. The rule is worth less than that, and the same coverage arrives
one record at a time through the template.

Deriving it from the tree costs the declaration, which is the thing worth
having. A derived answer is always true of the tree and says nothing about what
the author meant, so the disagreement between the two, which is the failure this
record exists to surface, becomes impossible to state. It would also make the
record's own claim unfalsifiable: with nothing declared, nothing can be wrong.

A boolean costs the reader the only part they can act on. Knowing that an
experiment needs hardware and not which hardware leaves somebody deciding
whether to reproduce it exactly where they started, and the harness already
holds its own tests to naming what they need in words for that reason.

Reading the build constraint costs a Go parser pointed at whatever anybody put
in an experiment directory, on a runner whose input is untrusted and whose
language was chosen partly for reading almost nothing. It buys very little: a
file behind the constraint and named for it is the convention the harness
already enforces on itself, and a file that carries the constraint under some
other name is not registered with the harness by any route this repository has.

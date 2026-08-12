# 0016. An answer names the commit its measurement was produced at

## What was decided

This record adds a field to the experiment record format record `0008` fixes,
and it supersedes that record for that one thing. Everything `0008` says stays
as it was written, and how the format grows at all is record `0013`, which this
answers to rather than restates.

An answer that quotes a measurement carries `Measurement-Commit` in the header,
whose value is the object name of the commit the measurement was produced at,
written in full.

    Measurement-Commit: e5067b1d0f4a2c8b6e3a9d7c1b5f8a2e4d6c0b93

The field is optional, because record `0013` makes every field added after it
optional and an absence is never refused. What is refused is a value that is not
the shape of an object name, which is the whole of what a checkout can be asked.

The full name rather than an abbreviation. Record `0004` already writes the
commit that removed an experiment's code with the full hash, for the reason an
abbreviation stops being unique as a repository grows, and a measurement is the
same shape one step earlier. A reader who cannot resolve the name they were
given is in the position this field exists to remove them from.

Why a field rather than a sentence in the prose. An answer already names the
command, the platform, the architecture and the toolchain, and none of that is
read by anything. What is missing is the version of the code the command ran
against, and record `0004` lets that code be removed entirely afterwards. When
it is, the answer keeps its numbers and loses the thing that produced them, and
it does so silently, because the record still reads as complete. A field is
where the runner already looks.

What this does not buy, written here because a green run will otherwise be read
as more than it is. The runner reads a checkout and opens no connection, so
nothing here asks git whether the object named is in this repository, or whether
it is a commit, or whether the command in the answer was ever run against it. A
record naming forty hexadecimal characters that resolve to nothing passes.
What the refusal converts is the case where the value could not be resolved by
anybody, which is the only case a checkout can separate from the rest.

And the absence stays unrefusable. A record quoting a measurement and carrying
no `Measurement-Commit` is legal, so the field makes the fact recordable and
never guaranteed. That cost is record `0013`'s, paid deliberately, and the
things that make a field usual without making it required are the template, the
review, and this record.

## What it applies to

Every `EXPERIMENT.md` under `experiments/`, from the commit this record lands
on, and the check that reads the field.

It applies to what is present. A record already on the default branch is not
edited, not migrated and not marked, which is record `0013` and is why the one
record on the board when this landed was left exactly as it was.

It applies to the template, which names the field in its prose rather than
carrying it in its header, for the reason `Answer-Written` is named the same
way: a template that ships a field filled in teaches every new record to declare
something it has no value for yet.

It does not apply to the decision records in `docs/decisions/`. Those carry the
four sections record `0000` fixes and change by superseding.

It does not decide anything about a measurement produced somewhere other than
this repository. An answer measuring another project's code has the same
problem and a different answer, and nothing here should be read as covering it.

## What else was considered

A convention in the answer prose, in the words the record already uses, with
nothing reading it.

Nothing at all, on the argument that a reader who wants the exact code has git
and the record's own history.

A field carrying the date of the measurement rather than the commit.

A field required of every record whose answer quotes a command.

Naming the commit with an abbreviated hash.

## What each rejected option would have cost

A convention costs exactly what it is: nothing reads it. The failure this is
about is a record that still reads as complete after the thing that produced its
numbers has gone, and a convention is another sentence in the file that already
read as complete. It also drifts, because two authors writing the same
convention in their own words produce two shapes, and the first check anybody
writes over it has to be right about both.

Nothing costs the reader the work, and it charges the reader least able to do
it. Somebody who was there can find the commit from the date and the log.
Somebody who was not is the person this whole record is for, and the answer they
get is that the numbers were true of some version of a directory that record
`0004` may have removed.

A date costs the precision that makes the field worth having. Several commits
land on one day here, and the answer already carries `Answer-Written`, so a date
would duplicate a field that exists and still not say which code ran.

Requiring the field costs the format its ability to grow, which is record
`0013`'s whole argument. It would also require the checker to decide what counts
as an answer quoting a measurement, which is a judgement about prose that no
reading of a tree makes: an answer holding a number and an answer holding a
number that was measured look the same to everything in this repository.

An abbreviation costs uniqueness at the moment somebody needs it. A short name
resolves today and stops resolving when the repository grows into a collision,
and the reader who meets that is the one who came back years later, which is the
case the field exists for. It also cannot be told from a typo: seven characters
that resolve to nothing and seven characters somebody mistyped are the same
string.

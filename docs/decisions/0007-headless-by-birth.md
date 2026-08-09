# 0007. Headless and unelevated, as a birth requirement

## What was decided

Every test in the default run completes with no graphical session and as an
ordinary user. Both halves bind from the first test this repository has, not
from the milestone that builds the mechanisms enforcing them.

The word that matters is birth. The test harness the scaffolding milestone
builds already owes a green run with no display and no elevated privileges, so
the requirement is load-bearing before the work that depends on it starts. A
requirement written down after the first thing that depends on it has been
retrofitted, whatever the document later says about it, and a retrofitted
requirement is one that has already been negotiated with once.

No display. Every test in the default run completes on a machine with no
graphical session available at all. That is what makes the suite runnable on a
build machine, inside a container, and over a remote shell, which between them
are most of the places it will ever run. A suite that needs a display runs on
one machine, and a suite that runs on one machine stops being evidence the first
time that machine is unavailable and somebody says the tests passed yesterday.

No elevation. Every test in the default run completes as an ordinary user. A
test that needs administrator rights is a defect rather than a step to write
into the instructions, and the reason is worth stating because the failure is
subtle rather than loud. Where elevation is needed it is usually needed for one
narrow thing, such as binding a socket to a real interface address instead of
loopback, and the operating system responds by asking a question only an
administrator can answer. On some systems that question is asked about the
executable path rather than about the project, so answering it once settles
nothing for the next build directory. The result is a suite that intermittently
stops on a dialog nobody is watching, on somebody else's machine, which is
indistinguishable from a hang.

A test that genuinely cannot meet both halves does not get an exception. It
moves to the separate hardware harness the headless milestone builds, which is a
different thing with a different name and is not part of the default run. The
default run stays clean, and a reader who sees it green knows what was covered
because the things it cannot cover are somewhere else with their own name rather
than skipped inside it. That harness does not exist at the commit this record
lands on, and the record names the rule rather than the mechanism. That is the
right order rather than a gap in it: the rule is what the harness will be built
to serve.

What the rule does not buy. Neither half is a security property. Running as an
ordinary user is not a sandbox, it does not make the runner safe to point at an
untrusted tree, and it does not bound what a test can read or write inside the
account it runs as. Running without a display is not isolation either. Both
halves are about where the evidence can be produced, so that a green suite means
the same thing on every machine that runs it, and neither says anything about
what the code under test is allowed to do. Treating this record as a security
control would be reading a portability rule as a boundary.

## What it applies to

The default run: whatever this repository's suite executes when somebody runs it
with no arguments and no opt-in, in a checkout or on a build machine.

It applies to the record checks and to the runner's own tests, which are what
runs on its own under record `0009`. It applies to any test added later that
lands in the default run, whoever adds it and whatever it is for.

It does not apply to the hardware harness. That harness exists precisely for the
tests that cannot meet these halves, and holding it to them would leave nowhere
for such a test to go.

It does not apply to an experiment's own tests. Record `0009` keeps every
automatic run out of `experiments/`, so nothing here runs them and this record
makes no claim about them. An experiment that can only be answered on a machine
with a screen attached is a legitimate experiment.

## What else was considered

Requiring it only of the checks that run automatically, and letting a test that
somebody runs by hand need a display or elevation.

Allowing a test to require elevation and documenting the requirement, so that
whoever runs the suite knows to start it as an administrator.

Allowing such a test to skip itself when it detects no display or no elevation.

Deferring the rule to the milestone that builds the harness, on the reasoning
that a rule with no mechanism is only prose.

## What each rejected option would have cost

Restricting the rule to automatic runs costs the case it exists for. The place
where a display requirement hurts is a contributor's machine, a container, or a
remote shell, and those are all hand runs. A rule that binds the build machine
and releases the contributor produces a suite that is green in the one place
nobody was worried about.

Documenting an elevation requirement costs the evidence. A suite somebody starts
as an administrator is a suite that runs differently on the machine of whoever
did not read the instruction, and it moves the requirement onto every person who
ever runs it rather than onto the one person who wrote the test. It also means
the consent question above gets asked in the ordinary course of work, which
trains whoever sits at the machine to answer it without reading it.

A self-skipping test is the most expensive of the four, because it looks like
the cheapest. The suite stays green while covering less, and the amount less is
invisible unless somebody reads the run output carefully enough to notice a skip
that has become normal. That is worse than a red test, because a red test is
acted on. Issue #31 makes a skip in the hardware harness say what was missing,
which is the same problem handled in the place where a skip is legitimate.

Deferring the rule costs the word birth and nothing else, which is the whole
point. The mechanism is not what this record is; the mechanism is issue #29,
which proves the default suite runs with no display and no elevation. A rule
written after the first suite exists is a rule the first suite was not built to,
and every test written in between is one somebody has to go back and check.

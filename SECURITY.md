# Reporting a security problem

Report privately, through the form under Security on this repository:

    https://github.com/Flowfin/lab/security/advisories/new

The route is a repository setting rather than a file, and it does not answer
today:

    gh api repos/Flowfin/lab/private-vulnerability-reporting
    {"enabled":false}

Run 2026-08-19. So this file names the destination and says plainly that it is
shut, rather than sending a reporter to a door that does not open. The address
is the one the form uses once the setting is on, so nothing here moves when it
changes. Until then the honest alternative is a public issue carrying as little
detail as the report can carry: the software, the shape of the problem, and an
offer to send the rest privately. That is a poor arrangement and is written
down as one. Issue #10 is where this file lands and where the gap belongs.

With no policy in this repository, what a reader has been shown until now is
the organisation default, which says the private form is enabled on every
repository here and that a first answer arrives within a few days. Neither is
true of this repository today. This file is what governs this one.

## What is here, because the description undersells it

The description calls it experiments that do not have to finish, and that is
what the board is for. It is not most of what is in the tree, which is a Go
toolchain that judges this repository, plus one experiment. `cmd/lab` walks a
checkout and reports what it examined, writing nothing to the tree it reads.
`cmd/pullrequest` reads the event payload the platform writes and asks git
about the range. `cmd/contexts` compares the check names the workflows declare
against the contexts the merge gate requires. `cmd/notices` renders the
third-party notices a binary owes, from the module set the binary records and
the licence texts in a module cache. There are no dependencies: `go.mod`
requires nothing and there is no `go.sum`.

## What somebody could actually report

A way past a bound the walk sets. Everything under `experiments/` is written by
whoever proposes an experiment, so the runner treats that tree as input rather
than as its environment: a record above 64 KiB is refused unopened, the walk
stops at sixteen levels, a symbolic link under that directory is refused rather
than followed, and a path named in a record that resolves outside the checkout
is refused rather than read. A tree that gets past any of those, and makes the
runner read a file outside this repository or read an unbounded amount, is the
report I most want. The runner is not a sandbox and does not claim to be one:
it runs with the privileges of whoever started it.

A value from a pull request that reaches git as something other than an
argument. The two ends of the range come out of the event payload, they are
held to being hexadecimal object names before they are passed on, and git is
invoked with an argument list rather than through a shell. Either of those
failing is a real finding.

A module path that walks out of the cache. `internal/notices` reads licence
texts from a directory laid out like the module cache, and module paths arrive
from the table inside a binary, written by whoever built it. A path that
escapes the cache root and gets some other file on the machine reproduced as
somebody's licence is in scope.

A workflow that grants more than its job needs, or that can be made to run a
fork's code with a token. Nothing here is triggered by `pull_request_target` or
`workflow_run`, which I checked across all thirteen files in
`.github/workflows` today, every workflow declares an empty or read-only scope
at the top level, `persist-credentials` is off at every checkout, and the
actions are pinned by commit. Anything that undoes one of those is worth
reporting before something uses it.

A break in a claim this repository publishes. It says the runner opens no
network connection and writes nothing to the tree it walks, and
`cmd/lab/network_test.go` holds the first of those. Making either sentence
false is a security problem here, because those two are most of what a reader
is asked to take on trust.

A live token or key committed by accident. Everything here is public from the
first commit, so that is the only confidential thing the tree can hold, and it
is worth telling me privately even though the commit is already public.

A gate on the mainline made to report clean over a tree it should have refused.
That is a different thing from a check being wrong, which is an ordinary issue.

## An experiment is not software to run against anything you care about

An experiment here is a question and whatever code answered it. It is not
supported software, nobody is asked to install it, and it may be insecure by
construction, because sometimes that is what the question was. Nothing under
`experiments/` should be pointed at a production server, at a real media
library, or at anything holding data that matters. That is the scope of this
policy rather than a disclaimer attached to it. Nothing automatic builds or
runs any of it either, which decision record 0009 fixes, so an experiment
carrying dangerous code is not a route into this repository's automation: a
person who runs one does so in their own checkout, when they want the answer.

## An experiment that finds a flaw in software somebody runs

That one does not come here at all, and the public-issue alternative above does
not apply to it. Decision record 0010 is the rule: an experiment whose answer
would be a working way to hurt an operator or their users goes to the affected
project through whatever private route that project publishes, and nothing here
carries it before that project has it. That covers tripping over something
while asking a different question, which is the common case. The work stops and
the report goes to them. What comes back here afterwards is the record, once
the flaw is fixed and the affected project has said what it wants said.

A problem in Jellyfin itself belongs to
[the Jellyfin project](https://github.com/jellyfin/jellyfin/security/policy). A
report that lands here instead is pointed the right way rather than closed.

## What is not a vulnerability here, and why the list looks empty

There is no server, no socket, no account, no session, no database and no
stored personal data anywhere in this repository. The runner reads files and
prints what it found, so most of what a reader arrives with does not exist
here: nothing to log in to, nothing to escalate into, nothing to enumerate, and
no request path to inject anything into. Nothing is published as a binary
either. `gh api repos/Flowfin/lab/releases` and `gh api repos/Flowfin/lab/tags`
both answer with an empty list today, so there is no artefact anybody
downloaded and none to have been tampered with.

A refusal you disagree with is not a vulnerability, and neither is a check that
is too strict about a legitimate tree, nor a supply-chain score lower than you
expected, which `docs/supply-chain.md` already triages check by check. Those
are issues and they are welcome as issues. Nor is anything about the licence.
[LICENSE](LICENSE) at the root carries the GNU General Public License version
3, which is the answer record 0018 writes down, and it is what this repository
declares to the checks that read a licence. Both of those are legal or housekeeping
questions rather than security ones.

## What a reporter gets

Every report is answered, including the ones that turn out not to be problems,
and one of those gets the reason it is not, which costs nothing to receive and
takes the guessing out of a silence. Credit goes to the reporter unless they
ask otherwise, and a working exploit is useful in the report and not in public
before there is a fix.

There is no response deadline and there will not be one. Nothing holds me to a
window, so stating one would be a promise that goes quietly wrong on the first
busy week, and a reporter told to expect an answer by a date who does not get
one is left guessing whether the report arrived at all.

What is covered is the default branch as it stands. There are no releases and
no tags, no version is maintained in parallel, and nothing to backport a fix
to: a fix is a commit on `main`, and whoever wants it takes a newer checkout.

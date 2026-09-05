# Reporting a security problem

Report privately, through the form under Security on this repository:

    https://github.com/Flowfin/lab/security/advisories/new

That is the whole channel. A report sent anywhere else is not received: no
mailbox is published for this board, and there is no second private route to
fall back to. A report about somebody's behaviour rather than about the software
takes the same form, and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) is where that is written down.

The route is a repository setting rather than a file, and this paragraph said
until now that it does not answer. It answers:

    gh api repos/Flowfin/lab/private-vulnerability-reporting
    {"enabled":true}

Run 2026-09-05, against the same repository as the earlier reading of the same
command, which returned `{"enabled":false}` on 2026-08-19 and is what the
paragraph here was written around. The address did not move when the setting
changed, which is why it was safe to publish while the door was shut.

WHAT THAT ENDS IS THE ALTERNATIVE THIS FILE USED TO OFFER. It said that until
the form opened, the honest fallback was a public issue carrying as little
detail as a report can carry. There is no such fallback now and there should not
be one: a public issue about a vulnerability is the outcome a security policy
exists to prevent, and it was written down here as a poor arrangement while it
was the only one. Do not open one. The form above is where a report goes.

The organisation default policy is still served on repositories in this
organisation that carry no policy of their own, and it says the private form is
enabled on every repository here. That is now true of this one. This file is
what governs this one, and where the two differ this one wins.

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
`workflow_run`, which I checked across all fifteen files in
`.github/workflows` today, every workflow declares an empty or read-only scope
at the top level, `persist-credentials` is off at every checkout, and the
actions are pinned by commit. Anything that undoes one of those is worth
reporting before something uses it.

A published artefact that does not match what this board says it published.
There is a release now, and every file in it is covered by `SHA256SUMS` with
`SHA256SUMS.sig` over that file, verified against the signing keys the platform
publishes for the account that cut it rather than against a key shipped beside
the signature. The release notes carry the two commands. A downloaded file whose
digest does not match its line, a checksum file whose signature does not verify,
or a release asset that appeared or changed after the release was published, are
each a report I want. Which of those is a compromise and which is a mistake on my
side is not something a reader can tell from outside, and that is exactly why it
comes here rather than being assumed to be the second.

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

How long that record waits is 90 days from the report, and it is published at
the end of them whether or not the fix and the statement have arrived. Waiting
on those two indefinitely would hand the schedule to whoever is slowest to
reply, and it would leave this board holding a record of a real finding that
nobody outside knows exists. There is exactly one extension, granted on a
reasoned request and written down, because a date fixed 90 days before anybody
had looked at the flaw is sometimes the wrong date for it, and moving that date
should be a choice somebody took rather than a slip. The reasoning, and what
each rejected option would have cost, is in
[decision record 0022](docs/decisions/0022-how-long-a-held-back-record-waits.md)
and is not restated here. Where the affected project publishes its own
disclosure policy and this board has reported into it, the earlier of the two
dates is the one that binds here.

That window is not a deadline anybody reporting to this board gets, and it is
not the sentence under "What a reporter gets" below saying there is no response
deadline. The two run in opposite directions and both stand. That one is about a
report arriving here, and how long a reporter waits for an answer from me. The
90 days are about a report leaving here, and how long this board holds back its
own record of a flaw it found in somebody else's software. A window on what this
board owes others is not a promise about what others may expect from it.

A problem in Jellyfin itself belongs to
[the Jellyfin project](https://github.com/jellyfin/jellyfin/security/policy). A
report that lands here instead is pointed the right way rather than closed.

## What is not a vulnerability here, and why the list looks empty

There is no server, no socket, no account, no session, no database and no
stored personal data anywhere in this repository. The runner reads files and
prints what it found, so most of what a reader arrives with does not exist
here: nothing to log in to, nothing to escalate into, nothing to enumerate, and
no request path to inject anything into.

WHAT THIS PARAGRAPH USED TO SAY NEXT WAS THAT NOTHING IS PUBLISHED AS A BINARY,
AND IT QUOTED TWO EMPTY LISTS FOR IT. Both answer with one entry now:

    gh api repos/Flowfin/lab/releases --jq 'length'
    1
    gh api repos/Flowfin/lab/tags --jq 'length'
    1

Run 2026-09-05. So there is an artefact somebody can have downloaded, and the
sentence saying there is none to have been tampered with has stopped being true.
What replaces it is the paragraph above about a published artefact, which is a
report I want rather than a class this file waves away.

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

What is covered is the default branch as it stands, and the most recent release.
This paragraph said there were no releases and no tags; there is one of each, and
the reading is in the section above. No version is maintained in parallel and
there is nothing to backport a fix to: a fix is a commit on `main`, the release
after it carries the fix, and whoever wants it takes a newer checkout or a newer
artefact. An older release is not patched in place and its assets are not
replaced, because replacing a file that a published checksum and signature
already cover is the shape a reader has no way to tell from tampering.

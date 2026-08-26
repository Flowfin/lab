# 0018. The licence of this board

## What was decided

GPL-3.0, and one licence for both the runner and the experiment content.

The reason is the direction the work actually moves. What a result on this
board is for is being taken somewhere it keeps running, and the place that
usually is sits in this organisation already: the plugin boards, every one of
which carries GPL-3.0 today. The platform fills that field in from the licence
file it finds at a root, and it is read one repository at a time:

    gh api repos/Flowfin/jellyfin-plugin-sso --jq '.license.spdx_id'
    GPL-3.0

Read that way when this record was written, the twelve plugin boards carry
GPL-3.0 and three other boards of this organisation carry AGPL-3.0. So this is
a choice about which of two neighbourhoods this board hands work to rather than
a house style everything here already follows. That population moves, and what
derives it rather than trusting the sentence above is:

    for r in $(gh repo list Flowfin --limit 100 --json name --jq '.[].name'); do
      spdx=$(gh api "repos/Flowfin/$r" --jq '.license.spdx_id // "none"')
      printf '%s %s\n' "$r" "$spdx"
    done

The same licence on both sides is what makes the hand-over cheap. Record `0005`
fixes promotion as a hand-over that names the terms the code leaves under, and
naming terms the receiving board already carries is a line somebody writes,
while naming different terms is a relicensing question somebody has to answer
before any code moves.

One licence rather than two, and the question is real rather than rhetorical.
The runner is downloaded and run, the experiment content is read and copied
from, and those are different audiences. What a second licence would buy is a
better fit for one of them. What it costs is that every reader of this tree has
to work out which of two sets of terms the thing in front of them is under
before they can use it, on a board whose whole purpose is lowering the cost of
trying something. One file governing everything is the answer, and record `0019`
is the one exception to it: code under another licence lives in a quarantined
directory inside the experiment that borrowed it, under the terms its own
licence file names.

**What this costs, in the direction it costs it.** A result that would have been
useful to a permissively licensed project cannot go there. An experiment that
wants to borrow permissively licensed code and give something back is
constrained, because what it gives back leaves under these terms. Neither is
hypothetical. They are the ordinary price of choosing the copyleft side, and
they are written here so that somebody meeting them later meets a decision
rather than a surprise.

**The file on the default branch says something else today, and this record
does not change it.** The platform reads the bytes at the root and reports what
it finds:

    gh api repos/Flowfin/lab --jq '.license.spdx_id'
    AGPL-3.0

So the answer above and the file disagree until issue #47 lands the file this
answer names, sets the licence this repository declares to the checks that read
one, and takes `README.md` with it. What this repository declares is a separate
thing from what the platform detects, and it is empty:

    git grep -n 'const DeclaredLicence' origin/main \
      -- internal/invariants/invariants.go
    origin/main:internal/invariants/invariants.go:81:const DeclaredLicence = ""

A reader arriving between this record and that change should trust neither the
sidebar nor this record alone. The sidebar says which file is there, this record
says what was decided, and the two agree only once #47 has landed.

**Adding a licence does not place the commits that came before it under those
terms.** Everything committed until that file lands arrived with no inbound
terms at all, which is a fact about copyright rather than about this tree, and
no file added at the root settles it afterwards. Repairing it is a question for
whoever holds the copyright in those commits. It is written here because the
alternative is a reader assuming the whole history is covered, and a reader who
assumes that is exactly the reader who would rely on it.

## What it applies to

Everything this repository holds, from the commit the licence file lands on: the
runner under `cmd/` and `internal/`, the experiment content under
`experiments/`, and the documents and records under `docs/`.

It applies to issue #47, which lands the file, the declaration and the line in
`README.md`, and it is what that issue reads to know which licence it is
landing.

It applies to promotion. The licence a promotion section names under record
`0005` is this one, unless what is being handed over is borrowed code under
record `0019`, in which case it is that directory's own.

It does not apply to borrowed code. Record `0019` puts code under another
licence in a quarantined directory carrying its own licence file, and that is
the one place in this tree where the answer above is not the answer.

It does not apply to the commits made before the file lands, for the reason
stated above. It also does not decide which licences may be borrowed from. Some
of them are incompatible with a hand-over into a GPL board whatever directory
the code sits in, and record `0019` says that is a separate question and does
not answer it either.

## What else was considered

Apache-2.0.

MIT.

CC0 or a public-domain dedication.

Leaving the repository unlicensed.

## What each rejected option would have cost

Apache-2.0 is permissive and carries an explicit patent grant, which matters
more than it looks for work touching authentication and media handling. Under it
a result could move into a GPL plugin board and also anywhere else. What it
costs is that the flow only runs one way: code borrowed from one of those plugin
boards cannot come back into an Apache-2.0 experiment without the experiment
becoming GPL, and an experiment that starts from existing plugin code is a
question this board expects to be asked. It buys reach in the direction this
board rarely goes and charges for it in the direction it goes most.

MIT has the same permissive shape as Apache-2.0, is shorter and is more widely
understood, and what it costs beyond Apache-2.0 is exactly that patent grant. On
work about authentication and media handling the grant is the part of
Apache-2.0 worth having, so taking MIT over it pays the same one-way-flow cost
and gets less for it.

CC0 or a public-domain dedication removes the attribution obligation entirely,
which suits a throwaway prototype and is the closest thing to no terms at all
that still counts as terms. What it costs is that some organisations will not
accept public-domain-dedicated code, because the dedication is not effective in
every jurisdiction, so the option chosen to simplify the promotion path can
block the promotion path instead. A cost that lands only sometimes, and only at
the moment the work is wanted, is worse than one that lands predictably.

Leaving the repository unlicensed costs everything the other options are being
compared on, and it grows rather than holding steady. Default copyright applies,
nobody has permission to reuse anything here, the sign-off gate asks
contributors to certify their right to contribute under terms that do not exist,
and every day it stays that way another contribution arrives with no inbound
terms. It is also the only option on the list that is not a decision. It is what
happens when nobody takes one.

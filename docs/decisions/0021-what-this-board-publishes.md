# 0021. What this board publishes

## What was decided

This board publishes downloadable release artefacts. They are signed. The
release notes say, in the notes themselves rather than only here, that the
artefacts exist for checking this board and are not part of what the
organisation ships to users.

The reason is the independent-verification argument, which is the whole point of
handing somebody a runner at all. The runner's job is reading a checkout and
reporting what it examined, and somebody who has to build that runner from the
repository they are checking is in a weaker position than somebody who downloads
it. Most people who might want to check this board's claims have no toolchain.
Publishing source only would have cost exactly that, and it would have left the
verification argument resting on the reader trusting the build they made from
the tree under examination.

The tension with the scope this board opened with is real and is not resolved by
being unwritten. That scope excludes anything a user is asked to install. What
separates the two is who the operator is: somebody checking this repository,
rather than somebody using a media server. The release-notes sentence is what
carries that distinction to the person holding the file, and its weakness is
that it depends on being read. Both halves are stated here so that a later
reader meets the residual rather than the assurance.

The artefacts are signed rather than published with a bare checksum. A checksum
published next to the file it checksums proves the download arrived intact and
says nothing about who built it, which is the claim somebody verifying this
board actually needs. The keys are the ones operations#1609 sets up for the
working accounts, and that is what makes this one key-custody story rather than
two: record 0023 requires a verified signature on a commit from the same keys.
Where those keys are held and how they are rotated belongs to that issue and is
not restated here, because a custody story written twice is a custody story with
two answers.

What publishing costs, named rather than discovered later. Signing. Checksums.
A bill of materials. A vulnerability surface, which a repository publishing
nothing does not have. And an expectation of continuity, which is the one that
cannot be withdrawn once somebody is running a version.

## What it applies to

Every release this board publishes, from the commit this record lands on, and
the issues that build them: the release workflow, the third-party notices and
bill of materials, the smoke run against a published artefact, and the first
release.

It does not apply to what an experiment produces. An experiment's prototype is
not published, is not something anybody is asked to install, and leaves this
board by promotion, which is
[0005-how-a-result-leaves.md](0005-how-a-result-leaves.md).

It decides that the artefacts are signed and does not decide what a signature is
checked with, by whom, or what a reader does when the check fails. A published
signature nobody verifies is a file beside a file.

## What else was considered

Publishing source only, with the tool run from a checkout.

Publishing artefacts with a checksum beside each one and no signature.

Publishing artefacts with no release-notes sentence separating them from what
the organisation ships to users.

## What each rejected option would have cost

Source only costs the reader a toolchain, which most people who would want to
check this board do not have, and it weakens the argument the publishing exists
to make. Somebody who builds the tool from the repository they are checking has
verified that the tree builds, not that the tree is what it claims to be. The
option is cheap for this board and expensive for exactly the person it was
supposed to serve.

A checksum with no signature costs the claim that matters. It proves the bytes
arrived as they left and nothing about where they left from, so anybody who can
place a file can place a checksum beside it. It is also the cheaper option only
until entry seven of the plan is answered, since a signed commit needs a key
held somewhere regardless, and taking it would have meant writing the custody
story once for commits and discovering later that artefacts needed their own.

Publishing with no separating sentence costs the scope exclusion its meaning. An
artefact with no such note is read as something the organisation ships, which is
the reading the exclusion was written against, and the cost of correcting that
reading arrives after somebody has already deployed against it. The sentence is
weak because it depends on being read; absent, there is nothing to depend on.

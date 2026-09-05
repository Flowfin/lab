# Changelog

One entry per release, newest first. An entry says what changed and what it
means for somebody moving from the release before it.

## What this board promises about compatibility

Nothing, and that is the useful answer rather than an evasive one. This is a
board for experiments that do not have to finish, the runner exists to read this
repository's own records, and
[CONTRIBUTING.md](CONTRIBUTING.md) says that nothing on another board is allowed
to depend on anything here. A version number implying guarantees I do not intend
to keep would be worse than no number at all.

What I do undertake is that a change is not made silently. That is the one thing
publishing an artefact creates that cannot be withdrawn afterwards, which
[docs/decisions/0021-what-this-board-publishes.md](docs/decisions/0021-what-this-board-publishes.md)
names as a cost of publishing at all: once somebody is running a version, they
have an expectation of continuity whether or not anybody offered them one. This
file is what that expectation is answered with.

The experiment record format may change. A change to it is decided in a record
under [docs/decisions/](docs/decisions/), and what happens to the records already
on the default branch on the day it changes is
[docs/decisions/0013-how-the-record-format-changes.md](docs/decisions/0013-how-the-record-format-changes.md),
which is where that was decided rather than restated here. A change that would
invalidate records already on the default branch is announced in the entry for
the release that carries it, under a heading naming it as such, and this file is
where somebody upgrading looks for that heading.

The exit codes the runner returns are the one part of it anything may key on,
and they are a contract in
[docs/decisions/0011-the-exit-codes.md](docs/decisions/0011-the-exit-codes.md).
The command interface around them is not a contract: a verb may be renamed and
the text a verb prints may be rewritten, and neither is announced anywhere except
here.

## v0.1.0 - 2026-09-04

The first release, so there is nothing to upgrade from and no behaviour that
changed. What this entry is for is saying what a downloaded file now is, since
until this tag the only way to run the runner was to build it out of the
repository it checks, which is a weaker position to check from.

**What is published.** A binary for each of the six platforms
[docs/decisions/0012-the-supported-platforms.md](docs/decisions/0012-the-supported-platforms.md)
fixes, with `NOTICE.md`, `LICENSE` and `privacy.md` beside them, plus the
third-party notices and the bill of materials rendered from the module table
inside a published binary. `SHA256SUMS` covers every one of those files and
`SHA256SUMS.sig` covers `SHA256SUMS`. The notes on the release carry the two
commands that check a download, and the signing keys they verify against are the
ones the platform publishes for the account rather than a key shipped in the
release.

**What the binary reports about itself.** `lab version` prints the version the
toolchain stamped from version control at build time rather than a constant
written into a source file, so the published binary reports the tag. Read off
the published asset rather than off a build made here, after checking its digest
against the line for it in `SHA256SUMS`:

    sha256sum --check --ignore-missing SHA256SUMS
    lab_v0.1.0_windows_amd64.exe: OK

    ./lab_v0.1.0_windows_amd64.exe version
    lab v0.1.0
    built from commit ed16621442c00a3ad47edd918d68edf68341753e, 2026-09-01T09:14:10Z

A build from an ordinary checkout reports a version derived from the commit
instead, a build from a tree carrying changes version control does not hold says
so on its own line, and the output explains which of the three it is handing you.

**Three things this release is not**, none of them new here. It is not something
anybody is asked to install, it is not advertised, and nothing on another board
depends on it, which is the scope
[CONTRIBUTING.md](CONTRIBUTING.md) opens with and which the release notes repeat
to whoever is holding the file rather than reading the repository. Who the
operator of an artefact is meant to be, and what publishing one costs, is
[docs/decisions/0021-what-this-board-publishes.md](docs/decisions/0021-what-this-board-publishes.md).

**What the archives do not contain.** There are no archives. The files travel as
separate assets beside the binaries, because an archive records a modification
time per entry and two runs from one tag would then produce two archives whose
bytes differ while the files inside them are identical, which would make the
reproducibility claim a claim about the archiver.

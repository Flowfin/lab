# What this is

A runner for reading and checking this board's own records. You point it at a
checkout of this repository and it reports what it examined and what it refused.

Every release published here carries this text, because it is assembled into the
notes by the release workflow rather than typed at the moment a release is cut.
A sentence that has to be remembered is a sentence that eventually is not.

## What this is not

Three exclusions, and they are the ones this board was opened with. The last of
them is the most important sentence here.

It is not something a media server user is asked to install. The artefacts exist
so that somebody checking this repository does not have to build the checker out
of the repository they are checking, which is a weaker position to verify from.
Whoever downloads one is an operator of this board and not a user of anything
this organisation ships.

It is not advertised anywhere. An experiment is allowed to fail, and something
that has been announced cannot fail quietly.

Nothing on any other board depends on it. Work that leaves here does so by
promotion, which is a hand-over recorded on both sides, and never as a
dependency edge pointing back at an experiment.

## What is in a release

A binary for each of the six platforms
[record 0012](https://github.com/Flowfin/lab/blob/main/docs/decisions/0012-the-supported-platforms.md)
fixes. Three of those six run the suite on every pull request and three are
compiled and not tested, which that record says out loud.

`THIRD-PARTY-NOTICES.md` and the bill of materials, both generated from the
module table inside a binary this release published rather than from a list
anybody maintains. Where this module carries no third-party dependency, the
notices say exactly that: an empty result that was produced is a different
statement from a file nobody created.

`NOTICE.md`, `LICENSE` and `privacy.md`, so that the terms the code arrives
under and what the runner does with what it reads travel with the file rather
than staying behind in a repository.

`SHA256SUMS`, covering every one of the files above, and `SHA256SUMS.sig`.

## Checking what you downloaded

The checksum file says the bytes arrived as they left. The signature says where
they left from, and that is the claim somebody verifying this board actually
needs, because anybody who can place a file can place a checksum beside it.

Verify the digest of a file you downloaded against the line for it:

```
sha256sum --check --ignore-missing SHA256SUMS
```

Verify the signature over that checksum file against the public signing keys the
platform publishes for the account that cut the release. The keys are not
shipped in the release: a key published beside the signature it verifies proves
nothing, because whoever placed one placed the other.

```
curl -s https://api.github.com/users/iderex/ssh_signing_keys \
  | sed -n 's/.*"key": "\(.*\)".*/iderex \1/p' > allowed-signers
ssh-keygen -Y verify -f allowed-signers -I iderex -n file \
  -s SHA256SUMS.sig < SHA256SUMS
```

Key custody is
[operations#1609](https://github.com/iderex/operations/issues/1609) and is not
restated here, because a custody story written twice is a custody story with two
answers.

What this does not tell you is what to do when a check fails. Record 0021
decides that the artefacts are signed and decides nothing about what a reader
does when the verification does not pass, and a published signature nobody
verifies is a file beside a file.

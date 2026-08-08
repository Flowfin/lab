# lab

Every other board in this organisation carries a promise, which is right for them and makes them a bad place to try something that will probably fail. Without somewhere to fail, experiments either do not happen or they happen inside a board that then has to explain them. The rule that keeps this from becoming a graveyard is that every experiment states its question before it starts and its answer when it stops, and the answer may be no. An experiment with no written answer is not finished, it is abandoned, and the difference should be visible. Everything here is public from the first commit, which means a failed experiment is public too.

Planning happens on the issue tracker first. Every decision that shapes
the architecture is argued there with its reasons before the code
that depends on it exists.

What survives the argument is written down in
[docs/decisions/](docs/decisions/). Each record is numbered, says what was
decided, what it applies to, what else was considered and what each rejected
option would have cost, and is replaced by a later record rather than edited.

Because everything here is public, some things may never be committed. No
credential, token or key, including expired ones. No personal data of any kind.
No copy of a real media library, real account data or real logs from a running
server. No file whose licence forbids it being here. The reasoning behind the
list is in
[docs/decisions/0006-everything-here-is-public.md](docs/decisions/0006-everything-here-is-public.md).

Nothing here uploads, phones home or reports usage, and real data an experiment
measures against stays on the machine it is already on.
[docs/privacy.md](docs/privacy.md) is the whole position, including what it does
not cover.

See [NOTICE.md](NOTICE.md) for the intended-use notice.

# lab

This is where questions that will probably fail get asked. Every other board in
this organisation carries a promise, which is right for them and makes them a
bad place to try something. Without somewhere to fail, an experiment either does
not happen or it happens inside a board that then has to explain it.

It takes questions and nothing else. Not software anybody is asked to install,
not work that is announced before it is done, and not anything another board is
allowed to depend on. Each of those is out for its own reason, and
[CONTRIBUTING.md](CONTRIBUTING.md) is where the three are argued.

## What an experiment is here

One directory, one question, and a record that says which. The question is
written and committed before the work starts, because a question written
afterwards is a question the result cannot have failed to fit. The record then
says how it ended.

No is an answer, and an experiment that answered no is finished rather than
failed. So is finding out that the question was the wrong question. What is not
finished is an experiment nobody wrote an answer for, and the record makes that
visible instead of leaving it to be guessed at. The three states a record can be
in, and what each one means, are in
[docs/decisions/0003-the-experiment-lifecycle.md](docs/decisions/0003-the-experiment-lifecycle.md).

Everything here is public from the first commit, a failed experiment included.

## Seeing what has been tried

From a checkout:

```
go run ./cmd/lab list
```

One line per experiment, oldest unanswered first, with how long each unanswered
question has been waiting. The output is not reproduced on this page on purpose:
a listing copied into a document is correct on the day it is copied and wrong on
the next experiment, and it is wrong in the direction that matters, which is a
record that exists in the tree and not in the page a visitor reads.

The same runner has a verb that judges the tree rather than describing it, and
[CONTRIBUTING.md](CONTRIBUTING.md) is where that one and the rest of the checks
are.

## Where the rest is

[docs/decisions/](docs/decisions/) holds every decision that shapes this board,
numbered, each with what was considered and what the rejected options would have
cost. A record is replaced by a later record rather than edited, so the
directory is a history of the argument rather than a description of today.

[NOTICE.md](NOTICE.md) is the intended-use notice.

[docs/privacy.md](docs/privacy.md) is the privacy position, including what it
does not cover.

## License

AGPL-3.0, copyright 2026 Nils Lehnen.

The full text is in [LICENSE](LICENSE).

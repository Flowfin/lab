# 0005. How a result leaves this board

## What was decided

Promotion is a hand-over rather than a transfer, and it is recorded on both
sides.

An experiment that answered yes and is wanted somewhere else gains a promotion
section in its record. The experiment's state does not change, because it is
still answered and record `0003` keeps the state saying what happened to the
question. What changes is that the record now says where the answer went.

The promotion section names four things and nothing about it is optional.

The destination repository. Where the work continued, written as the repository
rather than as a description of it.

The issue on the destination board that has agreed to pick the work up. Agreed,
not proposed. An issue that exists and says the work is wanted is the whole
evidence that this was a hand-over and not something thrown over a wall.

The commit range here that holds the result. What was handed over, at the
granularity somebody can actually check out, so that a reader on the other side
can see the thing as it was rather than as the tree became later.

The licence the code carries out. The terms the receiving board is taking it
under, stated at the moment of the hand-over rather than inferred afterwards
from whatever this repository's licence file said on some later day.

Promotion never happens by pushing code into another repository from here. A
product board takes work by its own rules, through its own issue and its own
review, and it is entitled to reject what it takes, rewrite it, or take the idea
and none of the code. This board's part ends at producing something worth taking
and a record that says what it is. Nothing in this repository ever writes to
another one, and no route that would need such a write is built.

The reason the destination issue is named rather than described is that a
description drifts and a link does not. A reader following a promotion record
should land on the place where the work continued, or on nothing at all, and
never on a sentence that used to be true. Landing on nothing is a readable
outcome, because it says the destination is gone; landing on prose that
disagrees with the destination is not, because there is no way to tell which of
the two is out of date.

What this cannot do. Nothing here checks that the destination issue exists, that
it agreed to anything, or that the commit range is the one the receiving board
took. A promotion section can name an issue nobody opened. The refusal built in
issue #39 catches a promotion section that points at nothing, which is the
structural half, and the rest is what a reader sees when they follow the link.

Whether a result may be promoted into a board whose terms differ from this one,
and who is entitled to place a contributor's work under those terms, is entry
three of issue #46 and is not decided here. This record fixes what the section
names, which is the same either way, and the answer to that entry decides who
may fill it in.

## What it applies to

Every experiment record under `experiments/` that has answered yes and whose
result is wanted elsewhere. The section is added by the change that hands the
work over.

It applies to the promotion checklist in issue #38, which is what somebody works
through before the section is written, and to the refusal in issue #39, which
reads the section afterwards.

It does not apply to an experiment that answered no. There is nothing to hand
over, and the record is finished as it stands. It also does not apply to
somebody copying an idea out of a record without any code moving, which needs no
section, because nothing left this board.

## What else was considered

Moving the code into the destination repository directly, by pushing a branch
there from here.

Recording the promotion only on the destination board, since that is where the
work continues.

Recording it only here, and leaving the destination board to decide whether it
wants to mention where the work came from.

Moving a promoted experiment to a state of its own, or out of `experiments/`
into a directory of things that graduated.

## What each rejected option would have cost

Pushing a branch into the destination repository costs that board its own
process. Work would arrive already written, outside its review, against
whatever conventions this board happens to have, and the receiving side would
have to either take it as it is or unpick it. It also needs a credential here
that can write to another repository, which is a permission this board has no
reason to hold and every reason not to, on a tree that is public and that
strangers are invited to read.

Recording it only on the destination board costs this side its own history. A
reader here would see an experiment that answered yes and stopped, with nothing
saying the result went anywhere, and would have to search the organisation to
find out. That reader is exactly the person the board exists for, since finding
out what came of an experiment is why anybody reads one.

Recording it only here costs the other half. A destination board carrying work
whose origin is not written down loses the ability to answer where a thing came
from, which is the question that gets asked when the work turns out to be wrong.
Naming the destination issue is what makes it recorded on both sides at once,
because the issue is on the other board and says the work was taken.

A separate state or a graduated directory costs the meaning of `answered` and
costs the walk. Record `0003` keeps the three states about the question rather
than about what happened to the code afterwards, and record `0002` fixes one
directory per experiment so a checker finds every record by walking one tree.
Moving a promoted experiment out of that tree means the records a reader most
wants to find are the ones that are no longer where the walk looks.

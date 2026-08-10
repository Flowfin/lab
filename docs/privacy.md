# Privacy

This board is public, its subject is software that sits next to people's media
libraries and accounts, and sooner or later something here will want to measure
against real data. This document says what that means for anybody running
anything from this repository, in plain language and before the first experiment
that touches anything real.

## The position

Personal data never leaves the host it is on unless the operator deliberately
federates it.

Nothing here uploads. Nothing phones home. Nothing reports usage, counts
installations, sends crash reports or ships a sample anywhere. There is no
telemetry to turn off, because there is none to turn on.

Where an experiment needs real data to answer its question, the data stays on the
operator's machine. Only the measurement is written down, and it is written in a
form that describes what was measured rather than carrying the thing that was
measured. A record can say that a library of a certain size took a certain time
to scan. It cannot carry the library, a sample of it, or the names of anything in
it.

The list of what may never be committed is in
[docs/decisions/0006-everything-here-is-public.md](decisions/0006-everything-here-is-public.md).
This document is the operator-facing half of the same position.

## What deliberately means

Deliberately is the load-bearing word above, so it is defined here rather than
left to be felt.

An operator federates deliberately when they take a specific action whose purpose
is to send data somewhere, knowing where it goes. Three properties, all of them
required. The action exists to send. The destination is named where the action is
taken. And doing nothing sends nothing.

None of the following is deliberate, and a future change can be measured against
this list:

- A default that sends. If data leaves on first run, the operator did not choose,
  they arrived.
- A setting that is on unless the operator finds it and turns it off. Opting out
  is not choosing.
- A prompt worded so that agreeing is the easy path, or where the agreeing button
  is the one styled to be pressed.
- Sending as a side effect of something else the operator asked for, where the
  sending is not what they asked for.
- A single agreement that covers destinations added later. Consent to one
  destination is not consent to the next one.

## What this does not cover

It says nothing about third-party services, which this board does not control. If
an experiment talks to a service somebody else runs, what that service does with
what it receives is between the operator and that service. A promise this board
cannot keep would be worse than the gap it names.

It says nothing about the operator's own machine. What else is running on the
host, what the host's own logs retain, and what a backup of it carries are all
outside anything written here.

It does not settle whether an experiment may use real data at all. This document
says what happens to real data if an experiment uses it. Whether that is
permitted in the first place is a wider question, open on issue #46, and issue
#35 is where the answer gets written down.

It is not a legal characterisation. This is a statement of what the tooling does
and does not do, not advice about what any particular law requires of an
operator. The intended-use notice is where that responsibility is placed.

## What stands behind this

The claim that the runner opens no network connection is tested, and the test
is in the default suite. Run it:

```
go test ./cmd/lab -count=1 -v -run 'TestTheRunnerLinksNoNetworkPackage|TestAFullRunOverTheLargestTreeCompletes'
```

Two legs, because each catches what the other misses. The first reads the
runner's whole transitive dependency set from the toolchain and refuses any
package that can open a socket, so a network client cannot arrive inside a
dependency nobody read. The second runs every verb the runner has over the two
largest trees in this repository and asserts each one reached a verdict, so a
route that is never exercised is not mistaken for one that is clean. The test
prints how many packages it read and how many files each tree held, so a reader
sees the size of what was covered rather than being told it was enough.

What the pair proves is that a full run executed in a binary that links nothing
able to open a connection. That is a structural argument rather than an
observation. Nothing here watches system calls, and a run under a tracer would
say more while saying it on one platform only, so this is not a claim that
something watched the process and saw no traffic. The bound is written at the
test as well, in `cmd/lab/network_test.go`, so it is not only here.

The other half of the same argument is what the repository depends on, which is
[docs/supply-chain.md](supply-chain.md). A tool with no dependencies has very
few places for a network call to hide, and the dependency leg above is what
holds that true rather than assumed.

The rule about real data staying on the host has no mechanism behind it and can
have none. Nothing in a checkout can tell a number somebody measured on their
own machine from a number they made up, and no check reads what an experiment
did before it wrote its record. Review is where a record carrying something it
should not is caught, and this is stated plainly here rather than left for a
reader to assume otherwise.

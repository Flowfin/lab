Slug: reading-a-tree-of-records
State: answered
Question-Written: 2026-08-11
Answer-Written: 2026-08-11
Needs-Hardware: none

## Question

On a tree of a thousand experiment records, does opening and reading every
record cost more wall-clock time than walking the directories that hold them?

## Method

`experiments/reading-a-tree-of-records/main.go` builds a thousand experiment
directories in a temporary directory, each holding a record of about a kilobyte,
which is the size most records in this repository are. It then times two passes
over that tree and removes it.

The first pass reads the directory, and for each experiment asks the filesystem
about the record without opening it. That is what the runner does before it
decides whether a record is inside the size bound. The second pass reads the
directory and opens every record. So the comparison is between asking about a
thousand files and reading a thousand files, rather than between walking and
doing nothing, and the difference between the two is the cost of the bytes.

Both passes run seven times in one process against one tree, and the number
reported is the fastest round rather than the mean. The fastest is the one least
contaminated by whatever else the machine was doing, and a mean over a busy
machine measures the other things.

Run it with:

    go run ./experiments/reading-a-tree-of-records

The measurement is about the machine it ran on. The platform, the processor
architecture and the toolchain version are printed next to the numbers, and no
number here is a claim about anybody else's machine.

## Answer

Yes, and by about three times. Reading the records is where the time goes.

    go run ./experiments/reading-a-tree-of-records
    windows/amd64, go1.26.5
    1000 experiments, 1108000 bytes of records, 7 rounds
    walking the directories:        16.6888ms
    walking and reading every file: 50.5875ms
    reading costs 3.03 times the walk

Two more runs of the same command on the same machine gave 3.04 and 2.93, so
the ratio is stable to about a tenth and the absolute numbers move by a few
milliseconds between runs. Those two runs are not pasted, and their numbers are
quoted from the same command as the one above.

What it does not say. This is one machine, one filesystem and one platform, and
the cost of asking a filesystem about a file rather than opening it is exactly
the sort of thing that differs between them. A tree of a thousand records is
also fifty times the size of this repository's own tree today, so the whole
measurement is milliseconds either way and none of it is a reason to change
anything yet. What it settles is which of the two numbers is worth watching if a
walk ever does get slow: the bytes, not the directories.

The question was a fair one to have asked, and it could have gone the other way.
A directory walk that has to open each directory in turn is not obviously
cheaper than reading a kilobyte out of a file the walk has already found, and on
a filesystem with a slower directory layer it may not be.

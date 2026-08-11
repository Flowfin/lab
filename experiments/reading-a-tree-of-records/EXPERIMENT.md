Slug: reading-a-tree-of-records
State: asking
Question-Written: 2026-08-11
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

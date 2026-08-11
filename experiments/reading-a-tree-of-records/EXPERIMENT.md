Slug: reading-a-tree-of-records
State: asking
Question-Written: 2026-08-11
Needs-Hardware: none

## Question

On a tree of a thousand experiment records, does opening and reading every
record cost more wall-clock time than walking the directories that hold them?

## Method

Not run yet. The plan is a program that builds a tree of a thousand experiment
directories in a temporary directory, each holding a record of about the size
the records in this repository are, and then times two passes over it: one that
only walks the directories and reads their names, and one that also opens every
record and reads its bytes. Both passes run several times in one process, the
tree is built once and reused, and the numbers reported are the fastest of the
runs rather than the mean, because the fastest is the one least contaminated by
whatever else the machine was doing.

The measurement is about the machine it runs on. The platform, the processor
architecture and the toolchain version are reported next to the numbers, and no
number here is a claim about anybody else's machine.

## Answer

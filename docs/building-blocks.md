# Building Blocks

This document introduces our use of Golang Interfaces and Structs
in `TTPForge` to create the concepts of `Acts` and `Steps`, which
form the building blocks of Tactics, Techniques, and Procedures (TTPs).

## Act

`Acts` contain the fields and functionality that are common across
different types of `Steps`.

This interface facilitates building TTPs with steps that don't need to
explicitly understand what the others are doing to function. Each step
has its own enforcement mechanisms and can function without needing to
understand how the previous step worked, simply providing information
for the current step to use.

This design enables two different step types to offer precise control of
process lineage and corresponding indicators of compromise.

## Step

`Steps` contain a high-level interface that can be used to define
instructions or `Steps` that make up a TTP. The following `Steps` exist today:

- `BasicStep`: Simulates interactive interpreter abuse, which can
  be used to simulate hands-on-keyboard TTPs.
- `FileStep`: Simulates direct process execution without an
  intermediate interpreter, which can be used to simulate an attacker
  executing logic from a file.
- `SubTTPStep`: A Collection of `TTPs` that are used together to represent a `TTP`.

## TTP

`TTPs` are collections of `Steps` that make up the logic to represent
a component of a Tactic, Technique, and Procedure.

## Cleanup

`Cleanup` is an embedded structure for all Acts to make use of in order
to clean up anything that a TTP has done.

## Resources

- [LogRocket: Exploring Structs and Interfaces in Go](https://blog.logrocket.com/exploring-structs-interfaces-go/)
- [Go by Example: Interfaces](https://gobyexample.com/interfaces)
- [A Tour of Go: Interfaces](https://go.dev/tour/methods/9)
- [Introduction to Tokenizers and Compilers](https://www.cs.man.ac.uk/~pjj/farrell/comp3.html)

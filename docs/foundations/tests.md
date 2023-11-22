# Tests for TTPs

You can write tests for your TTPs using the `tests:` section of a TTPForge YAML
file. These tests serve two purposes:

- They act as continuously-validated documentation for how users should run your
  TTP.
- They help ensure that the TTPForge engine will remain compatible with your TTP
  and provide warning if this compatibility is broken for any reason.

## Basic Test Cases

The simplest-possible test case that you can define for a TTP is shown below:

https://github.com/facebookincubator/TTPForge/blob/bf2fbb3312a227323d1930ba500b76f041329ca2/example-ttps/tests/minimal_test_case.yaml#L1-L14

When you run the test cases for this TTP via the command
`ttpforge test examples//tests/minimal_test_case.yaml`, TTPForge will call
`ttpforge run` and pass the absolute path to your TTP file as an argument. In
this instance, the `tests` syntax may seem superfluous, but even in this simple
case it plays a very important role: **by declaring a test case, you are telling
TTPForge that your TTP is safe to run as an automated test.**

## Test Cases with Arguments

The `tests` feature really starts to show its value when used for TTPs that
expect command-line arguments. An example of such a TTP, with two associated
test cases, is shown below:

https://github.com/facebookincubator/TTPForge/blob/bf2fbb3312a227323d1930ba500b76f041329ca2/example-ttps/tests/with_args.yaml#L1-L42

When you test this TTP via `ttpforge test examples//tests/with_args.yaml`, both
of the test cases in the above file will be run sequentially. TTPForge will
parse the provided `args` list, encode each entry in the string format
`--arg foo=bar`, and then append each resulting string to a dynamically
generated `ttpforge run` command. The subsequent execution of that command
verifies that the TTP functions correctly for that test case.

## Dry-Run Test Cases

Some TTPs can only be executed except under very specific conditions - for
example, Active Directory exploits that target domain controllers. It may not be
feasible to test execution of such a TTP in an automated setting; however, it is
still possible to verify that the TTP parses its arguments correctly and that
all TTPForge validation phases _prior to actual execution_ complete
successfully. To perform "validation without execution" in this manner, add
`dry_run: true` to your test case, as shown below:

https://github.com/facebookincubator/TTPForge/blob/bf2fbb3312a227323d1930ba500b76f041329ca2/example-ttps/tests/dry_run.yaml#L1-L30

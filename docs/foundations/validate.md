# Validating TTPs

## Validating TTP Structure and Syntax

To validate the structure and syntax of a TTP YAML file, use the command shown
below:

```bash
ttpforge validate [repo_name//path/to/ttp]
```

Unlike `--dry-run`, this command does not require values for all arguments or
OS/platform compatibility. This makes it ideal for CI/CD validation and linting.

The output reports errors, warnings, and informational messages about the TTP's
structure, metadata, arguments, and steps.

## Running Tests After Validation

You can optionally run the TTP's test cases after validation:

```bash
ttpforge validate [repo_name//path/to/ttp] --run-tests
```

Please note:

1. Tests are only executed if validation passes with no errors.
2. You can specify a timeout for each test case using `--timeout-seconds`
   (default is 10 seconds).

For more information about TTP tests, see [Tests for TTPs](tests.md).

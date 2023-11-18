# TTPForge Actions: `print_str`

The `print_str` action will print the specified string to the screen. Think of
it like using the `echo` shell command, but with the following improvements:

- It's easier to print large, "messy" strings containing metacharacters.
- It won't spawn a loud shell process and create erroneous telemetry for
  defenders.

## Fields

You can specify the following YAML fields for the `print_str` action:

- `print_str:` (type: `string`) the string to print.

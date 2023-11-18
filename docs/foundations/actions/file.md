# TTPForge Actions: `file`

The `file` action will execute the specified executable file using Golang's
`os/exec` package. Think of it like executing the file in a shell, but more
stealthy - this activity should be visible in process execution telemetry but
not in shell history. It's similar to using `execute -o program` in
[Sliver C2](https://github.com/BishopFox/sliver).

## Fields

You can specify the following fields for the `file:` action:

- `file:` (type: `string`) the path to the file to execute.
- `args:` (type: `list`) list of strings to pass as arguments to the invoked
  program.

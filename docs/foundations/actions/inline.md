# TTPForge Actions: `inline`

By default, the `inline` action runs the provided shell command in a new
instance of the `bash` shell:

https://github.com/facebookincubator/TTPForge/blob/0d62cf5139cb97686f4a6ef76fdf2bf7a30681be/example-ttps/actions/inline/basic.yaml#L1-L21

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/inline/basic.yaml
```

## Fields

You can specify the following YAML fields for the `inline:` action:

- `inline:` (type: `string`) the command that you want to run.
- `executor:` (type: `string`) the program that should run your command. The
  program you specify will be launched and your command will be sent to its
  STDIN. Default: `bash`.

## Notes

Key things to remember about `inline:` actions:

- By default, `inline` passes `-o errexit` to the `bash` executor, meaning that
  the failure of any single command will terminate the step and start cleanup.
  This prevents silent failures and makes TTPs more reliable.
- Each separate `inline` action instance runs in its own shell. Sharing shell
  variables between different `inline` steps is not supported yet.

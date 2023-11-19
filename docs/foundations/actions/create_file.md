# TTPForge Actions: `create_file`

The `create_file` action can be used to drop files on disk without the need to
loudly invoke a shell and use `cat` or `echo`. Check out the TTP below to see
how it works:

https://github.com/facebookincubator/TTPForge/blob/0d62cf5139cb97686f4a6ef76fdf2bf7a30681be/example-ttps/actions/create_file/basic.yaml#L1-L22

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/create_file/basic.yaml
```

## Fields

You can specify the following YAML fields for the `create_file:` action:

- `create_file:` (type: `string`) the path to the file you want to create.
- `contents:` (type: `string`) the contents that you want placed in the new
  file.
- `overwrite:` (type: `bool`) whether the file should be overwritten if it
  already exists.
- `mode:` the octal permission mode (`chmod` style) for the new file.

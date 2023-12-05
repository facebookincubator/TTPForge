# TTPForge Actions: `create_file`

The `create_file` action can be used to drop files on disk without the need to
loudly invoke a shell and use `cat` or `echo`. Check out the TTP below to see
how it works:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/actions/create-file/basic.yaml#L1-L22

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/create-file/basic.yaml
```

## Fields

You can specify the following YAML fields for the `create_file:` action:

- `create_file:` (type: `string`) the path to the file you want to create.
- `contents:` (type: `string`) the contents that you want placed in the new
  file.
- `overwrite:` (type: `bool`) whether the file should be overwritten if it
  already exists.
- `mode:` the octal permission mode (`chmod` style) for the new file.
- `cleanup:` you can set this to `default` in order to automatically cleanup the
  created file, or define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics).

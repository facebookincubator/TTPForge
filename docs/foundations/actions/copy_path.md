# TTPForge Actions: `copy_path`

The `copy_path` action can be used to copy files on disk without the need to
loudly invoke a shell and use `cat`, `echo`, or `cp`. Check out the TTP below to
see how it works:

```yaml
---
name: copy_path_example
description: |
  This TTP shows you how to use the copy_path action type
  to copy a file on disk.
requirements:
  platforms:
    - os: darwin
    - os: linux
steps:
  - name: copy-passwd-file
    copy_path: /etc/passwd
    to: /tmp/ttpforge_copy_{{randAlphaNum 10}}
    mode: 0600
    cleanup: default
```

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/copy-path/basic.yaml
```

## Fields

You can specify the following YAML fields for the `copy_path:` action:

- `copy_path:` (type: `string`) the path to the file or directory you want to
  copy.
- `to:` (type: `string`) the path to the file or directory you want to write the
  copy to file.
- `recursive:` (type: `bool`) whether or not the copy action should be recursive
  (copy all files in directory)
- `overwrite:` (type: `bool`) whether the file(s) should be overwritten if they
  already exist in the destination.
- `mode:` the octal permission mode (`chmod` style) for the new file.
- `cleanup:` you can set this to `default` in order to automatically cleanup the
  created file, or define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics).

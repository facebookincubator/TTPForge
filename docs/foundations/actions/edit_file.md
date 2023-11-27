# TTPForge Actions: `edit_file`

The `edit_file` action is useful for automating malicious modifications to files
(for example, adding yourself to `/etc/sudoers` or commenting out important
logging code). `edit_file` can append, delete, or replace lines in the target
file - check out the examples below to learn more.

## Appending and Deleting Lines

This example shows how to use the `append` and `delete` functionality of the
`edit_file` action:

https://github.com/facebookincubator/TTPForge/blob/bf2fbb3312a227323d1930ba500b76f041329ca2/example-ttps/actions/edit_file/append_delete.yaml#L1-L35

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/edit_file/append_delete.yaml
```

## Replacing Lines

You can also use `edit_file` to replace lines in a file and optionally use
powerful regular expressions to perform complex transformations. The next
example shows this functionality in action:

https://github.com/facebookincubator/TTPForge/blob/bf2fbb3312a227323d1930ba500b76f041329ca2/example-ttps/actions/edit_file/replace.yaml#L1-L47

Try out the above TTP by running this command:

```bash
ttpforge run examples//actions/edit_file/replace.yaml
```

## Fields

You can specify the following YAML fields for the `edit_file` action:

- `edit_file:` (type: `string`) the path to the file you want to edit (must
  exist).
- `backup_file:` (type: `string`) the backup path to which the original file
  should be copied.
- `edits:` (type: `list`) a list of edits to make. Each entry can contain the
  following fields:
  - `delete:` (type: `string`) string/pattern to delete - pair with
    `regexp: true` to treat as a Golang
    [regular expression](https://pkg.go.dev/regexp/syntax) and delete all
    matches thereof.
  - `append:` (type `string`) line(s) to append to the end of the file.
  - `old:` (type: `string`) string/pattern to replace - pair with `regexp: true`
    to treat as a Golang [regular expression](https://pkg.go.dev/regexp/syntax)
    and replace all matches thereof. Must always be paired with `new:`
  - `new:` (type: `string`) string with which to replace the string/pattern
    specified by `old:` - must always be paired with `old:`
- `cleanup:` you can set this to `default` in order to automatically restore the
  original file once the TTP completes. **Note**: this only works when
  `backup_file` is set. You can also define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics).

## Notes

- `edit_file` will read the entire file into memory, perform all specified
  edits, and then write out the results. Be careful when using it against very
  large files.
- `edit_file` does not support editing binary files.
- The `edits` list is looped through from top to bottom and all edits are
  applied sequentially to the copy of the file contents residing in memory. This
  means, for example, that if you `append` and then later `delete` that same
  line, the resulting final file won't contain that line.

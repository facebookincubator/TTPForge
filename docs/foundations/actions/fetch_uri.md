# TTPForge Actions: `fetch_uri`

The `fetch_uri` action can be used to download files via http to disk without
the need to invoke a shell and use `wget` or `curl`. Check out the TTP below to
see how it works:

https://github.com/facebookincubator/TTPForge/blob/a8cd35133e4100ed7b50ee14d51da78e19df9786/example-ttps/actions/fetchuri/basic.yaml#L1-L11

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/fetchuri/basic.yaml
```

## Fields

You can specify the following YAML fields for the `fetch_uri:` action:

- `fetch_uri:` (type: `string`) the uri to the file you wish to download.
- `location:` (type: `string`) the path to save the file on disk.
- `proxy:` (type: `string`) the http proxy url to use for the request.
- `overwrite:` (type: `bool`) whether the file should be overwritten if it
  already exists.
- `cleanup:` you can set this to `default` in order to automatically cleanup the
  created file, or define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics).

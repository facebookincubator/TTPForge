# TTPForge Actions: `copy_path`

The `copy_path` action copies files or directories on disk without invoking a
shell — no `cp`, `cat`, or `echo` appears in shell history. This simulates
C2-style file operations.

## Fields

- `copy_path:` (type: `string`) the source path to copy from.
- `to:` (type: `string`) the destination path to copy to.
- `recursive:` (type: `bool`) set to `true` to copy directories and their
  contents. Required when the source is a directory.
- `overwrite:` (type: `bool`) whether existing destination files should be
  overwritten. Defaults to `false`.
- `mode:` (type: `int`) octal permission mode (`chmod` style) for copied files.
  Defaults to `0666`.
- `direction:` (type: `string`) controls which filesystem the source and
  destination refer to when used with a `remote:` block. See
  [Remote File Transfer](#remote-file-transfer) below.
- `cleanup:` set to `default` to automatically remove the destination on
  cleanup, or define a custom
  [cleanup action](../cleanup.md#cleanup-basics).

## Basic Usage

```yaml
steps:
  - name: copy_config
    copy_path: /etc/app/config.yaml
    to: /tmp/config_backup.yaml
```

Copy a directory recursively:

```yaml
steps:
  - name: copy_logs
    copy_path: /var/log/app
    to: /tmp/app_logs
    recursive: true
    cleanup: default
```

## Remote File Transfer

When used with a [`remote:` block](../remote.md), `copy_path` operates on the
remote filesystem by default. To transfer files *between* the local machine and
a remote host, use the `direction` field:

- `direction: upload` — reads from the local filesystem, writes to the remote
  host
- `direction: download` — reads from the remote filesystem, writes to the local
  machine
- omitted — copies within the same filesystem (local or remote depending on
  whether `remote:` is set)

The `direction` field requires a `remote:` block on the step. If you need a
local-only copy, simply omit `remote:` from the step.

To transfer directories, set `recursive: true`. This works with all directions.

### Upload a local file to a remote host

```yaml
steps:
  - name: setup-connection
    connect:
      protocol: ssh
      host: "{{ .Args.target_host }}"
      auth: agent
      connection_name: target

  - name: upload_payload
    remote: target
    copy_path: /tmp/local_payload.bin
    to: /opt/payload.bin
    direction: upload
```

### Download a remote file to the local machine

```yaml
steps:
  - name: download_loot
    remote: target
    copy_path: /etc/shadow
    to: /tmp/loot/shadow
    direction: download
    cleanup: default
```

When `direction: download` is used with `cleanup: default`, the cleanup action
removes the destination from the **local** filesystem (not the remote host).

### Upload a directory

```yaml
steps:
  - name: upload_tools
    remote: target
    copy_path: /opt/tools
    to: /tmp/tools
    direction: upload
    recursive: true
    cleanup: default
```

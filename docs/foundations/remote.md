# Remote Execution over SSH

TTPForge supports executing steps on remote hosts over SSH. You first establish
a named connection with a `connect` step, then reference that connection by name
with `remote:` on subsequent steps. Steps without `remote:` continue to run on
the local machine, allowing you to freely mix local and remote actions within a
single TTP.

## Basic Usage

Use a `connect` step to establish a named SSH connection, then reference it with
`remote: <connection_name>` on any step:

```yaml
steps:
  - name: setup-connection
    connect:
      host: my-server.example.com
      user: root
      auth: agent
      connection_name: my-server

  - name: enumerate_remote
    remote: my-server
    inline: whoami && hostname

  # This step runs locally (no remote: field)
  - name: local_step
    inline: echo "running on local machine"
```

## The `connect` Step

The `connect` step establishes a named SSH connection that can be reused across
multiple steps. It eagerly connects at execution time so that authentication
failures surface immediately rather than on the first step that uses the
connection.

### Fields

The `connect` block supports the following fields:

- `host:` (type: `string`, **required**) the hostname or IP address of the
  remote target.
- `connection_name:` (type: `string`, **required**) the name used to reference
  this connection from `remote:` fields on subsequent steps.
- `protocol:` (type: `string`, default: `ssh`) the protocol to use. Currently
  only `ssh` is supported.
- `port:` (type: `int`, default: `22`) the SSH port on the remote host.
- `user:` (type: `string`, default: current user) the SSH username.
- `auth:` (type: `string`, default: `agent`) the authentication method. See
  [Authentication Methods](#authentication-methods) below.
- `key_file:` (type: `string`) path to the SSH private key file. Required when
  `auth: key`.
- `password:` (type: `string`) the SSH password. Required when `auth: password`.
  Use with `{{ .Args.x }}` to avoid hardcoding passwords in YAML files.
- `password_env:` (type: `string`) name of the environment variable containing
  the SSH password. Required when `auth: password_env`.
- `known_hosts:` (type: `string`) path to a known_hosts file for host key
  verification. If omitted, host key checking is disabled.
- `jump_host:` (type: `string`) a bastion/jump host to connect through, in
  `host:port` format.
- `shell:` (type: `string`, default: `posix`) the shell type on the remote
  host. Controls how environment variables, directory changes, and argument
  quoting are constructed. Supported values:
  - `posix` — POSIX-compatible shells (bash, sh, zsh). Uses `export`,
    single-quote escaping, and `&&` chaining.
  - `powershell` — Windows PowerShell. Uses `$env:`, backtick escaping, and
    `;` chaining. Process management uses `Stop-Process`/`Get-Process`.
  - `cmd` — Windows cmd.exe. Uses `set`, double-quote escaping, and `&`
    chaining. Process management uses `taskkill`/`tasklist`.

All string fields support `{{ .Args.x }}` template variables, so connection
parameters can be passed as TTP arguments.

## Referencing Connections with `remote:`

Once a connection is established with a `connect` step, reference it by name
on any subsequent step:

```yaml
steps:
  - name: setup-connection
    connect:
      host: "{{ .Args.target_host }}"
      user: "{{ .Args.target_user }}"
      auth: key
      key_file: "{{ .Args.key_file }}"
      connection_name: target

  - name: step_one
    remote: target
    inline: whoami

  - name: step_two
    remote: target
    inline: hostname

  - name: drop_file
    remote: target
    create_file: /tmp/payload.sh
    contents: |
      #!/bin/bash
      echo "hello from $(hostname)"
    mode: 0755
    cleanup: default
```

## Multiple Connections

You can establish multiple named connections to different hosts:

```yaml
steps:
  - name: connect-web
    connect:
      host: web-server.example.com
      user: deploy
      auth: key
      key_file: ~/.ssh/deploy_key
      connection_name: web

  - name: connect-db
    connect:
      host: db-server.example.com
      user: admin
      auth: agent
      connection_name: db

  - name: check-web
    remote: web
    inline: curl -s localhost:8080/health

  - name: check-db
    remote: db
    inline: psql -c "SELECT 1"
```

## Authentication Methods

### `agent` (default)

Uses the SSH agent (via `SSH_AUTH_SOCK`) for authentication. The agent must be
running and have the appropriate key loaded:

```yaml
connect:
  host: target.example.com
  auth: agent
  connection_name: target
```

```bash
# Ensure your key is in the agent
ssh-add /path/to/your/key
```

### `key`

Reads a private key directly from disk. If a corresponding certificate file
exists (following the OpenSSH convention of `<key_file>-cert.pub`), it is
automatically loaded for certificate-based authentication:

```yaml
connect:
  host: target.example.com
  auth: key
  key_file: /path/to/id_rsa
  connection_name: target
```

### `password`

Uses a password provided directly in the YAML. Combine with `{{ .Args.x }}`
to accept the password as a TTP argument:

```yaml
args:
  - name: ssh_password
    type: string

steps:
  - name: setup
    connect:
      host: target.example.com
      auth: password
      password: "{{ .Args.ssh_password }}"
      connection_name: target

  - name: run
    remote: target
    inline: whoami
```

```bash
ttpforge run my-ttp.yaml --arg ssh_password=secret
```

### `password_env`

Reads the password from an environment variable. This avoids hardcoding
passwords in YAML files:

```yaml
connect:
  host: target.example.com
  auth: password_env
  password_env: SSH_PASSWORD
  connection_name: target
```

```bash
export SSH_PASSWORD="my-secret-password"
ttpforge run my-ttp.yaml
```

## Connection Pooling

TTPForge automatically reuses SSH connections across steps that reference the
same named connection. Connections are created when the `connect` step executes
and closed when the TTP finishes. If two `connect` steps point to the same
host/port/user combination, they share the underlying SSH connection.

## Supported Actions

The following action types support remote execution:

- `inline:` — shell commands run on the remote host
- `file:` — scripts/binaries execute on the remote host
- `create_file:` — files are created on the remote filesystem
- `remove_path:` — files/directories are removed from the remote filesystem
- `copy_path:` — files are copied on the remote filesystem (or transferred
  between local and remote with `direction:`)
- `edit_file:` — files are edited on the remote filesystem
- `fetch_uri:` — fetched content is written to the remote filesystem
- `change_directory:` — working directory is changed on the remote filesystem
- `kill_process:` — processes are killed on the remote host

Output from remote `inline:` and `file:` steps is streamed line-by-line in
real time, matching the behavior of local execution.

## How Remote Actions Work

Remote actions use one of two mechanisms:

- **`inline:`, `file:`, `kill_process:`** execute shell commands via SSH
  sessions, producing the same process telemetry as running the command
  directly on the host.
- **All other actions** (`create_file`, `remove_path`, `copy_path`, `edit_file`,
  `fetch_uri`, `change_directory`) operate through SFTP — no shell process is
  spawned and no shell history is written on the target. This simulates
  C2-style file operations.

## Unsupported Actions

The following action types do **not** support remote execution and will return
an error if used with a `remote:` block:

- `expect:` — requires a local PTY for interactive console automation, which
  cannot be forwarded over SSH.
- `http_request:` — makes HTTP calls from the local machine. To make HTTP
  requests from the remote host, use `inline:` with `curl` or similar.

## Checks

By default, success checks defined on a remote step inherit the step's
`remote:` field — file existence checks use the remote filesystem, and command
checks execute on the remote host.

Each check can optionally specify its own `remote:` field to override this
behavior:

| Check YAML | Behavior |
|---|---|
| No `remote:` field | Inherits the step's `remote:` (backwards compatible) |
| `remote: <connection_name>` | Runs against that named connection |
| `remote: local` | Forces local execution even if the step is remote |

```yaml
steps:
  - name: deploy_payload
    remote: target
    inline: cp payload.bin /opt/payload.bin
    checks:
      - msg: Payload exists on remote
        path_exists: /opt/payload.bin
        # no remote: → inherits step's "target"

      - msg: Local log updated
        remote: local
        path_exists: /var/log/deploy.log

      - msg: C2 callback received
        remote: c2_server
        command: grep -q "callback" /var/log/c2.log
```

## Cleanup

Cleanup actions support independent remote targeting:

- **`cleanup: default`** inherits the step's `remote:` field, so cleanup runs on
  the same host as the step. For example, `cleanup: default` on a remote
  `create_file` step removes the file from the remote filesystem.
- **Custom cleanup** (a YAML mapping) runs **locally by default**. Add
  `remote: <connection_name>` inside the cleanup block to run it on a specific
  remote host.

```yaml
steps:
  - name: drop-payload
    remote: target
    create_file: /tmp/payload.sh
    contents: echo "hello"
    cleanup: default              # inherits remote: target

  - name: run-payload
    remote: target
    inline: /tmp/payload.sh
    cleanup:
      remote: target              # explicit remote for cleanup
      inline: rm -f /tmp/payload.sh

  - name: exfil
    remote: target
    inline: cat /etc/passwd
    cleanup:
      inline: rm -f /tmp/local-evidence.log   # no remote: → runs locally
```

## Cross-Platform Support

By default, TTPForge assumes the remote host uses a POSIX-compatible shell. To
target Windows hosts, set the `shell:` field in the `connect` step:

```yaml
steps:
  - name: connect-windows
    connect:
      host: windows-server.example.com
      user: Administrator
      auth: key
      key_file: ~/.ssh/windows_key
      shell: powershell
      connection_name: winbox

  - name: enumerate_windows
    remote: winbox
    inline: Get-ComputerInfo | Select-Object CsName, OsName
```

The `shell` setting affects:

- **Environment variables**: `export K='V'` vs `$env:K = "V"` vs `set K=V`
- **Directory changes**: `cd 'path'` vs `Set-Location "path"` vs `cd /d "path"`
- **Argument quoting**: single-quote vs backtick-escaped double-quote vs
  double-quote
- **Command chaining**: `&&` vs `;` vs `&`
- **Process management**: `kill`/`pgrep` vs `Stop-Process`/`Get-Process` vs
  `taskkill`/`tasklist`
- **Verification checks**: `sh -c` vs `powershell -Command` vs `cmd /c`

## Example

See the full example TTP at
[`example-ttps/actions/inline/remote-ssh.yaml`](../../../example-ttps/actions/inline/remote-ssh.yaml):

```bash
ttpforge run examples//actions/inline/remote-ssh.yaml \
  --arg target_host=my-server.example.com \
  --arg target_user=root \
  --arg target_key_file=/path/to/id_rsa
```

# TTPForge Actions: `kill_process`

The `kill_process` action can be used to terminate processes running on a system.
Check out the TTP below to see how it works:

[Kill Process Unix](https://github.com/facebookincubator/TTPForge/blob/0deabc567751d90078e5db3c2a84574396b43dc1/example-ttps/actions/kill-process/kill-process-unix.yaml)
[Kill Process Windows](https://github.com/facebookincubator/TTPForge/blob/0deabc567751d90078e5db3c2a84574396b43dc1/example-ttps/actions/kill-process/kill-process-windows.yaml)

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/kill-process/kill-process-[unix/windows].yaml
```

## Fields

You can specify the following YAML fields for the `kill_process:` action:

- `id:` (type: `string`) the process ID of the process that you wish to kill
- `name:` (type: `string`) the process name of the process that you wish to kill
- `error_on_find_process_failure:` (type: `bool`) whether to raise an error if
finding the process name/id fails
- `error_on_kill_failure:` (type: `bool`) whether to raise an error if killing
 the process fails

**Note:** You must specify exactly one of `id` or `name`. You cannot specify both,
and at least one is required.

## Example Usage

Kill a process by name:

```yaml
- name: Kill ping process
  kill_process:
    name: "ping"
  error_on_find_process_failure: true
```

Kill a process by ID:

```yaml
- name: Kill process by ID
  kill_process:
    id: "1234"
```

## Additional Notes

Both the flags `error_on_find_process_failure` and `error_on_kill_failure` are
set to `false` by default.

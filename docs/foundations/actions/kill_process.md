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

- `kill_process_id:` (type: `string`) the process ID of the process that
you wish to kill
- `kill_process_name:` (type: `string`) the process name of the process that
you wish to kill
- `error_on_find_process_failure:` (type: `bool`) whether to raise an error if
finding the process name/id fails
- `error_on_kill_failure:` (type: `bool`) whether to raise an error if killing
 the process fails

## Additional Notes

If both `kill_process_id` and `kill_process_name` are specified, the action
will only consider process ID as long as it is valid.
If an invalid `kill_process_id` is specified, the action will fall back to
using `kill_process_name` to kill the processes.
Both the flags `error_on_find_process_failure` and `error_on_kill_failure` are
set to `false` by default.

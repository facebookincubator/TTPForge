# Reliable Post-Execution Cleanup of TTP Actions

**Note**: to run the examples in this section, make sure you have the `examples`
repository installed with `ttpforge list repos` and if not run:

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

## Goals

TTPs will often include destructive (or at the very least, messy) actions, such
as:

- Editing System Files (such as `/etc/sudoers` or `/root/.ssh/authorized_keys`)
- Killing/Disabling Critical System Services (especially endpoint security
  solutions such as MDE)
- Launching Cloud Resources (EC2 Instances/Docker Containers/Kubernetes Pods)

Failure to clean up properly after such activity will at the very least
inconvenience the user, and at worst may actually create severe new security
vulnerabilities in the target system.

Interpreter scripts usually lack standardized, platform-independent, and
reliable methods for cleaning up attacker activity. In particular, it is
especially difficult to ensure that these clean up processes are resilient to
runtime errors. For example, when writing TTPs in bash, one would need to make
very careful use of the
[trap](https://tldp.org/LDP/Bash-Beginners-Guide/html/sect_12_02.html) feature.

TTPForge's native support for cleanup actions provides users with a reliable
solution to these problems.

## Cleanup Basics

Every step in a TTPForge TTP can be associated with a cleanup action. That
cleanup action can be any valid TTPForge [action type](actions.md). Here's an
example TTP with several steps that have associated cleanup actions:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/cleanup/basic.yaml#L1-L25

You can run this example with the command:
`ttpforge run examples//cleanup/basic.yaml` - based on the output, notice the
following key aspects about how cleanup actions work:

1. Every time a step completes successfully, its cleanup action is added to the
   cleanup queue.
1. Not every step defines a cleanup action, and that's just fine. No cleanup
   action will be added to the queue upon completion of those steps.
1. Once the TTP completes (or a step fails), the actions in the cleanup queue
   are executed in reverse (last-in, first-out) order.

Pay particular attention to point (3) - TTPForge executes cleanup actions in
reverse order because it is quite common for later steps to depend on earlier
ones. For instance, one would definitely want the steps of the following
scenario cleaned up in reverse order:

1. Attacker adds self to `/etc/sudoers`
1. Attacker loads a malicious kernel module for persistence.

If we cleaned up step (1) first, we would then lose the privileges required to
cleanup (2).

## Delaying or Disabling Cleanup

Sometimes, one may wish to execute a given TTP and then to leave the target
system in a "dirty" state for further analysis. For these purposes, TTPForge
provides two useful command line flags for the `ttpforge run` command:

- `--cleanup-delay-seconds` - delay cleanup execution for the specified integer
  number of seconds
- `--no-cleanup` - do not run any cleanup actions; instead, simply exit when the
  last step completes.

## Default Cleanup Actions

Certain action types (such as [create_file](actions/create_file.md) and
[edit_file](actions/edit_file.md)) have a default cleanup action that can be
invoked by specifying `cleanup: default` in their YAML configuration. In the
case of `create_file`, the default cleanup action removes the created file.
Check out the example below, which you can run with
`ttpforge run examples//cleanup/default.yaml`:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/cleanup/default.yaml#L1-L12

## Handling Failures Gracefully

Whenever a step fails, the cleanup process will begin from the last successful
step (the step immediately preceding the failure). The example below (which you
can run with `ttpforge run examples//cleanup/failure.yaml`) shows how cleanup
actions gracefully handle the failure of a step:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/cleanup/failure.yaml#L1-L27

Note that **we don't clean up the failed step itself**, because that is usually
not desired behavior. Consider the following example situations:

- The step failed to create file due to a permissions issue. The cleanup action
  to delete the file would also fail because the file was never created in the
  first place.
- The step failed to setup a cloud resource due to a capacity constraint. The
  cleanup action to remove the resource would also fail because no resource was
  ever provisioned in the first place.

If a cleanup action fails, all remaining cleanup actions in the queue are
abandoned and not run. In this situation, it's likely that something is
fundamentally wrong with the TTP/test system and we want to prompt the user to
investigate rather than pushing forward and perhaps deleting something that we
shouldn't.

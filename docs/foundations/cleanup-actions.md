# Reliable Post-Execution Cleanup

## Goals 

TTPs will often include destructive (or at the very least, messy) actions, such as:

* Editing System Files (such as `/etc/sudoers` or `/root/.ssh/authorized_keys`)
* Killing/Disabling Critical System Services (especially endpoint security solutions such as MDE)
* Launching Cloud Resources (EC2 Instances/Docker Containers/Kubernetes Pods)

Failure to clean up properly after such activity will at the very least inconvenience the user, 
and at worst may actually create severe new security vulnerabilities in the target system. 

Interpreter scripts usually lack standardized, platform-independent, and reliable methods for
cleaning up attacker activity. In particular, it is especially difficult to ensure that these clean up
processes are resilient to runtime errors. For example, when writing TTPs in bash, one would
need to make very careful use of the [trap](https://tldp.org/LDP/Bash-Beginners-Guide/html/sect_12_02.html) feature.

TTPForge's native support for cleanup actions provides users
a reliable solution to these problems.

## A Basic Cleanup Action

Every [step](steps.md) in a TTPForge TTP can be associated with a specific cleanup action.
For example, we could create a cleanup action that removes a file created during the 
execution of the TTP, as shown below:

https://github.com/facebookincubator/TTPForge/blob/main/cmd/test-resources/repos/example-ttps/cleanup/single.yaml

This cleanup action will be executed once the TTP completes.
Cleanup actions are most commonly `inline:` commands, as in the example above. 
However, you can use any valid TTPForge step type as a cleanup action.

## Multiple Cleanup Actions for Multiple Steps

If we have multiple steps, we can create an appropriate cleanup action for each one.
The example TTP below demonstrates how this works - for clarity, we've simplified the actual
commands by replacing them with placeholders:

https://github.com/facebookincubator/TTPForge/blob/main/cmd/test-resources/repos/example-ttps/cleanup/multiple.yaml

This example highlights a crucial TTPForge feature: cleanup actions are run in last-in, 
first-out (LIFO) order. Therefore, In this example, the malicious launch agent will be deleted prior to
re-enabling EDR. This is critical; if cleanup actions were instead run in the same
order as the steps themselves (first in, first out), then defenders would 
receive unrealistic signal (EDR would detect the payload) and the fidelity of the simulation
would be compromised. 

## Fault Tolerance - Cleaning When Things Go Wrong

If a particular step from the TTP fails to execute, prior steps that already executed
still need to be cleaned up properly. Therefore, after each step, the corresponding 
cleanup action is enqueued and, in the case of failure, all enqueued steps are executed
in reverse (LIFO) order. The below TTP illustrates this behavior:

https://github.com/facebookincubator/TTPForge/blob/main/cmd/test-resources/repos/example-ttps/cleanup/lifo.yaml

As the third step is guaranteed to fail, 
this TTP will always print the following cleanup activity

```
cleaning up second step
cleaning up first step
```

## Delaying or Disabling Cleanup

Sometimes, one may wish to execute a given TTP and then to leave the target system in 
a "dirty" state for further analysis. For these purposes, TTPForge provides
two useful command line flags for the `ttpforge run` command:

* `--cleanup-delay-seconds` - delay cleanup execution for the specified integer number of seconds
* `--no-cleanup` - do not run any cleanup actions; instead, simply exit when the last step completes.
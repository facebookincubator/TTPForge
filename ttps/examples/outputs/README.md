# TTPs

These TTPs illustrate how to pass outputs between different steps.

## step-outputs.yaml


Run this TTP as follows:

```
ttpforge run ttps/examples/outputs/step-outputs.yaml
```

Expected output:

```
this will be accessible in stdout
previous step output is this will be accessible in stdout
{"foo":"bar"}
bar
```
# TTPs

These TTPs illustrate how to use the CLI argument features of TTPForge.

## define-args.yaml

Run this TTP as follows:

```bash
ttpforge run examples/args/define-args.yaml \
  --arg a_message=foo \
  --arg a_number=1337
```

Expected output:

```text
hi! You passed the message: foo
You passed the number: 1337
has_a_default has the value: 'this is the default value'
```

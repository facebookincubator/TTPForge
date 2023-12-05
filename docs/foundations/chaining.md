# Chaining TTPs Together

TTPForge provides users with the ability to chain multiple existing TTPs
together to form new composite TTPs. This is useful for two primary reasons:

1. Users can simulate complex multi-stage cyberattacks.
1. Duplication of code is avoided because steps that are shared across multiple
   TTPs can be combined together.

## Syntax for Chaining TTPs

To chain multiple TTPs together, use the `ttp:` action, as shown in the example
below:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/chaining/basic.yaml#L1-L21

Run this example TTP by executing the following command:

```bash
ttpforge run examples//chaining/basic.yaml
```

Notice that the steps of each sub-TTP referenced via the `ttp:` action are
executed in sequence - our example has therefore combined two smaller TTPs into
a single larger one. TTP chains may consist of as many TTPs as desired.

The `ttp:` action accepts a
[TTP reference](repositories.md#listing-and-examining-ttps-in-repositories) as
its argument. The `//` prefix indicates that the provided TTP path is relative
to root of the current repository's
[TTP Search Path](repositories.md#repository-configuration-files). Therefore, in
the case of this repository, the provided path is rooted in the `example-ttps`
directory. Consult the [TTP Repositories](repositories.md) documentation for
further details about how TTP references are resolved.

**Note**: for legacy reasons, TTPForge also supports omitting the `//` prefix in
`ttp:` actions. Paths provided without the `//` prefix are still resolved
relative to the TTP search path root, just as if the `//` was present. This
compatibility may be removed in a later version of TTPForge; therefore, new TTPs
should always use the `//`.

## Passing Arguments to Sub-TTPs

The example above also showcases the `args:` syntax that is used to pass
arguments to sub-TTPs. The specified argument values are mapped directly to the
[command-line arguments](args.md) that are declared in the YAML file of the
sub-TTP.

## Cleaning Up TTP Chains

The TTPForge [cleanup](cleanup.md) feature works somewhat differently than usual
for TTP chains. TTPForge automatically adds a special cleanup action to each
`ttp:` step. This cleanup action runs the cleanup actions defined in the
referenced sub-TTP file. If a step from the sub-TTP fails, this cleanup action
will begin sub-TTP cleanup execution from the last successful step of the
sub-TTP.

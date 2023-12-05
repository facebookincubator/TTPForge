# Customizing Your TTPs with Command-Line Arguments

**Note**: to run the examples in this section, make sure you have the `examples`
repository installed with `ttpforge list repos` and if not run:

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

## Basics of Command-Line Arguments

TTPForge allows users to control TTP execution through its support for
command-line arguments - check out the TTP below to see how it works:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/args/basic.yaml#L1-L33

You can run this TTP and provide values for all relevant arguments as follows:

```bash
ttpforge run examples//args/basic.yaml \
  --arg str_to_print=hello \
  --arg run_second_step=true
```

Try out the following exercises to increase your understanding of how arguments
work in TTPForge:

- Remove `--arg str_to_print="..."` - the TTP will now refuse to run because the
  user is required to specify a value for that argument since it has no
  `default` value specified.
- Explicitly set `int_arg` with `--arg int_arg=5` - the `default` value will be
  overridden.
- Try to pass values with invalid types, such as `--arg int_arg=foo` or
  `--arg run_second_step=bar`. TTPForge validates argument types and should
  throw an error for both of these cases. Note that `string` type arguments (the
  default) will pretty much accept anything.
- Disable the second step by removing the `--arg run_second_step=true` line.

Focus in particular on the last item above, concerning `run_second_step=true`.
TTPForge TTP files are processed using Golang's
[text/template](https://pkg.go.dev/text/template) package to expand all argument
values prior to execution. We can use advanced templating features, such as the
[if-else-end](https://pkg.go.dev/text/template#hdr-Actions) shown above, to
precisely control execution based on argument values.

## Argument Types

TTPForge supports the following argument types (which you can specify with the
`type:` field as shown in the example above):

- `string` (this is the default if no `type` is specified)
- `int`
- `bool`
- `path` (a very important one - see below)

## The `path` Argument Type

If your TTP will accept arguments that refer to file paths on disk, you should
**almost always** use `type: path` when declaring those arguments, as shown in
the example below (just focus on the `args:` section for now, though you can
check out the [edit_file documentation](actions/edit_file.md) if you are
curious):

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/actions/edit-file/append-delete.yaml#L1-L35

You must use `type: path` because when you execute `ttpforge run [ttp]`,
**TTPForge changes its working directory to the folder containing the TTP.**
This means that relative paths such as `foo/bar` won't retain their original
meaning by default - however, when you declare your argument using `type: path`,
TTPForge knows to expand its value to an absolute path prior to changing
directories, ensuring that everything will work as intended.

## Predefined Choices for Argument Values

Sometimes only certain specific values make sense for a given argument. TTPForge
lets you restrict the allowed values for an argument using the `choices:`
keyword, as shown below

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/args/choices.yaml#L1-L29

You can run the above TTP as follows:

```bash
ttpforge run examples//args/choices.yaml \
  --arg arg_with_choices=C
```

Notice the following key aspects of the `choices` feature:

- TTPForge will reject your arguments if you specify an invalid choice such as
  `arg_with_choices=D`.
- If you use `choices` and `default` together, the `default` value must be one
  of the valid choices.

## Validating Arguments with Regular Expressions

In order to require user-provided argument values to match a particular regular
expression (which is useful for ensuring that you don't get strange errors
halfway through a TTP due to user error) you can use the `regexp:` syntax
demonstrated below:

https://github.com/facebookincubator/TTPForge/blob/7634dc65879ec43a108a4b2d44d7eb2105a2a4b1/example-ttps/args/regexp.yaml#L1-L20

You can use any regular expression allowed by the Golang regular expression
[syntax](https://pkg.go.dev/regexp), although if you use YAML metacharacters
such as `:` you are advised to put quotes around your regexp to ensure that your
TTP YAML remains valid.

You can run the above TTP as follows:

```bash
ttpforge run examples//args/regexp.yaml \
  --arg must_contain_ab=xabyabz \
  --arg must_start_with_1_end_with_7=1337
```

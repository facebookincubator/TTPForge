# TTPForge Actions: `expect`

The `expect` action is designed to automate interactions with command-line
programs that require user input. By scripting responses to prompts, `expect`
allows you to execute commands that would otherwise pause for manual input,
making it ideal for automating tasks in interactive environments. This action
leverages the go-expect library to handle the automation of these interactions.

## Automating Interactive Scripts

This example demonstrates how to use the `expect` action to automate interaction
with an interactive Python script:
https://github.com/facebookincubator/TTPForge/blob/ffe3d206d747c27d1043cd0a10517831568ee364/example-ttps/actions/expect/expect.yaml#L1C1-L30C29

You can experiment with the above TTP by installing the `examples` TTP
repository (skip this if `ttpforge list repos` shows that the `examples` repo is
already installed):

```bash
ttpforge install repo https://github.com/facebookincubator/TTPForge --name examples
```

and then running the below command:

```bash
ttpforge run examples//actions/expect/expect.yaml
```

## Fields

You can specify the following YAML fields for the `expect` action:

- `expect:` (type: `string`) the command to execute that requires interaction.
- `responses:` (type: `list`) a list of responses to provide. Each entry can
  contain the following fields:
  - `prompt:` (type: `string`) the text prompt to expect from the command.
  - `response:` (type: `string`) the response to provide when the prompt is
    encountered.
- `cleanup:` Define a custom
  [cleanup action](https://github.com/facebookincubator/TTPForge/blob/main/docs/foundations/cleanup.md#cleanup-basics)
  to execute after the expect action completes..

## Notes

- The `expect` action is particularly useful for automating tasks that involve
  interactive command-line tools, ensuring that your TTPs can run unattended.
- Ensure that the prompts specified in `responses` match exactly with the output
  of the command to ensure correct automation.
- The `expect` action can be used in combination with other actions to create
  complex, automated workflows.

# TTPForge/blocks

The `blocks` package is a part of the TTPForge.

---

## Table of contents

- [Functions](#functions)
- [Installation](#installation)
- [Usage](#usage)
- [Tests](#tests)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### BasicStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the BasicStep and returns an error if any occur.

---

### BasicStep.IsNil()

```go
IsNil() bool
```

IsNil checks if a BasicStep is considered empty or uninitialized.

---

### BasicStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the BasicStep, checking for the necessary attributes and dependencies.

---

### CreateFileStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the step and returns an error if any occur.

---

### CreateFileStep.GetDefaultCleanupAction()

```go
GetDefaultCleanupAction() Action
```

GetDefaultCleanupAction will instruct the calling code
to remove the path created by this action

---

### CreateFileStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the step is nil or empty and returns a boolean value.

---

### CreateFileStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the step

**Returns:**

error: An error if any validation checks fail.

---

### EditStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the EditStep and returns an error if any occur.

---

### EditStep.IsNil()

```go
IsNil() bool
```

IsNil checks if an EditStep is considered empty or uninitialized.

---

### EditStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the EditStep, checking for the necessary attributes and dependencies.

---

### FetchAbs(string, string)

```go
FetchAbs(string, string) string, error
```

FetchAbs returns the absolute path of a file given its path and the
working directory. It handles cases where the path starts with "~/",
is an absolute path, or is a relative path from the working directory.
It logs any errors and returns them.

**Parameters:**

path: A string representing the path to the file.

workdir: A string representing the working directory.

**Returns:**

fullpath: A string representing the absolute path to the file.

error: An error if the path cannot be resolved to an absolute path.

---

### FetchEnv(map[string]string)

```go
FetchEnv(map[string]string) []string
```

FetchEnv converts an environment variable map into a slice of strings that
can be used as an argument when running a command.

**Parameters:**

environ: A map of environment variable names to values.

**Returns:**

[]string: A slice of strings representing the environment variables
and their values.

---

### FetchURIStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) *ActResult, error
```

Cleanup is a method to establish a link with the Cleanup interface.
Assumes that the type is the cleanup step and is invoked by
f.CleanupStep.Cleanup.

---

### FetchURIStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the FetchURIStep and returns an error if any occur.

---

### FetchURIStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the FetchURIStep is nil or empty and returns a boolean value.

---

### FetchURIStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the FetchURIStep. It checks that the
Act field is valid, Location is set with
a valid file path, and Uri is set.

If Location is set, it ensures that the path exists and retrieves
its absolute path.

**Returns:**

error: An error if any validation checks fail.

---

### FileStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) *ActResult, error
```

Cleanup is a method to establish a link with the Cleanup interface.
Assumes that the type is the cleanup step and is invoked by
f.CleanupStep.Cleanup.

---

### FileStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the FileStep and returns an error if any occur.

---

### FileStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the FileStep is nil or empty and returns a boolean value.

---

### FileStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the FileStep. It checks that the
Act field is valid, and that either FilePath is set with
a valid file path, or InlineLogic is set with valid code.

If FilePath is set, it ensures that the file exists and retrieves
its absolute path.

If Executor is not set, it infers the executor based on the file extension.
It then checks that the executor is in the system path, and if CleanupStep
is not nil, it validates the cleanup step as well.
It logs any errors and returns them.

**Returns:**

error: An error if any validation checks fail.

---

### FindFilePath(string, string, fs.StatFS)

```go
FindFilePath(string, string, fs.StatFS) string, error
```

FindFilePath checks if a file exists given its path, the working directory,
and an optional fs.StatFS. It handles cases where the path starts with "../",
"~/", or is a relative path. It also checks a list of paths in InventoryPath
for the file. It logs any errors and returns them.

**Parameters:**

path: A string representing the path to the file.

workdir: A string representing the working directory.

system: An optional fs.StatFS that can be used to check if the file exists.

**Returns:**

string: A string representing the path to the file, or an empty string
if the file does not exist.

error: An error if the file cannot be found or if other errors occur.

---

### InferExecutor(string)

```go
InferExecutor(string) string
```

InferExecutor infers the executor based on the file extension and
returns it as a string.

---

### LoadTTP(string, afero.Fs, *TTPExecutionConfig, []string)

```go
LoadTTP(string afero.Fs *TTPExecutionConfig []string) *TTP *TTPExecutionContext error
```

LoadTTP reads a TTP file and creates a TTP instance based on its contents.
If the file is empty or contains invalid data, it returns an error.

**Parameters:**

ttpFilePath: the absolute or relative path to the TTP YAML file.
fsys: an afero.Fs that contains the specified TTP file path

**Returns:**

*TTP: Pointer to the created TTP instance, or nil if the file is empty or invalid.
TTPExecutionContext: the initialized TTPExecutionContext suitable for passing to TTP.Execute(...)
err: An error if the file contains invalid data or cannot be read.

---

### NewBasicStep()

```go
NewBasicStep() *BasicStep
```

NewBasicStep creates a new BasicStep instance with an initialized Act struct.

---

### NewCreateFileStep()

```go
NewCreateFileStep() *CreateFileStep
```

NewCreateFileStep creates a new CreateFileStep instance and returns a pointer to it.

---

### NewEditStep()

```go
NewEditStep() *EditStep
```

NewEditStep creates a new EditStep instance with an initialized Act struct.

---

### NewFetchURIStep()

```go
NewFetchURIStep() *FetchURIStep
```

NewFetchURIStep creates a new FetchURIStep instance and returns a pointer to it.

---

### NewFileStep()

```go
NewFileStep() *FileStep
```

NewFileStep creates a new FileStep instance and returns a pointer to it.

---

### NewStepResultsRecord()

```go
NewStepResultsRecord() *StepResultsRecord
```

NewStepResultsRecord generates an appropriately initialized StepResultsRecord

---

### NewSubTTPStep()

```go
NewSubTTPStep() *SubTTPStep
```

NewSubTTPStep creates a new SubTTPStep and returns a pointer to it.

---

### PrintStrAction.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the step and returns an error if any occur.

---

### PrintStrAction.IsNil()

```go
IsNil() bool
```

IsNil checks if the step is nil or empty and returns a boolean value.

---

### PrintStrAction.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the step

**Returns:**

error: An error if any validation checks fail.

---

### RemovePathAction.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the step and returns an error if any occur.

---

### RemovePathAction.IsNil()

```go
IsNil() bool
```

IsNil checks if the step is nil or empty and returns a boolean value.

---

### RemovePathAction.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the step

**Returns:**

error: An error if any validation checks fail.

---

### RenderTemplatedTTP(string, *TTPExecutionConfig)

```go
RenderTemplatedTTP(string, *TTPExecutionConfig) *TTP, error
```

RenderTemplatedTTP is a function that utilizes Golang's `text/template` for template substitution.
It replaces template expressions like `{{ .Args.myarg }}` with corresponding values.
This function must be invoked prior to YAML unmarshaling, as the template syntax `{{ ... }}`
may result in invalid YAML under specific conditions.

**Parameters:**

ttpStr: A string containing the TTP template to be rendered.
execCfg: A pointer to a TTPExecutionConfig that represents the execution configuration for the TTP.

**Returns:**

*TTP: A pointer to the TTP object created from the template.
error: An error if the rendering or unmarshaling process fails.

---

### ShouldUseImplicitDefaultCleanup(Action)

```go
ShouldUseImplicitDefaultCleanup(Action) bool
```

ShouldUseImplicitDefaultCleanup is a hack
to make subTTPs always run their default
cleanup process even when `cleanup: default` is
not explicitly specified - this is purely for backward
compatibility

---

### Step.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) *ActResult, error
```

Cleanup runs the cleanup action associated with this step

---

### Step.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs the action associated with this step

---

### Step.ParseAction(*yaml.Node)

```go
ParseAction(*yaml.Node) Action, error
```

ParseAction decodes an action (from step or cleanup) in YAML
format into the appropriate struct

---

### Step.ShouldCleanupOnFailure()

```go
ShouldCleanupOnFailure() bool
```

ShouldCleanupOnFailure specifies that this step should be cleaned
up even if its Execute(...)  failed.
We usually don't want to do this - for example,
you shouldn't try to remove_path a create_file that failed)
However, certain step types (especially SubTTPs) need to run cleanup even if they fail

---

### Step.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML implements custom deserialization
process to ensure that the step action and its
cleanup action are decoded to the correct struct type

---

### Step.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate checks that both the step action and cleanup
action are valid

---

### SubTTPStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute runs each step of the TTP file associated with the SubTTPStep
and manages the outputs and cleanup steps.

---

### SubTTPStep.GetDefaultCleanupAction()

```go
GetDefaultCleanupAction() Action
```

GetDefaultCleanupAction will instruct the calling code
to cleanup all successful steps of this subTTP

---

### SubTTPStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the SubTTPStep is empty or uninitialized.

---

### SubTTPStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate checks the validity of the SubTTPStep by ensuring
the following conditions are met:
The associated Act is valid.
The TTP file associated with the SubTTPStep can be successfully unmarshalled.
The TTP file path is not empty.
The steps within the TTP file do not contain any nested SubTTPSteps.
If any of these conditions are not met, an error is returned.

---

### TTP.Execute(*TTPExecutionContext)

```go
Execute(*TTPExecutionContext) *StepResultsRecord, error
```

Execute executes all of the steps in the given TTP,
then runs cleanup if appropriate

**Parameters:**

execCfg: The TTPExecutionConfig for the current TTP.

**Returns:**

*StepResultsRecord: A StepResultsRecord containing the results of each step.
error: An error if any of the steps fail to execute.

---

### TTP.MarshalYAML()

```go
MarshalYAML() interface{}, error
```

MarshalYAML is a custom marshalling implementation for the TTP structure.
It encodes a TTP object into a formatted YAML string, handling the
indentation and structure of the output YAML.

**Returns:**

interface{}: The formatted YAML string representing the TTP object.
error: An error if the encoding process fails.

---

### TTP.RunSteps(*TTPExecutionContext)

```go
RunSteps(*TTPExecutionContext) *StepResultsRecord, int, error
```

RunSteps executes all of the steps in the given TTP.

**Parameters:**

execCtx: The current TTPExecutionContext

**Returns:**

*StepResultsRecord: A StepResultsRecord containing the results of each step.
int: the index of the step where cleanup should start (usually the last successful step)
error: An error if any of the steps fail to execute.

---

### TTP.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate ensures that all components of the TTP are valid
It checks key fields, then iterates through each step
and validates them in turn

**Parameters:**

execCtx: The TTPExecutionContext for the current TTP.

**Returns:**

error: An error if any part of the validation fails, otherwise nil.

---

### TTPExecutionContext.ExpandVariables([]string)

```go
ExpandVariables([]string) []string, error
```

ExpandVariables takes a string containing the following types of variables
and expands all of them to their appropriate values:

* Step outputs: ($forge.steps.bar.outputs.baz)

**Parameters:**

inStrs: the list of strings that have variables expanded

**Returns:**

[]string: the corresponding strings with variables expanded
error: an error if there is a problem

---

### actionDefaults.GetDefaultCleanupAction()

```go
GetDefaultCleanupAction() Action
```

GetDefaultCleanupAction provides a default implementation
of the GetDefaultCleanupAction method from the Action interface.
This saves us from having to declare this function for every steps
If a specific action needs a default cleanup action (such as a create_file action),
it can override this step

---

### subTTPCleanupAction.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ActResult, error
```

Execute will cleanup the subTTP starting from the last successful step

---

### subTTPCleanupAction.IsNil()

```go
IsNil() bool
```

IsNil is not needed here, as this is not a user-accessible step type

---

### subTTPCleanupAction.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate is not needed here, as this is not a user-accessible step type

---

## Installation

To use the TTPForge/blocks package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/facebookincubator/ttpforge/blocks
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/facebookincubator/ttpforge/blocks"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/blocks`:

```bash
go test -v
```

---

## Contributing

Pull requests are welcome. For major changes,
please open an issue first to discuss what
you would like to change.

---

## License

This project is licensed under the MIT
License - see the [LICENSE](https://github.com/facebookincubator/TTPForge/blob/main/LICENSE)
file for details.

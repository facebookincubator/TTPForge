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

### Act.CheckCondition()

```go
CheckCondition() bool, error
```

CheckCondition checks the condition specified for an Act and returns true
if it matches the current OS, false otherwise. If the condition is "always",
the function returns true.
If an error occurs while checking the condition, it is returned.

**Returns:**

bool: true if the condition matches the current OS or the
condition is "always", false otherwise.

error: An error if an error occurs while checking the condition.

---

### Act.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error explaining why the Act is invalid.

**Returns:**

error: An error explaining why the Act is invalid, or nil
if the Act is valid.

---

### Act.IsNil()

```go
IsNil() bool
```

IsNil checks whether the Act is nil (i.e., it does not have a name).

**Returns:**

bool: True if the Act has no name, false otherwise.

---

### Act.MakeCleanupStep(*yaml.Node)

```go
MakeCleanupStep(*yaml.Node) CleanupAct, error
```

MakeCleanupStep creates a CleanupAct based on the given yaml.Node.
If the node is empty or invalid, it returns nil. If the node contains a
BasicStep or FileStep, the corresponding CleanupAct is created and returned.

**Parameters:**

node: A pointer to a yaml.Node containing the parameters to
create the CleanupAct.

**Returns:**

CleanupAct: The created CleanupAct, or nil if the node is empty or invalid.

error: An error if the node contains invalid parameters.

---

### Act.SetDir(string)

```go
SetDir(string)
```

SetDir sets the working directory for the Act.

**Parameters:**

dir: A string representing the directory path to be set
as the working directory.

---

### Act.StepName()

```go
StepName() string
```

StepName returns the name of the Act.

**Returns:**

string: The name of the Act.

---

### Act.Validate()

```go
Validate() error
```

Validate checks the Act for any validation errors, such as the presence of
spaces in the name.

**Returns:**

error: An error if any validation errors are found, or nil if
the Act is valid.

---

### BasicStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) *ActResult, error
```

Cleanup is an implementation of the CleanupAct interface's Cleanup method.

---

### BasicStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ExecutionResult, error
```

Execute runs the BasicStep and returns an error if any occur.

---

### BasicStep.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error with an explanation of why a BasicStep is invalid.

---

### BasicStep.GetCleanup()

```go
GetCleanup() []CleanupAct
```

GetCleanup returns the cleanup steps for a BasicStep.

---

### BasicStep.GetType()

```go
GetType() StepType
```

GetType returns the step type for a BasicStep.

---

### BasicStep.IsNil()

```go
IsNil() bool
```

IsNil checks if a BasicStep is considered empty or uninitialized.

---

### BasicStep.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML custom unmarshaler for BasicStep to handle decoding from YAML.

---

### BasicStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the BasicStep, checking for the necessary attributes and dependencies.

---

### EditStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ExecutionResult, error
```

Execute runs the EditStep and returns an error if any occur.

---

### EditStep.GetCleanup()

```go
GetCleanup() []CleanupAct
```

GetCleanup returns the cleanup steps for a EditStep.
Currently this is always empty because we use backup
files instead for this type of step

---

### EditStep.GetType()

```go
GetType() StepType
```

GetType returns the step type for a EditStep.

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
Execute(TTPExecutionContext) *ExecutionResult, error
```

Execute runs the FileStep and returns an error if any occur.

---

### FileStep.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error message explaining why the FileStep
is invalid.

**Returns:**

error: An error message explaining why the FileStep is invalid.

---

### FileStep.GetCleanup()

```go
GetCleanup() []CleanupAct
```

GetCleanup returns a slice of CleanupAct if the CleanupStep is not nil.

---

### FileStep.GetType()

```go
GetType() StepType
```

GetType returns the type of the step as StepType.

---

### FileStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the FileStep is nil or empty and returns a boolean value.

---

### FileStep.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML decodes a YAML node into a FileStep instance. It uses
the provided struct as a template for the YAML data, and initializes the
FileStep instance with the decoded values.

**Parameters:**

node: A pointer to a yaml.Node representing the YAML data to decode.

**Returns:**

error: An error if there is a problem decoding the YAML data.

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
LoadTTP(string, afero.Fs, *TTPExecutionConfig, []string) *TTP, error
```

LoadTTP reads a TTP file and creates a TTP instance based on its contents.
If the file is empty or contains invalid data, it returns an error.

**Parameters:**

ttpFilePath: the absolute or relative path to the TTP YAML file.
fsys: an afero.Fs that contains the specified TTP file path

**Returns:**

ttp: Pointer to the created TTP instance, or nil if the file is empty or invalid.
err: An error if the file contains invalid data or cannot be read.

---

### NewBasicStep()

```go
NewBasicStep() *BasicStep
```

NewBasicStep creates a new BasicStep instance with an initialized Act struct.

---

### NewEditStep()

```go
NewEditStep() *EditStep
```

NewEditStep creates a new EditStep instance with an initialized Act struct.

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

### SubTTPStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) *ActResult, error
```

Cleanup runs the cleanup actions associated with all successful sub-steps

---

### SubTTPStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) *ExecutionResult, error
```

Execute runs each step of the TTP file associated with the SubTTPStep
and manages the outputs and cleanup steps.

---

### SubTTPStep.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid checks for invalid data in the SubTTPStep
and returns an error explaining any issues found.
Currently, it checks if the TtpFile field is empty.

---

### SubTTPStep.GetCleanup()

```go
GetCleanup() []CleanupAct
```

GetCleanup returns a slice of CleanupAct associated with the SubTTPStep.

---

### SubTTPStep.GetType()

```go
GetType() StepType
```

GetType returns the type of the step (StepSubTTP for SubTTPStep).

---

### SubTTPStep.IsNil()

```go
IsNil() bool
```

IsNil checks if the SubTTPStep is empty or uninitialized.

---

### SubTTPStep.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML is a custom unmarshaller for SubTTPStep which decodes
a YAML node into a SubTTPStep instance.

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

### TTP.RunSteps(TTPExecutionConfig)

```go
RunSteps(TTPExecutionConfig) *StepResultsRecord, error
```

RunSteps executes all of the steps in the given TTP.

**Parameters:**

t: The TTP to execute the steps for.

**Returns:**

error: An error if any of the steps fail to execute.

---

### TTP.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML is a custom unmarshalling implementation for the TTP structure.
It decodes a YAML Node into a TTP object, handling the decoding and
validation of the individual steps within the TTP.

**Parameters:**

node: A pointer to a yaml.Node that represents the TTP structure
to be unmarshalled.

**Returns:**

error: An error if the decoding process fails or if the TTP structure contains invalid steps.

---

### TTP.ValidateSteps(TTPExecutionContext)

```go
ValidateSteps(TTPExecutionContext) error
```

ValidateSteps iterates through each step in the TTP and validates it.
It sets the working directory for each step before calling its Validate
method. If any step fails validation, the method returns an error.
If all steps are successfully validated, the method returns nil.

**Returns:**

error: An error if any step validation fails, otherwise nil.

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

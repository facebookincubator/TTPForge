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

CheckCondition checks the condition specified for an Act and returns true if it matches the current OS, false otherwise.
If the condition is "always", the function returns true. If an error occurs while checking the condition, it is returned.

Returns:

bool: true if the condition matches the current OS or the condition is "always", false otherwise.
error: An error if an error occurs while checking the condition.

---

### Act.Cleanup(map[string]string)

```go
Cleanup(map[string]string) error
```

Cleanup is a placeholder function for the base Act. Subtypes can override
this method to implement their own cleanup logic.

Returns:

error: Always returns nil for the base Act.

---

### Act.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error explaining why the Act is invalid.

Returns:

error: An error explaining why the Act is invalid, or nil if the Act is valid.

---

### Act.FetchArgs([]string)

```go
FetchArgs([]string) []string
```

FetchArgs processes a slice of arguments and returns a new slice with the
output values of referenced steps.

Parameters:

args: A slice of strings representing the arguments to be processed.

Returns:

[]string: A slice of strings containing the processed output values of referenced steps.

---

### Act.GetOutput()

```go
GetOutput() map[string]any
```

GetOutput returns the output map of the Act.

Returns:

map[string]any: The output map of the Act.

---

### Act.IsNil()

```go
IsNil() bool
```

IsNil checks whether the Act is nil (i.e., it does not have a name).

Returns:

bool: True if the Act has no name, false otherwise.

---

### Act.MakeCleanupStep(*yaml.Node)

```go
MakeCleanupStep(*yaml.Node) CleanupAct, error
```

MakeCleanupStep creates a CleanupAct based on the given yaml.Node. If the node is empty or invalid, it returns nil.
If the node contains a BasicStep or FileStep, the corresponding CleanupAct is created and returned.

Parameters:

node: A pointer to a yaml.Node containing the parameters to create the CleanupAct.

Returns:

CleanupAct: The created CleanupAct, or nil if the node is empty or invalid.
error: An error if the node contains invalid parameters.

---

### Act.SearchOutput(string)

```go
SearchOutput(string) string
```

SearchOutput searches for the Output value of a step by parsing the provided
argument.

Parameters:

arg: A string representing the argument in the format "steps.step_name.output".

Returns:

string: The Output value of the step as a string, or the original argument
if the step is not found or the argument is in an incorrect format.

---

### Act.SetDir(string)

```go
SetDir(string)
```

SetDir sets the working directory for the Act.

Parameters:

dir: A string representing the directory path to be set as the working directory.

---

### Act.SetOutputSuccess(*bytes.Buffer, int)

```go
SetOutputSuccess(*bytes.Buffer, int)
```

SetOutputSuccess sets the output of an Act to a given buffer and sets the success flag to true or false depending on the exit code.
If the output can be unmarshalled into a JSON structure, it is stored as a string in the Act's output map.

Parameters:

output: A pointer to a bytes.Buffer containing the output to set as the Act's output.
exit: An integer representing the exit code of the Act.

---

### Act.Setup(map[string]string, map[string]Step)

```go
Setup(map[string]string, map[string]Step)
```

Setup initializes the Act with the given environment and output reference maps.

Parameters:

env: A map of environment variables, where the keys are variable names and the values are variable values.
outputRef: A map of output references, where the keys are step names and the values are Step instances.

Returns:

map[string]: Step instances.

---

### Act.StepName()

```go
StepName() string
```

StepName returns the name of the Act.

Returns:

string: The name of the Act.

---

### Act.Success()

```go
Success() bool
```

Success returns the success status of the Act.

Returns:

bool: The success status of the Act.

---

### Act.Validate()

```go
Validate() error
```

Validate checks the Act for any validation errors, such as the presence of
spaces in the name. It returns an error if any validation errors are found.

Returns:

error: An error if any validation errors are found, or nil if the Act is valid.

---

### BasicStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) error
```

Cleanup executes the CleanupAct associated with this BasicStep.

**Parameters:**

execCtx: The execution context for the TTP.

**Returns:**

error: An error if any issue occurs while executing the CleanupAct.

---

### BasicStep.CleanupName()

```go
CleanupName() string
```

CleanupName returns the name of the cleanup step associated with this
BasicStep.

**Returns:**

string: The name of the cleanup step.

---

### BasicStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) error
```

Execute runs this BasicStep and returns an error if any issue occurs. It
sets up a context with a timeout, then if an inline command is provided,
executes it.

**Parameters:**

execCtx: The execution context for the TTP.

**Returns:**

error: An error if any issue occurs during execution.

---

### BasicStep.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error with an explanation of why this
BasicStep is invalid. The error explains that the 'inline' field is
empty.

**Returns:**

error: An error explaining that the 'inline' field is empty.

---

### BasicStep.GetCleanup()

```go
GetCleanup() []CleanupAct
```

GetCleanup returns the cleanup steps for a BasicStep.

**Returns:**

[]CleanupAct: A slice of CleanupAct, which is a cleanup step for this
BasicStep.

---

### BasicStep.GetType()

```go
GetType() StepType
```

GetType returns the step type for this BasicStep.

**Returns:**

StepType: The type of this step.

---

### BasicStep.IsNil()

```go
IsNil() bool
```

IsNil checks if this BasicStep is considered empty or uninitialized.

**Returns:**

bool: True if this BasicStep is empty or uninitialized, false otherwise.

---

### BasicStep.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML is a custom unmarshaler for BasicStep to handle decoding
from YAML.

This function first decodes the provided yaml.Node into a temporary
structure. Then, it copies the relevant data to the receiver BasicStep,
and checks if the step is not empty. If the step includes a cleanup step
and is not of type StepCleanup, it tries to create a CleanupStep from
the corresponding yaml.Node.

**Parameters:**

node: The yaml.Node to decode.

**Returns:**

error: An error if decoding, checking the step, or creating a
CleanupStep fails.

---

### BasicStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates this BasicStep, checking for the necessary attributes
and dependencies. It checks that the Act and inline command are valid,
the executor is either provided or defaults to "bash", the executor is
present in the system path, and if a CleanupStep is provided, it is
valid.

**Parameters:**

execCtx: The execution context for the TTP.

**Returns:**

error: An error if any issue occurs during validation.

---

### Contains(string, map[string]any)

```go
Contains(string, map[string]any) bool
```

Contains checks if a key exists in a map.

Parameters:

key: A string representing the key to search for.
search: A map of keys and values.

Returns:

bool: A boolean value indicating if the key was found in the map.

---

### EditStep.CleanupName()

```go
CleanupName() string
```

CleanupName returns the name of the cleanup step.

---

### EditStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) error
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

FetchAbs returns the absolute path of a file given its path and the working directory. It handles cases where the path starts with "~/",
is an absolute path, or is a relative path from the working directory. It logs any errors and returns them.

Parameters:

path: A string representing the path to the file.
workdir: A string representing the working directory.

Returns:

fullpath: A string representing the absolute path to the file.
error: An error if the path cannot be resolved to an absolute path.

---

### FetchEnv(map[string]string)

```go
FetchEnv(map[string]string) []string
```

FetchEnv converts an environment variable map into a slice of strings that can be used as an argument when running a command.

Parameters:

environ: A map of environment variable names to values.

Returns:

[]string: A slice of strings representing the environment variables and their values.

---

### FileStep.Cleanup(TTPExecutionContext)

```go
Cleanup(TTPExecutionContext) error
```

Cleanup is a method to establish a link with the Cleanup interface.
Assumes that the type is the cleanup step and is invoked by f.CleanupStep.Cleanup.

---

### FileStep.CleanupName()

```go
CleanupName() string
```

CleanupName returns the name of the cleanup action as a string.

---

### FileStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) error
```

Execute runs the FileStep and returns an error if any occur.

---

### FileStep.ExplainInvalid()

```go
ExplainInvalid() error
```

ExplainInvalid returns an error message explaining why the FileStep is invalid.

Returns:

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

UnmarshalYAML decodes a YAML node into a FileStep instance. It uses the provided
struct as a template for the YAML data, and initializes the FileStep instance with the
decoded values.

Parameters:

node: A pointer to a yaml.Node representing the YAML data to decode.

Returns:

error: An error if there is a problem decoding the YAML data.

---

### FileStep.Validate(TTPExecutionContext)

```go
Validate(TTPExecutionContext) error
```

Validate validates the FileStep. It checks that the
Act field is valid, and that either FilePath is set with
a valid file path, or InlineLogic is set with valid code.

If FilePath is set, it ensures that the file exists and retrieves its absolute path.

If Executor is not set, it infers the executor based on the file extension. It then checks that the executor is in the system path,
and if CleanupStep is not nil, it validates the cleanup step as well. It logs any errors and returns them.

Returns:

error: An error if any validation checks fail.

---

### FindFilePath(string, string, fs.StatFS)

```go
FindFilePath(string, string, fs.StatFS) string, error
```

FindFilePath checks if a file exists given its path, the working directory, and an optional fs.StatFS. It handles cases where the path starts with "../",
"~/", or is a relative path. It also checks a list of paths in InventoryPath for the file. It logs any errors and returns them.

Parameters:

path: A string representing the path to the file.
workdir: A string representing the working directory.
system: An optional fs.StatFS that can be used to check if the file exists.

Returns:

* A string representing the path to the file, or an empty string if the file does not exist.
* An error if the file cannot be found or if other errors occur.

---

### InferExecutor(string)

```go
InferExecutor(string) string
```

InferExecutor infers the executor based on the file extension and returns it as a string.

---

### JSONString(any)

```go
JSONString(any) string, error
```

JSONString returns a string representation of an object in JSON format.

Parameters:

in: An object of any type.

Returns:

string: A string representing the object in JSON format.
error: An error if the object cannot be encoded as JSON.

---

### LoadTTP(string, fs.StatFS)

```go
LoadTTP(string, fs.StatFS) *TTP, error
```

LoadTTP reads a TTP file and creates a TTP instance based on its contents.
If the file is empty or contains invalid data, it returns an error.

Parameters:

ttpFilePath: the absolute or relative path to the TTP file.
system: An optional fs.StatFS from which to load the TTP

Returns:

ttp: Pointer to the created TTP instance, or nil if the file is empty or invalid.
err: An error if the file contains invalid data or cannot be read.

---

### NewAct()

```go
NewAct() *Act
```

NewAct is a constructor for the Act struct.

---

### NewBasicStep()

```go
NewBasicStep() *BasicStep
```

NewBasicStep creates a new BasicStep instance, initializing an Act
struct within it.

**Returns:**

*BasicStep: A pointer to a new BasicStep instance.

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

### NewSubTTPStep()

```go
NewSubTTPStep() *SubTTPStep
```

NewSubTTPStep creates a new SubTTPStep and returns a pointer to it.

---

### SubTTPStep.Execute(TTPExecutionContext)

```go
Execute(TTPExecutionContext) error
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

Validate checks the validity of the SubTTPStep by ensuring the following conditions are met:
1. The associated Act is valid.
2. The TTP file associated with the SubTTPStep can be successfully unmarshalled.
3. The TTP file path is not empty.
4. The steps within the TTP file do not contain any nested SubTTPSteps.
If any of these conditions are not met, an error is returned.

---

### TTP.Cleanup(TTPExecutionContext, map[string]Step, []CleanupAct)

```go
Cleanup(TTPExecutionContext, map[string]Step, []CleanupAct) error
```

Cleanup executes all cleanup steps in the TTP.

Parameters:

execCtx: The execution context for the TTP.
availableSteps: A map of available steps.
cleanupSteps: A list of cleanup actions.

Returns:

error: An error if any of the cleanup steps fail to execute.

---

### TTP.MarshalYAML()

```go
MarshalYAML() interface{}, error
```

MarshalYAML is a custom marshalling implementation for the TTP structure.
It encodes a TTP object into a formatted
YAML string, handling the indentation and structure of the output YAML.

Returns:

interface{}: The formatted YAML string representing the TTP object.

error: An error if the encoding process fails.

---

### TTP.RunSteps(TTPExecutionContext)

```go
RunSteps(TTPExecutionContext) error
```

RunSteps executes all steps in the TTP.

Parameters:

execCtx: The execution context for the TTP.

Returns:

error: An error if any of the steps fail to execute.

---

### TTP.UnmarshalYAML(*yaml.Node)

```go
UnmarshalYAML(*yaml.Node) error
```

UnmarshalYAML is a custom unmarshalling implementation for the TTP structure. It decodes a YAML Node into a TTP object,
handling the decoding and validation of the individual steps within the TTP.

Parameters:

node: A pointer to a yaml.Node that represents the TTP structure to be unmarshalled.

Returns:

error: An error if the decoding process fails or if the TTP structure contains invalid steps.

---

### TTP.ValidateSteps(TTPExecutionContext)

```go
ValidateSteps(TTPExecutionContext) error
```

ValidateSteps iterates through the TTP steps and validates each.
The working directory for each step is set before calling its
Validate method.

**Parameters:**

execCtx: The execution context for the TTP.

Returns:

error: An error if any step validation fails, otherwise nil.

---

## Installation

To use the TTPForge/blocks package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/ttpforge/facebookincubator/blocks
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/ttpforge/facebookincubator/blocks"
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
License - see the [LICENSE](../LICENSE)
file for details.
package blocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Constants representing supported executor types.
const (
	ExecutorPython     = "python3"
	ExecutorBash       = "bash"
	ExecutorSh         = "sh"
	ExecutorPowershell = "powershell"
	ExecutorRuby       = "ruby"
	ExecutorBinary     = "binary"
	ExecutorCmd        = "cmd.exe"
)

// StepType represents the type of a TTP (Tactics, Techniques, and Procedures) step.
type StepType string

// Constants representing supported step types.
const (
	StepUnset   = "UNSET"
	StepFile    = "FILESTEP"
	StepBasic   = "BASICSTEP"
	StepSubTTP  = "SUBTTP"
	StepCleanup = "CLEANUP"
)

// Act represents a single action within a TTP (Tactics, Techniques, and Procedures) step. It contains information
// about the condition that must be met for the action to execute, the environment variables that should be set,
// the name of the action, the working directory, the step type, and the success status of the action.
// It also includes a map of step references and a map of output values.
type Act struct {
	Condition   string            `yaml:"if,omitempty"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	Name        string            `yaml:"name"`
	WorkDir     string            `yaml:"-"`
	Type        StepType          `yaml:"-"`
	success     bool
	stepRef     map[string]Step
	output      map[string]any
}

// NewAct is a constructor for the Act struct.
func NewAct() *Act {
	return &Act{
		output: make(map[string]interface{}),
	}
}

// CleanupAct is required for anything that requires a cleanup step.
type CleanupAct interface {
	Cleanup() error
	CleanupName() string
	Setup(env map[string]string, outputRef map[string]Step)
	SetDir(dir string)
	IsNil() bool
	Success() bool
	Validate() error
}

// Step is an interface that represents a TTP step. Types that implement this interface must
// provide methods for setting up the environment and output references, setting the working
// directory, getting the cleanup actions, executing the step, checking if the step is empty,
// explaining validation errors, validating the step, fetching arguments, getting output,
// searching output, setting output success status, checking success status, returning the
// step name, and getting the step type.
type Step interface {
	Setup(env map[string]string, outputRef map[string]Step)
	SetDir(dir string)
	// Need list in case some steps are encapsulating many cleanup steps
	GetCleanup() []CleanupAct
	// Execute will need to take care of the condition checks/etc...
	Execute() error
	IsNil() bool
	ExplainInvalid() error
	Validate() error
	FetchArgs(args []string) []string
	GetOutput() map[string]any
	SearchOutput(arg string) string
	SetOutputSuccess(output *bytes.Buffer, exit int)
	Success() bool
	StepName() string
	GetType() StepType
}

// SetDir sets the working directory for the Act.
//
// Parameters:
//
// dir: A string representing the directory path to be set as the working directory.
func (a *Act) SetDir(dir string) {
	a.WorkDir = dir
}

// IsNil checks if the Act is considered nil (i.e., it has no name). It returns true if the Act has no name, false otherwise.
//
// Returns:
//
// bool: True if the Act has no name, false otherwise.
func (a *Act) IsNil() bool {
	switch {
	case a.Name == "":
		return true
	default:
		return false
	}
}

// ExplainInvalid returns an error explaining why the Act is invalid. If the Act is valid, it returns nil.
//
// Returns:
//
// error: An error explaining why the Act is invalid, or nil if the Act is valid.
func (a *Act) ExplainInvalid() error {
	switch {
	case a.Name == "":
		return errors.New("no name provided for current step")
	default:
		return nil
	}
}

// Cleanup is a placeholder function for the base Act. Subtypes can override this method to implement their own cleanup logic.
//
// Returns:
//
// error: Always returns nil for the base Act.
func (a *Act) Cleanup() error {
	// base act will not do anything, this allows sub types to do what they need
	return nil
}

// StepName returns the name of the Act.
//
// Returns:
//
// string: The name of the Act.
func (a *Act) StepName() string {
	return a.Name
}

// GetOutput returns the output map of the Act.
//
// Returns:
//
// map[string]any: The output map of the Act.
func (a *Act) GetOutput() map[string]any {
	return a.output
}

// Success returns the success status of the Act.
//
// Returns:
//
// bool: The success status of the Act.
func (a *Act) Success() bool {
	return a.success
}

// Validate checks the Act for any validation errors, such as the presence of spaces in the name. It returns an error if any validation errors are found.
//
// Returns:
//
// error: An error if any validation errors are found, or nil if the Act is valid.
func (a *Act) Validate() error {
	// Make sure name is of format we can index
	if strings.Contains(a.Name, " ") {
		return errors.New("name must not contain spaces")
	}

	return nil
}

// FetchArgs processes a slice of arguments and returns a new slice with the output values of referenced steps.
//
// Parameters:
//
// args: A slice of strings representing the arguments to be processed.
//
// Returns:
//
// []string: A slice of strings containing the processed output values of referenced steps.
func (a *Act) FetchArgs(args []string) []string {
	Logger.Sugar().Debug("Fetching args data")
	Logger.Sugar().Debug(a.output)
	var inputs []string
	for _, arg := range args {
		inputs = append(inputs, a.SearchOutput(arg))
	}
	Logger.Sugar().Debugw("full list of inputs", "inputs", inputs)

	return inputs
}

// Setup initializes the Act with the given environment and output reference maps.
//
// Parameters:
//
// env: A map of environment variables, where the keys are variable names and the values are variable values.
// outputRef: A map of output references, where the keys are step names and the values are Step instances.
//
// Returns:
//
// map[string]: Step instances.
func (a *Act) Setup(env map[string]string, outputRef map[string]Step) {
	a.stepRef = outputRef
	a.output = make(map[string]any)

	stepEnv := env
	Logger.Sugar().Debugw("supplied environment", "env", a.Environment)
	// Logger.Sugar().Debugw("supplied environment", "env", env)
	for k, v := range a.Environment {
		valLookup := a.SearchOutput(v)
		stepEnv[k] = valLookup
	}
	a.Environment = stepEnv
}

// SearchOutput searches for the output value of a step by parsing the provided argument. The argument should be in the
// format "steps.step_name.output". It returns the value as a string, converting the value to a string representation if
// necessary. If the argument is not in the correct format, or the step is not found, the original argument is returned.
//
// Parameters:
//
// arg: A string representing the argument in the format "steps.step_name.output".
//
// Returns:
//
// string: The output value of the step as a string, or the original argument if the step is not found or the argument is in an incorrect format.
func (a *Act) SearchOutput(arg string) string {
	Logger.Sugar().Debugw("fetch arg", "arg", arg)
	val, err := a.search(arg)
	if err != nil {
		Logger.Sugar().Debugw("bad arg name", "arg", arg, "err", err)
		return arg
	}
	switch v := val.(type) {
	case string:
		return v
	case int:
		return fmt.Sprint(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		b, err := json.Marshal(val)
		if err != nil {
			Logger.Sugar().Warnw("value improperly parsed, defaulting to arg as string", "val", val, "err", err)
			return arg
		}
		return string(b)
	}
}

func (a *Act) search(arg string) (any, error) {
	if !strings.HasPrefix(arg, "steps.") {
		return nil, errors.New("name is not of format steps.step_name.output")
	}

	steps := strings.SplitN(arg, "steps.", 2)
	splitNames := strings.Split(steps[1], ".")

	if len(splitNames) < 2 {
		return nil, errors.New("invalid argument supplied")
	}

	stepName := splitNames[0]
	outputKeys := splitNames[1:]

	step, ok := a.stepRef[stepName]
	if !ok {
		return nil, errors.New("failed to locate step in args")
	}

	return getOutputValue(step.GetOutput(), outputKeys)
}

func getOutputValue(output map[string]interface{}, keys []string) (any, error) {
	if len(keys) == 0 {
		return nil, errors.New("no output keys provided")
	}

	value := output
	for i, key := range keys {
		v, ok := value[key]
		if !ok {
			return nil, fmt.Errorf("failed to locate output key: %s", strings.Join(keys[:i+1], "."))
		}

		if i == len(keys)-1 {
			return v, nil
		}

		value, ok = v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("output key %s is not a nested structure", strings.Join(keys[:i+1], "."))
		}
	}

	return nil, errors.New("unexpected error while retrieving output value")
}

// CheckCondition checks the condition specified for an Act and returns true if it matches the current OS, false otherwise.
// If the condition is "always", the function returns true. If an error occurs while checking the condition, it is returned.
//
// Returns:
//
// bool: true if the condition matches the current OS or the condition is "always", false otherwise.
// error: An error if an error occurs while checking the condition.
func (a *Act) CheckCondition() (bool, error) {
	switch a.Condition {
	case "windows":
		if runtime.GOOS == "windows" {
			return true, nil
		}
	case "darwin":
		if runtime.GOOS == "darwin" {
			return true, nil
		}
	case "linux":
		if runtime.GOOS == "linux" {
			return true, nil
		}
	// Run even if a previous step has failed.
	case "always":
		return true, nil

	default:
		return false, nil
	}
	return false, nil
}

// SetOutputSuccess sets the output of an Act to a given buffer and sets the success flag to true or false depending on the exit code.
// If the output can be unmarshalled into a JSON structure, it is stored as a string in the Act's output map.
//
// Parameters:
//
// output: A pointer to a bytes.Buffer containing the output to set as the Act's output.
// exit: An integer representing the exit code of the Act.
//
// Returns:
//
// None.
func (a *Act) SetOutputSuccess(output *bytes.Buffer, exit int) {
	a.success = true
	if exit != 0 {
		a.success = false
	}

	outStr := strings.TrimSpace(output.String())
	var jsonOutput map[string]any
	if err := json.Unmarshal(output.Bytes(), &jsonOutput); err != nil {
		Logger.Sugar().Debugw("failed to marshal output into JSON structure", zap.Error(err))
		Logger.Sugar().Infow("treating output as single string", "output", outStr)
		a.output["output"] = outStr
		return
	}

	Logger.Sugar().Debugw("unmarshalled output to JSON", "json", jsonOutput)
	a.output = jsonOutput
}

// MakeCleanupStep creates a CleanupAct based on the given yaml.Node. If the node is empty or invalid, it returns nil.
// If the node contains a BasicStep or FileStep, the corresponding CleanupAct is created and returned.
//
// Parameters:
//
// node: A pointer to a yaml.Node containing the parameters to create the CleanupAct.
//
// Returns:
//
// CleanupAct: The created CleanupAct, or nil if the node is empty or invalid.
// error: An error if the node contains invalid parameters.
func (a *Act) MakeCleanupStep(node *yaml.Node) (CleanupAct, error) {
	// TODO: REFACTOR FOR CLARITY
	// we don't care if cleanup fails so move on.
	if node.IsZero() {
		return nil, nil
	}
	var berr, ferr error

	// we do it piecemiel to build our struct
	basic := NewBasicStep()

	berr = node.Decode(&basic)
	if basic != nil && basic.Name == "" {
		basic.Name = fmt.Sprintf("cleanup-%s", a.Name)
		basic.Type = StepCleanup
	}

	if berr == nil && !basic.IsNil() {
		Logger.Sugar().Debugw("cleanup step found", "basicstep", basic)
		return basic, nil
	}

	file := NewFileStep()
	ferr = node.Decode(&file)
	if file != nil && file.Name == "" {
		file.Name = fmt.Sprintf("cleanup-%s", a.Name)
		file.Type = StepCleanup
	}

	if ferr == nil && !file.IsNil() {
		Logger.Sugar().Debugw("cleanup step found", "filestep", file)
		return file, nil
	}

	err := fmt.Errorf("invalid parameters for cleanup steps with basic [%v], file [%v]", berr, ferr)
	Logger.Sugar().Errorw(err.Error(), zap.Error(err))
	return nil, err
}

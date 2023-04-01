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

const (
	ExecutorPython     = "python3"
	ExecutorBash       = "bash"
	ExecutorSh         = "sh"
	ExecutorPowershell = "powershell"
	ExecutorRuby       = "ruby"
	ExecutorBinary     = "binary"
	ExecutorCmd        = "cmd.exe"
)

type StepType string

const (
	UNSET     = "UNSET"
	FILESTEP  = "FILESTEP"
	BASICSTEP = "BASICSTEP"
	SUBTTP    = "SUBTTP"
	CLEANUP   = "CLEANUP"
)

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

// Anything that needs a cleanup step must implement this method
type CleanupAct interface {
	Cleanup() error
	CleanupName() string
	Setup(env map[string]string, outputRef map[string]Step)
	SetDir(dir string)
	IsNil() bool
	Success() bool
	Validate() error
}

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

func (a *Act) SetDir(dir string) {
	a.WorkDir = dir
}

func (a *Act) IsNil() bool {
	switch {
	case a.Name == "":
		return true
	// case a.Condition != "",
	// 	a.Environment != nil:
	default:
		return false
	}
}

func (a *Act) ExplainInvalid() error {
	switch {
	case a.Name == "":
		return errors.New("no name provided for current step")
	// case a.Condition != "",
	// 	a.Environment != nil:
	default:
		return nil
	}
}

func (a *Act) Cleanup() error {
	// base act will not do anything, this allows sub types to do what they need
	return nil
}

func (a *Act) StepName() string {
	return a.Name
}

func (a *Act) GetOutput() map[string]any {
	return a.output
}

func (a *Act) Success() bool {
	return a.success
}

func (a *Act) Validate() error {

	// Make sure name is of format we can index
	if strings.Contains(a.Name, " ") {
		return errors.New("name must not contain spaces")
	}

	return nil

}

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

func (a *Act) SearchOutput(arg string) string {
	Logger.Sugar().Debugw("fetch arg", "arg", arg)
	val, err := a.search(arg)
	if err != nil {
		Logger.Sugar().Debugw("bad arg name", "arg", arg, "err", err)
		return arg
	}
	switch val.(type) {
	case string:
		return string(val.(string))
	case int:
		return fmt.Sprint(val.(int))
	case bool:
		return string(strconv.FormatBool(val.(bool)))
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
	steps := strings.SplitN(arg, "steps.", 2)
	Logger.Sugar().Debugw("remove steps", "args", steps)
	if len(steps) != 2 {
		return nil, errors.New("name is not of format steps.step_name.output")
	}
	splitNames := strings.Split(steps[1], ".")
	Logger.Sugar().Debugw("split up args", "args", splitNames)
	if len(splitNames) > 2 {
		stepName := splitNames[0]
		Logger.Sugar().Debugw("name of step to fetch", "stepname", stepName)
		outputs := splitNames[2:]
		if val, ok := a.stepRef[stepName]; ok {
			return searchMap("output", val.GetOutput(), outputs)
		}
		return nil, errors.New("failed to locate step in args")
	} else if len(splitNames) == 2 {
		stepName := splitNames[0]
		Logger.Sugar().Debugw("split name of step to fetch", "stepname", stepName)
		Logger.Sugar().Debugw("output", "out", a.stepRef[stepName])
		if val, ok := a.stepRef[stepName]; ok {
			return val.GetOutput(), nil
		}
		return nil, errors.New("failed to locate step in args")

	}

	return nil, errors.New("invalid argument supplied")
}

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
		// run even if failure occurred
	case "always":
		return true, nil

	default:
		return false, nil
	}
	return false, nil
}

func (a *Act) SetOutputSuccess(output *bytes.Buffer, exit int) {
	a.success = true
	if exit != 0 {
		a.success = false
	}

	outStr := strings.TrimSpace(string(output.Bytes()))
	var outJSON map[string]any
	err := yaml.Unmarshal(output.Bytes(), &outJSON)
	if err != nil {
		// TODO - error here: failed to unmarshal output into json structure  {"err": "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `HELLO W...` into map[string]interface {}"}
		Logger.Sugar().Debugw("failed to unmarshal output into json structure", "err", err)
		Logger.Sugar().Infow("Command output: ", "output", outStr)
		a.output["output"] = outStr
		return
	}

	Logger.Sugar().Debugw("json marshalled to JSONOutput", "json", outJSON)
	a.output = outJSON
}

func (a *Act) MakeCleanupStep(node *yaml.Node) (CleanupAct, error) {
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
		basic.Type = CLEANUP
	}

	if berr == nil && !basic.IsNil() {
		Logger.Sugar().Debugw("cleanup step found", "basicstep", basic)
		return basic, nil
	}

	file := NewFileStep()
	ferr = node.Decode(&file)
	if file != nil && file.Name == "" {
		file.Name = fmt.Sprintf("cleanup-%s", a.Name)
		file.Type = CLEANUP
	}

	if ferr == nil && !file.IsNil() {
		Logger.Sugar().Debugw("cleanup step found", "filestep", file)
		return file, nil
	}

	err := fmt.Errorf("invalid parameters for cleanup steps with basic [%v], file [%v]", berr, ferr)
	Logger.Sugar().Errorw(err.Error(), zap.Error(err))
	return nil, err
}

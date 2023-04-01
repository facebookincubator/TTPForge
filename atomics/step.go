package atomics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/facebookincubator/TTP-Runner/blocks"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Executor constants represent the different executor types supported by the Action struct.
const (
	ExecutorPython     = "ExecutorPython3"
	ExecutorBash       = "ExecutorBash"
	ExecutorSh         = "sh"
	ExecutorPowershell = "powershell"
	ExecutorRuby       = "ruby"
	ExecutorBinary     = "binary"
	ExecutorCmd        = "cmd.exe"
)

// Action represents a single action to be performed in a Step.
// It includes information such as the action's name, file path, TTP, executor,
// inline code, condition, working directory, arguments, success status, and environment variables.
type Action struct {
	Name        string            `yaml:"name"`
	FilePath    string            `yaml:"file,omitempty"`
	TTP         string            `yaml:"ttp,omitempty"`
	Executor    string            `yaml:"executor,omitempty"`
	Inline      string            `yaml:"inline,flow"`
	Condition   string            `yaml:"if,omitempty"`
	Chdir       bool              `yaml:"chdir,omitempty"`
	Args        []string          `yaml:"args,omitempty,flow"`
	Success     bool              `yaml:"success,omitempty"`
	Environment map[string]string `yaml:"env,flow,omitempty"`
	JSONOutput  map[string]any
	outputRef   map[string]*Action
	altBaseDir  string
}

// Step represents a single step in a process, consisting of an Action and an optional Cleanup Action.
// The Cleanup Action can be used to revert or clean up any changes made by the main Action.
type Step struct {
	Action  `yaml:",inline"`
	Cleanup *Action `yaml:"cleanup,omitempty"`
}

func (a *Action) setupEnv(env map[string]string) {
	stepEnv := env
	Logger.Sugar().Debugw("Provided environment", "env", a.Environment)
	Logger.Sugar().Debugw("Provided environment", "env", env)
	for k, v := range a.Environment {
		valLookup := a.FetchArg(v)
		stepEnv[k] = valLookup
	}
	a.Environment = stepEnv
}

func (a *Action) setupStep(env map[string]string, outputRef map[string]*Action) {
	a.outputRef = outputRef
	a.JSONOutput = make(map[string]any)
	a.setupEnv(env)
}

func (a *Action) fetchEnv() []string {
	var envSlice []string
	for k, v := range a.Environment {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

func (a *Action) validate() error {
	if a.Inline != "" && a.FilePath != "" {
		err := errors.New("inline and filepath are mutually exclusive")
		Logger.Error(err.Error(), zap.Error(err))
		return err
	} else if a.Inline == "" && a.FilePath == "" {
		err := errors.New("inline or filepath must be provided")
		Logger.Error(err.Error(), zap.Error(err))
		return err
	}

	// Make sure name is formatted in a way that we can index.
	if strings.Contains(a.Name, " ") {
		err := errors.New("name must not contain spaces")
		Logger.Error(err.Error(), zap.Error(err))
		return err
	}

	// Filepath is set so we must check files existence
	if a.FilePath != "" {
		// Update to use blocks.FindFilePath
		foundPath, err := blocks.FindFilePath(a.FilePath, a.altBaseDir, nil)
		if err != nil {
			Logger.Sugar().Error(zap.Error(err))
			return err
		}

		a.FilePath = foundPath
		Logger.Sugar().Debugw("Updated file path using FindFilePath", "filepath", a.FilePath)

		// If no executor is provided, infer it from the file extension
		if a.Executor == "" {
			a.Executor = blocks.InferExecutor(a.FilePath)
			Logger.Sugar().Infow("Executor set via extension", "exec", a.Executor)
		}
	}

	if a.Executor == "" && a.Inline != "" {
		Logger.Sugar().Debug("defaulting to ExecutorBash since executor was not provided")
		a.Executor = "ExecutorBash"
	}

	if a.Executor == ExecutorBinary {
		return nil
	}

	if _, err := exec.LookPath(a.Executor); err != nil {
		Logger.Sugar().Error(err)
		Logger.Sugar().Error(zap.Error(err))
		return err
	}

	Logger.Sugar().Debugw("command found in path", "executor", a.Executor)

	return nil
}

func (a *Action) searchOutput(arg string) (any, error) {
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
		if val, ok := a.outputRef[stepName]; ok {
			return searchMap("output", val.JSONOutput, outputs)
		}
		return nil, errors.New("failed to locate step in args")
	} else if len(splitNames) == 2 {
		stepName := splitNames[0]
		Logger.Sugar().Debugw("split name of step to fetch", "stepname", stepName)
		Logger.Sugar().Debugw("output", "out", a.outputRef[stepName])
		if val, ok := a.outputRef[stepName]; ok {
			return val.JSONOutput["output"], nil
		}
		return nil, errors.New("failed to locate step in args")

	}

	return nil, errors.New("invalid argument supplied")
}

func (a *Action) FetchArg(arg string) string {
	Logger.Sugar().Debugw("fetch arg", "arg", arg)
	val, err := a.searchOutput(arg)
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

func (a *Action) FetchArgs() []string {
	Logger.Sugar().Debug("Fetching args data")
	Logger.Sugar().Debug(a.outputRef)
	var inputs []string
	for _, arg := range a.Args {
		inputs = append(inputs, a.FetchArg(arg))
	}
	Logger.Sugar().Debugw("full list of inputs", "inputs", inputs)

	return inputs
}

func (a *Action) Exec() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Minute)
	defer cancel()
	Logger.Sugar().Debugw("available data", "outputs", a.outputRef)
	Logger.Sugar().Info("========= Executing ==========")
	if a.FilePath != "" {
		err = a.fileExec()
	} else if a.Inline != "" {
		err = a.executeExecutorBashStdin(ctx)
	}
	Logger.Sugar().Info("========= Result ==========")
	return err
}

func (a *Action) fileExec() error {

	// change directory to support scripts that need to be executed from their working directory
	if a.Chdir {
		cwd, err := os.Getwd()
		if err != nil {

			return err
		}
		fileDir := filepath.Dir(a.FilePath)
		os.Chdir(fileDir)
		defer os.Chdir(cwd)
	}

	var cmd *exec.Cmd
	if a.Executor == ExecutorBinary {
		cmd = exec.Command(a.FilePath, a.FetchArgs()...)
	} else {
		args := []string{a.FilePath}
		args = append(args, a.FetchArgs()...)

		Logger.Sugar().Debugw("command line execution:", "exec", a.Executor, "args", args)
		cmd = exec.Command(a.Executor, args...)
	}
	cmd.Env = a.fetchEnv()

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	if err != nil {
		Logger.Sugar().Errorw("bad exit of process", "stdout", outStr, "stderr", errStr, "exit code", cmd.ProcessState.ExitCode())
		return err
	}
	Logger.Sugar().Debugw("output of process", "stdout", outStr, "stderr", errStr, "status", cmd.ProcessState.ExitCode())

	// cleanup actions don't have this
	if a.JSONOutput != nil {
		a.setOutputSuccess(&stdoutBuf, cmd.ProcessState.ExitCode())
	}

	return nil
}

func (a *Action) setOutputSuccess(output *bytes.Buffer, exit int) {
	a.Success = true
	if exit != 0 {
		a.Success = false
	}

	outStr := strings.TrimSpace(string(output.Bytes()))
	var outJSON map[string]any
	err := yaml.Unmarshal(output.Bytes(), &outJSON)
	if err != nil {
		Logger.Sugar().Debugw("failed to unmarshal output into json structure", "err", err)
		Logger.Sugar().Infow("treating output as single string", "output", outStr)
		a.JSONOutput["output"] = outStr
		return
	}

	Logger.Sugar().Debugw("json marshalled to JSONOutput", "json", outJSON)
	a.JSONOutput["output"] = outJSON

}

func (a *Action) CheckCondition() (bool, error) {
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

func (a *Action) executeExecutorBashStdin(ptx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ptx)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.Executor)
	cmd.Env = a.fetchEnv()
	inline := a.Inline
	if a.FetchArgs() != nil {
		inline = fmt.Sprintf("%s %s", a.Inline, strings.Join(a.FetchArgs(), " "))
	}
	input := strings.NewReader(inline)

	cmd.Stdin = input

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	if err != nil {
		Logger.Sugar().Errorw("bad exit of process", "stdout", outStr, "stderr", errStr, "exit code", cmd.ProcessState.ExitCode())
		return err
	}

	Logger.Sugar().Debugw("output of process", "stdout", outStr, "stderr", errStr, "status", cmd.ProcessState.ExitCode())

	// cleanup actions don't have this
	if a.JSONOutput != nil {
		a.setOutputSuccess(&stdoutBuf, cmd.ProcessState.ExitCode())
	}

	return nil
}

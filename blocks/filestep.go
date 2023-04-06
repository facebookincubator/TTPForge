package blocks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type FileStep struct {
	*Act        `yaml:",inline"`
	FilePath    string     `yaml:"file,omitempty"`
	Executor    string     `yaml:"executor,omitempty"`
	CleanupStep CleanupAct `yaml:"cleanup,omitempty,flow"`
	Args        []string   `yaml:"args,omitempty,flow"`
	altBaseDir  string
}

func NewFileStep() *FileStep {
	return &FileStep{
		Act: &Act{
			Type: FILESTEP,
		},
	}
}

func (f *FileStep) UnmarshalYAML(node *yaml.Node) error {

	type FileStepTmpl struct {
		Act         `yaml:",inline"`
		FilePath    string    `yaml:"file,omitempty"`
		Executor    string    `yaml:"executor,omitempty"`
		CleanupStep yaml.Node `yaml:"cleanup,omitempty,flow"`
		Args        []string  `yaml:"args,omitempty,flow"`
		altBaseDir  string
	}

	var tmpl FileStepTmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	f.Act = &tmpl.Act
	f.Args = tmpl.Args
	f.FilePath = tmpl.FilePath
	f.Executor = tmpl.Executor

	if f.IsNil() {
		return f.ExplainInvalid()
	}

	if tmpl.CleanupStep.IsZero() || f.Type == CLEANUP {
		return nil
	}

	Logger.Sugar().Debugw("step", "name", tmpl.Name)
	cleanup, err := f.MakeCleanupStep(&tmpl.CleanupStep)
	Logger.Sugar().Debugw("step", "err", err)
	if err != nil {
		return err
	}

	f.CleanupStep = cleanup

	return nil
}

func (f *FileStep) GetType() StepType {
	return FILESTEP
}

// Method to establish link with Cleanup Interface
// Assumes that type is the cleanup step
// invoked by f.CleanupStep.Cleanup
func (f *FileStep) Cleanup() error {
	return f.Execute()
}

func (f *FileStep) GetCleanup() []CleanupAct {
	if f.CleanupStep != nil {
		f.CleanupStep.SetDir(f.WorkDir)
		return []CleanupAct{f.CleanupStep}
	}
	return []CleanupAct{}
}
func (f *FileStep) CleanupName() string {
	return f.Name
}

func (f *FileStep) ExplainInvalid() error {
	var err error
	if f.FilePath == "" {
		err = fmt.Errorf("(filepath) empty")
	}
	if f.Name != "" && err != nil {
		return fmt.Errorf("[!] invalid filestep: [%s] %w", f.Name, err)
	}
	return err
}

func (f *FileStep) IsNil() bool {
	switch {
	case f.Act.IsNil():
		return true
	case f.FilePath == "":
		return true
	default:
		return false
	}
}

func (f *FileStep) Execute() (err error) {
	Logger.Sugar().Debugw("available data", "outputs", f.output)
	Logger.Sugar().Info("========= Executing ==========")
	if f.FilePath != "" {
		err = f.fileExec()
	}
	Logger.Sugar().Info("========= Result ==========")
	return err
}

func (f *FileStep) fileExec() error {

	var cmd *exec.Cmd
	if f.Executor == BINARY {
		cmd = exec.Command(f.FilePath, f.FetchArgs(f.Args)...)
	} else {
		args := []string{f.FilePath}
		args = append(args, f.FetchArgs(f.Args)...)

		Logger.Sugar().Debugw("command line looks like", "exec", f.Executor, "args", args)
		cmd = exec.Command(f.Executor, args...)
	}
	cmd.Env = FetchEnv(f.Environment)
	cmd.Dir = f.WorkDir
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Run()
	outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
	if err != nil {
		Logger.Sugar().Warnw("bad exit of process", "stdout", outStr, "stderr", errStr, "exit code", cmd.ProcessState.ExitCode())
		return err
	}
	Logger.Sugar().Debugw("output of process", "stdout", outStr, "stderr", errStr, "status", cmd.ProcessState.ExitCode())

	f.SetOutputSuccess(&stdoutBuf, cmd.ProcessState.ExitCode())

	return nil
}

func (f *FileStep) Validate() error {
	err := f.Act.Validate()
	if err != nil {
		return err
	}

	// Validate one or the other is set
	// Check for cases:
	// 		both set
	// 		both unset
	if f.FilePath == "" {
		return errors.New("inline or filepath must be provided")
	}

	// Filepath is set so we must check files existence
	fullpath, err := CheckExist(f.FilePath, f.WorkDir, nil)
	if err != nil {
		return err
	}

	f.FilePath, err = FetchAbs(fullpath, f.WorkDir)
	if err != nil {
		return err
	}
	// Filepath checks
	if f.Executor == "" {
		ext := filepath.Ext(f.FilePath)
		Logger.Sugar().Debugw("file extension inferred", "filepath", f.FilePath, "ext", ext)
		switch ext {
		case ".sh":
			f.Executor = SH
		case ".py":
			f.Executor = PYTHON
		case ".rb":
			f.Executor = RUBY
		case ".pwsh":
			f.Executor = POWERSHELL
		case ".ps1":
			f.Executor = POWERSHELL
		case ".bat":
			f.Executor = CMD
		case "":
			f.Executor = BINARY
		default:
			if runtime.GOOS == "windows" {
				f.Executor = CMD
			} else {
				f.Executor = SH
			}
		}
		Logger.Sugar().Infow("executor set via extension", "exec", f.Executor)
	}

	if f.Executor == "" {
		// TODO: add os handling using the runtime (ezpz)
		Logger.Sugar().Debug("defaulting to bash since executor was not provided")
		f.Executor = "bash"
	}

	if f.Executor == BINARY {
		return nil
	}
	_, err = exec.LookPath(f.Executor)
	if err != nil {
		return err
	}

	if f.CleanupStep != nil {
		err := f.CleanupStep.Validate()
		if err != nil {
			return err
		}
	}
	Logger.Sugar().Debugw("command found in path", "executor", f.Executor)

	return nil
}

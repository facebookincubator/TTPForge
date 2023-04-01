package blocks

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SubTTPStep struct {
	*Act       `yaml:",inline"`
	TtpFile    string    `yaml:"ttp"`
	FileSystem fs.StatFS `yaml:"-,omitempty"`
	// omit because the sub steps will contain the cleanups
	CleanupSteps []CleanupAct `yaml:"-,omitempty"`
	ttp          TTP
}

func NewSubTTPStep() *SubTTPStep {
	return &SubTTPStep{}
}

func (s *SubTTPStep) GetCleanup() []CleanupAct {

	return s.CleanupSteps
}

func (s *SubTTPStep) UnmarshalYAML(node *yaml.Node) error {
	type Subtmpl struct {
		Act     `yaml:",inline"`
		TtpFile string `yaml:"ttp"`
	}
	var substep Subtmpl

	err := node.Decode(&substep)
	if err != nil {
		return err
	}
	Logger.Sugar().Debugw("step found", "substep", substep)

	s.Act = &substep.Act
	s.TtpFile = substep.TtpFile

	return nil
}

func (s *SubTTPStep) UnmarshalSubTTP() error {
	Logger.Sugar().Debugw("parameters used to grab file", "filename", s.TtpFile, "workdir", s.WorkDir)
	fullpath, err := FindFilePath(s.TtpFile, s.WorkDir, s.FileSystem)
	if err != nil {
		return err
	}

	s.TtpFile = fullpath

	err = s.loadSubTTP()
	if err != nil {
		return err
	}

	return nil

}

func (s *SubTTPStep) Execute() error {
	Logger.Sugar().Infof("[*] Executing Sub TTP: %s", s.Name)
	availableSteps := make(map[string]Step)

	for _, step := range s.ttp.Steps {
		stepCopy := step
		stepCopy.Setup(s.Environment, availableSteps)
		Logger.Sugar().Infof("[+] Running current step: %s", step.StepName())
		err := stepCopy.Execute()
		if err != nil {
			return err
		}

		output := stepCopy.GetOutput()

		availableSteps[stepCopy.StepName()] = stepCopy
		s.output[stepCopy.StepName()] = output

		stepClean := stepCopy.GetCleanup()
		if stepClean != nil {
			Logger.Sugar().Debugw("adding cleanup step", "cleanup", stepClean)
			s.CleanupSteps = append(stepCopy.GetCleanup(), s.CleanupSteps...)
		}
		Logger.Sugar().Debugw("available step data", "data", availableSteps[stepCopy.StepName()].GetOutput())
		Logger.Sugar().Infof("[+] Finished running step: %s", stepCopy.StepName())
	}
	Logger.Sugar().Info("Finished execution of sub ttp file")
	return nil
}

func (s *SubTTPStep) loadSubTTP() error {
	ttps, err := LoadTTP(s.TtpFile)
	if err != nil {
		return err
	}
	s.ttp = ttps

	// uses the directory of the yaml file as full path reference
	// s.TtpFile may be a relative dir
	fp, err := FetchAbs(s.TtpFile, s.WorkDir)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fp)

	// run validate to flesh out issues
	Logger.Sugar().Infof("[*] Validating Sub TTP: %s", s.Name)
	for _, step := range s.ttp.Steps {

		stepCopy := step
		// pass in the directory
		stepCopy.SetDir(dir)
		err := stepCopy.Validate()
		if err != nil {
			return err
		}
	}
	Logger.Sugar().Infof("[*] Finished validating Sub TTP")

	return nil
}

func (s *SubTTPStep) GetType() StepType {
	return SUBTTP
}

func (s *SubTTPStep) ExplainInvalid() error {
	var err error
	if s.TtpFile == "" {
		err = fmt.Errorf("(ttpfile) empty")
	}
	if s.Name != "" && err != nil {
		return fmt.Errorf("[!] invalid subttp: [%s] %w", s.Name, err)
	}
	return err
}

func (s *SubTTPStep) IsNil() bool {
	Logger.Sugar().Info(s.Act)
	switch {
	case s.Act.IsNil():
		return true
	case s.TtpFile == "":
		return true
	default:
		return false
	}
}

func (s *SubTTPStep) Validate() error {
	err := s.Act.Validate()
	if err != nil {
		return err
	}
	err = s.UnmarshalSubTTP()
	if err != nil {
		return err
	}

	if s.TtpFile == "" {
		return errors.New("ttp file must be specified")
	}

	// first check if steps contain any sub ttps themselves, if they do, error
	for _, steps := range s.ttp.Steps {
		if steps.GetType() == SUBTTP {
			return errors.New("nested ttp step found in nested ttp step, remove to continue execution")
		}
	}

	return nil
}

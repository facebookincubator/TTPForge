/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func init() {
	// TODO: ensure that a supported template is input
	ttpCmd.Flags().StringVarP(&ttpInput.Template, "template", "t", "", "Template to use for generating the TTP (e.g., bash, python)")
	ttpCmd.Flags().StringVarP(&ttpInput.Path, "path", "p", "", "Path for the generated TTP")
	ttpCmd.Flags().StringVarP(&ttpInput.StepType, "step-type", "s", "", "Step type to create for the generated TTP ('file' or 'inline')")
	ttpCmd.Flags().StringSliceVarP(&ttpInput.Args, "args", "a", []string{}, "Arguments to include in the generated TTP")
	ttpCmd.Flags().StringToStringVarP(&ttpInput.Env, "env", "e", nil, "Environment variables to include in the generated TTP "+
		"in the format KEY=VALUE")
	ttpCmd.Flags().BoolVar(&ttpInput.Cleanup, "cleanup", true, "Include a cleanup step in the generated TTP")

	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(ttpCmd)
}

var ttpInput TTPInput

// TTPInput contains the inputs required to create a new TTP from a template.
type TTPInput struct {
	Template string
	Path     string
	StepType string
	Args     []string
	Cleanup  bool
	Env      map[string]string
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new resource",
	Args:  cobra.NoArgs,
}

var ttpCmd = &cobra.Command{
	Use:   "ttp",
	Short: "Create a new TTP",
	Long: `Create a new TTP using the specified template and path.
For example:

  ./ttpforge -c config.yaml ttp --template bash --step-type file --args --cleanup --path ttps/lateral-movement/ssh/rogue-ssh-key`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		requiredFlags := []string{
			"template",
			"path",
			"step-type",
		}
		return checkRequiredFlags(cmd, requiredFlags)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ttp, err := createTTP()
		if err != nil {
			logging.Logger.Sugar().Errorw("failed to create TTP with:", ttpInput, zap.Error(err))
			cobra.CheckErr(err)
		}
		if _, err := yaml.Marshal(ttp); err != nil {
			logging.Logger.Sugar().Errorw("failed to marshal TTP to YAML", zap.Error(err))
			cobra.CheckErr(err)
		}

		saveTTPToFile(ttp, ttpInput.Path)
	},
}

func createTTP() (*blocks.TTP, error) {
	ttp := &blocks.TTP{
		Name:        "Example TTP",
		Description: "This is an example TTP created based on user input",
		Environment: ttpInput.Env,
	}

	// Create a new step based on user input
	var step blocks.Step

	if ttpInput.StepType == "file" {
		step = blocks.NewFileStep()
		step.(*blocks.FileStep).Act.Name = "Example file step"
		step.(*blocks.FileStep).FilePath = "example_file.sh"
		step.(*blocks.FileStep).Args = ttpInput.Args
	} else {
		step = blocks.NewBasicStep()
		step.(*blocks.BasicStep).Act.Name = "Example basic step"
		step.(*blocks.BasicStep).Inline = "echo 'Hello, World!'"
		step.(*blocks.BasicStep).Args = ttpInput.Args
	}

	if ttpInput.Cleanup {
		cleanupStep := blocks.NewBasicStep()
		cleanupStep.Act.Name = "Cleanup step"
		cleanupStep.Executor = "inline"
		cleanupStep.Inline = "echo 'Cleanup done'"

		if ttpInput.StepType == "file" {
			cleanupFileStep := blocks.NewFileStep()
			cleanupFileStep.Act.Name = "Cleanup step"
			cleanupFileStep.Executor = "file"
			cleanupFileStep.FilePath = "example_cleanup_file.sh"
			cleanupFileStep.CleanupStep = cleanupFileStep
		}

		switch step := step.(type) {
		case *blocks.FileStep:
			step.CleanupStep = cleanupStep
		case *blocks.BasicStep:
			step.CleanupStep = cleanupStep
		}
	}

	// Add the step to the TTP
	ttp.Steps = append(ttp.Steps, step)

	return ttp, nil
}

func saveTTPToFile(ttp *blocks.TTP, path string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create TTP file: %v", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(ttp)
	if err != nil {
		return fmt.Errorf("failed to encode TTP to YAML: %v", err)
	}

	return nil
}

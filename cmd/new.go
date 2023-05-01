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
	"text/template"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/files"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(NewTTPBuilderCmd())
}

var newTTPInput NewTTPInput
var ttpDir string

// NewTTPInput contains the inputs required to create a new TTP from a template.
type NewTTPInput struct {
	Template string
	Path     string
	TTPType  string
	Args     []string
	Cleanup  bool
	Env      map[string]string
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new resource",
	Args:  cobra.NoArgs,
}

// NewTTPBuilderCmd creates a new TTP from a template using the
// provided input to customize it.
func NewTTPBuilderCmd() *cobra.Command {
	newTTPBuilderCmd := &cobra.Command{
		Use:   "ttp",
		Short: "Create a new TTP",
		Long: `Create a new TTP using the specified template and path.
    Examples:

	# Create bash TTP that employs logic from a provided bash script file.
	./ttpforge -c config.yaml new ttp --path ttps/lateral-movement/ssh/rogue-ssh-key.yaml --template bash --ttp-type file --args "arg1,arg2,arg3" --cleanup --env "EXAMPLE_ENV_VAR=example_value"

	# Create bash TTP that employs inline logic provided in the ttp YAML.
	./ttpforge -c config.yaml new ttp --path ttps/lateral-movement/ssh/ssh-master-mode.yaml --template bash --ttp-type basic --cleanup
	`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			requiredFlags := []string{
				"template",
				"path",
				"ttp-type",
			}

			for _, flag := range requiredFlags {
				if err := cmd.MarkFlagRequired(flag); err != nil {
					return err
				}
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			inventoryPaths := viper.GetStringSlice("inventory")

			templatePath := filepath.Join("templates", newTTPInput.Template)
			ttpTemplatePath := filepath.Join(templatePath, fmt.Sprintf("%sTTP.yaml.tmpl", newTTPInput.Template))
			readmeTemplatePath := filepath.Join(templatePath, "README.md.tmpl")
			fileTTPTemplatePath := filepath.Join(templatePath, fmt.Sprintf("%sTTP.%s.tmpl", newTTPInput.Template, newTTPInput.Template))

			// Iterate through inventory paths and find the first matching template
			for _, invPath := range inventoryPaths {
				invPath = files.ExpandHomeDir(invPath)
				exists, err := files.TemplateExists(ttpTemplatePath, []string{invPath})
				cobra.CheckErr(err)

				if exists {
					break
				}
			}

			// Create the filepath for the input TTP if it doesn't already exist.
			ttpDir = filepath.Dir(newTTPInput.Path)
			if err := files.CreateDirIfNotExists(ttpDir); err != nil {
				cobra.CheckErr(err)
			}

			// Create ttp from template
			ttp, err := createTTP()
			if err != nil {
				logging.Logger.Sugar().Errorw("failed to create TTP with:", newTTPInput, zap.Error(err))
				cobra.CheckErr(err)
			}

			tmpl := template.Must(
				template.ParseFiles(newTTPInput.Path))

			yamlF, err := os.Create(newTTPInput.Path)
			cobra.CheckErr(err)
			defer yamlF.Close()

			if err := tmpl.Execute(yamlF, ttp); err != nil {
				cobra.CheckErr(err)
			}

			// Create README from template
			readme := filepath.Join(ttpDir, "README.md")
			readmeF, err := os.Create(readme)
			tmpl = template.Must(
				template.ParseFiles(readmeTemplatePath))

			cobra.CheckErr(err)
			defer readmeF.Close()

			if err := tmpl.Execute(readmeF, ttp); err != nil {
				cobra.CheckErr(err)
			}

			// Create file-based TTP from template (if applicable)
			if newTTPInput.TTPType == "file" {
				if newTTPInput.Template == "bash" {
					tmpl = template.Must(
						template.ParseFiles(fileTTPTemplatePath))

					bashScriptF, err := os.Create(filepath.Join(ttpDir, "bashTTP.bash"))
					cobra.CheckErr(err)
					defer bashScriptF.Close()

					if err := tmpl.Execute(bashScriptF, ttp); err != nil {
						cobra.CheckErr(err)
					}
				}
			}
		},
	}
	newTTPBuilderCmd.Flags().StringVarP(&newTTPInput.Template, "template", "t", "", "Template to use for generating the TTP (e.g., bash, python)")
	newTTPBuilderCmd.Flags().StringVarP(&newTTPInput.Path, "path", "p", "", "Path for the generated TTP")
	newTTPBuilderCmd.Flags().StringVarP(&newTTPInput.TTPType, "ttp-type", "", "", "Type of TTP to create ('file' or 'inline')")
	newTTPBuilderCmd.Flags().StringSliceVarP(&newTTPInput.Args, "args", "a", []string{}, "Arguments to include in the generated TTP")
	newTTPBuilderCmd.Flags().StringToStringVarP(&newTTPInput.Env, "env", "e", nil, "Environment variables to include in the generated TTP "+
		"in the format KEY=VALUE")
	newTTPBuilderCmd.Flags().BoolVar(&newTTPInput.Cleanup, "cleanup", false, "Include a cleanup step in the generated TTP")

	return newTTPBuilderCmd
}

func createTTP() (*blocks.TTP, error) {
	ttp := &blocks.TTP{
		Name:        filepath.Base(newTTPInput.Path),
		Description: "This is an example TTP created based on user input",
		Environment: newTTPInput.Env,
	}

	// Create a new step based on user input
	var step blocks.Step

	if newTTPInput.TTPType == "file" {
		step = blocks.NewFileStep()
		step.(*blocks.FileStep).Act.Name = "example_file_step"
		step.(*blocks.FileStep).FilePath = filepath.Join(ttpDir, fmt.Sprintf("%sTTP.sh", newTTPInput.Template))
		step.(*blocks.FileStep).Args = newTTPInput.Args
	} else {
		step = blocks.NewBasicStep()
		step.(*blocks.BasicStep).Act.Name = "example_basic_step"
		step.(*blocks.BasicStep).Inline = "echo Hello, World"
		step.(*blocks.BasicStep).Args = newTTPInput.Args
	}

	if newTTPInput.Cleanup {
		cleanupStep := blocks.NewBasicStep()
		cleanupStep.Act.Name = "cleanup_step"
		cleanupStep.Inline = "echo 'Cleanup done'"

		if newTTPInput.TTPType == "file" {
			cleanupFileStep := blocks.NewFileStep()
			cleanupFileStep.Act.Name = "cleanup_step"
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

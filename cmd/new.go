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
	"html/template"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/files"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/l50/goutils/v2/sys"
	"github.com/spf13/afero"
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
	Short: "Create a new TTP",
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
		RunE: runNewTTPBuilderCmd,
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

func runNewTTPBuilderCmd(cmd *cobra.Command, args []string) error {
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

	// Check if --args are provided and --ttp-type is not set to 'file'
	if len(newTTPInput.Args) > 0 && newTTPInput.TTPType != "file" {
		err := fmt.Errorf("--args can only be provided if --ttp-type is set to 'file'")
		return err
	}

	inventoryPaths := viper.GetStringSlice("inventory")

	templatePath := filepath.Join("templates", newTTPInput.Template)
	ttpTemplatePath := filepath.Join(templatePath, fmt.Sprintf("%sTTP.yaml.tmpl", newTTPInput.Template))
	readmeTemplatePath := filepath.Join(templatePath, "README.md.tmpl")
	fileTTPTemplatePath := filepath.Join(templatePath, fmt.Sprintf("%sTTP.%s.tmpl", newTTPInput.Template, newTTPInput.Template))

	// Create an afero.Fs instance
	fsys := afero.NewOsFs()

	var absTmplPath string

	// Iterate through inventory paths and find the first matching template
	for _, invPath := range inventoryPaths {
		invPath = sys.ExpandHomeDir(invPath)
		tmplPath, err := files.TemplateExists(fsys, ttpTemplatePath, []string{invPath})
		if err != nil {
			return err
		}

		if tmplPath != "" {
			absTmplPath = tmplPath
			break
		}
	}

	// Create the filepath for the input TTP if it doesn't already exist.
	ttpDir = filepath.Dir(absTmplPath)
	if err := files.CreateDirIfNotExists(fsys, ttpDir); err != nil {
		return err
	}

	// Create ttp from template
	ttp, err := createTTP()
	if err != nil {
		logging.Logger.Sugar().Errorw("failed to create TTP with:", newTTPInput, zap.Error(err))
		return err
	}

	fs := afero.NewOsFs()
	fullTemplatePath, err := files.TemplateExists(fs, ttpTemplatePath, inventoryPaths)
	if err != nil {
		return err
	}

	if fullTemplatePath == "" {
		return fmt.Errorf("input template %s does not exist", ttpTemplatePath)
	}

	if err := createYAMLFile(ttpTemplatePath, ttp); err != nil {
		return err
	}

	if err := createReadmeFile(readmeTemplatePath, ttp); err != nil {
		return err
	}

	if err := createFileBasedTTP(fileTTPTemplatePath, ttp); err != nil {
		return err
	}

	return nil
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
		step.(*blocks.FileStep).FilePath = filepath.Join(ttpDir, fmt.Sprintf("%sTTP.bash", newTTPInput.Template))
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
			cleanupFileStep.FilePath = "example_cleanup_file.bash"
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

// createYAMLFile creates the YAML file for the TTP.
func createYAMLFile(ttpTemplatePath string, ttp *blocks.TTP) error {
	tmpl := template.Must(template.ParseFiles(ttpTemplatePath))
	yamlF, err := os.Create(newTTPInput.Path)
	if err != nil {
		return err
	}
	defer yamlF.Close()

	return tmpl.Execute(yamlF, ttp)
}

// createReadmeFile creates the README file for the TTP.
func createReadmeFile(readmeTemplatePath string, ttp *blocks.TTP) error {
	readme := filepath.Join(ttpDir, "README.md")
	readmeF, err := os.Create(readme)
	if err != nil {
		return err
	}
	defer readmeF.Close()

	tmpl := template.Must(template.ParseFiles(readmeTemplatePath))
	return tmpl.Execute(readmeF, ttp)
}

// createFileBasedTTP creates a file-based TTP if necessary.
func createFileBasedTTP(fileTTPTemplatePath string, ttp *blocks.TTP) error {
	if newTTPInput.TTPType != "file" {
		return nil
	}

	if newTTPInput.Template == "bash" {
		tmpl := template.Must(template.ParseFiles(fileTTPTemplatePath))
		bashScriptF, err := os.Create(filepath.Join(ttpDir, "bashTTP.bash"))
		if err != nil {
			return err
		}
		defer bashScriptF.Close()

		return tmpl.Execute(bashScriptF, ttp)
	}

	return nil
}

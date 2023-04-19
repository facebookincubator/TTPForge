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

	"github.com/spf13/cobra"
)

// TTPInput contains the inputs required to create a new TTP from a template.
type TTPInput struct {
	Template string
	Path     string
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

  ./ttpforge -c config.yaml ttp --template bash --path ttps/lateral-movement/ssh/rogue-ssh-key`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		requiredFlags := []string{
			"template",
			"path",
		}
		return checkRequiredFlags(cmd, requiredFlags)
	},
	Run: func(cmd *cobra.Command, args []string) {
		input := getTTPInput(cmd)
		handleTTP(input)
	},
}

func init() {
	ttpCmd.Flags().String("template", "", "Template to use (e.g., bash, python)")
	ttpCmd.Flags().String("path", "", "Path to create the folder")

	rootCmd.AddCommand(newCmd)
	newCmd.AddCommand(ttpCmd)
}

func getTTPInput(cmd *cobra.Command) *TTPInput {
	template, _ := cmd.Flags().GetString("template")
	path, _ := cmd.Flags().GetString("path")

	return &TTPInput{
		Template: template,
		Path:     path,
	}
}
func handleTTP(input *TTPInput) {

	fmt.Printf("Created TTP using template: %s, path: %s\n", input.Template, input.Path)
}

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

	"github.com/facebookincubator/ttpforge/pkg/parseutils"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// This command is used to test the parsing of a TTP file when creating a diff as part of sandcastle job.
func buildParseYamlCommand(cfg *Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "parse-yaml [repo_name//path/to/ttp]",
		Short: "Parse YAML to convert to TTP structure.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			fs := afero.NewOsFs()

			for _, ttpRef := range args {
				_, path, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
				if err != nil {
					fmt.Printf("Failed to resolve TTP reference %v: %v", ttpRef, err)
					return err
				}
				content, err := afero.ReadFile(fs, path)
				if err != nil {
					fmt.Printf("Error reading TTP ref: %v on path %v with error: %v", ttpRef, path, err)
					return err
				}
				_, err = parseutils.ParseTTP(content, path)
				// Can add field enforcement here in future...
				if err != nil {
					fmt.Printf("Error parsing TTP ref: %v with error: %v\n", ttpRef, err)
					return err
				}
				fmt.Println("Successfully Parsed TTP: ", ttpRef)
			}
			return nil
		},
	}
	return runCmd
}

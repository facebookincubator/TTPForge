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

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/spf13/cobra"
)

func buildRunCommand() *cobra.Command {
	var argsList []string
	var ttpCfg blocks.TTPExecutionConfig
	runCmd := &cobra.Command{
		Use:   "run [repo_name//path/to/ttp]",
		Short: "Run the TTP found in the specified YAML file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			ttpRef := args[0]
			// find the TTP file
			foundRepo, ttpAbsPath, err := Conf.repoCollection.ResolveTTPRef(ttpRef)
			if err != nil {
				return fmt.Errorf("failed to resolve TTP reference %v: %v", ttpRef, err)
			}

			// load TTP and process argument values
			// based on the TTPs argument value specifications
			ttpCfg.Repo = foundRepo

			// Set the --no-cleanup value if provided
			ttpCfg.NoCleanup, err = cmd.Flags().GetBool("no-cleanup")
			if err != nil {
				return fmt.Errorf("failed to process the no-cleanup arg: %v", err)
			}

			ttp, err := blocks.LoadTTP(ttpAbsPath, foundRepo.GetFs(), &ttpCfg, argsList)
			if err != nil {
				return fmt.Errorf("could not load TTP at %v:\n\t%v", ttpAbsPath, err)
			}

			if _, err := ttp.RunSteps(ttpCfg); err != nil {
				return fmt.Errorf("failed to run TTP at %v: %v", ttpAbsPath, err)
			}
			return nil
		},
	}
	runCmd.PersistentFlags().BoolVar(&ttpCfg.NoCleanup, "no-cleanup", false, "Disable cleanup (useful for debugging and daisy-chaining TTPs)")
	runCmd.PersistentFlags().UintVar(&ttpCfg.CleanupDelaySeconds, "cleanup-delay-seconds", 0, "Wait this long after TTP execution before starting cleanup")
	runCmd.Flags().StringArrayVarP(&argsList, "arg", "a", []string{}, "variable input mapping for args to be used in place of inputs defined in each ttp file")

	return runCmd
}

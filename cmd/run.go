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
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/cobra"
)

func buildRunCommand(cfg *Config) *cobra.Command {
	var argsList []string
	var ttpCfg blocks.TTPExecutionConfig
	runCmd := &cobra.Command{
		Use:               "run [repo_name//path/to/ttp]",
		Short:             "Run the TTP found in the specified YAML file",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTTPRef(cfg, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			// capture output for tests if needed
			if cfg.testCfg != nil {
				ttpCfg.Stdout, ttpCfg.Stderr = cfg.testCfg.Stdout, cfg.testCfg.Stderr
			}

			// find the TTP file
			ttpRef := args[0]
			foundRepo, ttpAbsPath, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
			if err != nil {
				return fmt.Errorf("failed to resolve TTP reference %v: %v", ttpRef, err)
			}

			// load TTP and process argument values
			// based on the TTPs argument value specifications
			ttpCfg.Repo = foundRepo

			ttp, execCtx, err := blocks.LoadTTP(ttpAbsPath, foundRepo.GetFs(), &ttpCfg, map[string]string{}, argsList)
			if err != nil {
				return fmt.Errorf("could not load TTP at %v:\n\t%v", ttpAbsPath, err)
			}

			if ttpCfg.DryRun {
				logging.L().Info("Dry-Run Requested - Returning Early")
				return nil
			}

			runErr := ttp.Execute(*execCtx)
			// Run clean up always
			cleanupErr := ttp.RunCleanup(*execCtx)

			if cleanupErr != nil {
				logging.L().Warnf("Failed to run cleanup: %v", cleanupErr)
			}

			if runErr != nil {
				return fmt.Errorf("failed to run TTP at %v: %w", ttpAbsPath, runErr)
			}
			return nil
		},
	}
	runCmd.PersistentFlags().BoolVar(&ttpCfg.DryRun, "dry-run", false, "Parse arguments and validate TTP Contents, but do not actually run the TTP")
	runCmd.PersistentFlags().BoolVar(&ttpCfg.NoCleanup, "no-cleanup", false, "Disable cleanup (useful for debugging and daisy-chaining TTPs)")
	runCmd.PersistentFlags().BoolVar(&ttpCfg.NoChecks, "no-checks", false, "Skip/ignore checks")
	runCmd.PersistentFlags().BoolVar(&ttpCfg.NoProxy, "no-proxy", false, "Ignore proxy settings defined in TTPs")
	runCmd.PersistentFlags().UintVar(&ttpCfg.CleanupDelaySeconds, "cleanup-delay-seconds", 0, "Wait this long after TTP execution before starting cleanup")
	runCmd.Flags().StringArrayVarP(&argsList, "arg", "a", []string{}, "Variable input mapping for args to be used in place of inputs defined in each ttp file")

	return runCmd
}

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
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/backends"
	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/parseutils"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildRunCommand(cfg *Config) *cobra.Command {
	var argsList []string
	var ttpCfg blocks.TTPExecutionConfig
	var ttpUUID string
	runCmd := &cobra.Command{
		Use:               "run [repo_name//path/to/ttp]",
		Short:             "Run the TTP found in the specified YAML file",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeTTPRef(cfg, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			// capture output for tests if needed
			if cfg.testCfg != nil {
				ttpCfg.Stdout, ttpCfg.Stderr = cfg.testCfg.Stdout, cfg.testCfg.Stderr
			}

			var ttpRef string
			var err error

			// Determine TTP reference - either from UUID flag or positional argument
			if ttpUUID != "" {
				// Look up TTP by UUID
				ttpRef, err = findTTPByUUID(cfg.repoCollection, ttpUUID)
				if err != nil {
					return fmt.Errorf("failed to find TTP with UUID %v: %w", ttpUUID, err)
				}
				logging.L().Infof("Found TTP for UUID %s: %s", ttpUUID, ttpRef)
			} else if len(args) > 0 {
				ttpRef = args[0]
			} else {
				return fmt.Errorf("must provide either a TTP reference or --uuid flag")
			}

			// find the TTP file
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

			// Initialize connection pool here so it is shared
			// across both Execute and RunCleanup
			execCtx.ConnPool = backends.NewConnectionPool()
			defer execCtx.ConnPool.CloseAll()

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
	runCmd.Flags().StringVar(&ttpUUID, "uuid", "", "UUID of the TTP to run (will search all repos to find the TTP)")

	return runCmd
}

// findTTPByUUID searches all repositories for a TTP with the given UUID
// and returns its reference path (repo_name//path/to/ttp.yaml)
func findTTPByUUID(rc repos.RepoCollection, targetUUID string) (string, error) {
	// Get all TTPs from all repos
	ttpRefs, err := rc.ListTTPs()
	if err != nil {
		return "", fmt.Errorf("failed to list TTPs: %w", err)
	}

	for _, ttpRef := range ttpRefs {
		repo, ttpAbsPath, err := rc.ResolveTTPRef(ttpRef)
		if err != nil {
			logging.L().Debugf("Failed to resolve TTP ref %s: %v", ttpRef, err)
			continue
		}

		// Read the TTP file using the repo's filesystem
		content, err := afero.ReadFile(repo.GetFs(), ttpAbsPath)
		if err != nil {
			logging.L().Debugf("Failed to read TTP file %s: %v", ttpAbsPath, err)
			continue
		}

		// Use ParseTTP which only parses the preamble (before steps:)
		// This avoids YAML parsing issues with Go templates in the steps section
		ttp, err := parseutils.ParseTTP(content, ttpAbsPath)
		if err != nil {
			logging.L().Debugf("Failed to parse TTP file %s: %v", ttpAbsPath, err)
			continue
		}

		// Compare UUIDs (case-insensitive)
		if strings.EqualFold(ttp.UUID, targetUUID) {
			// Convert absolute path back to reference format
			ref, err := rc.ConvertAbsPathToAbsRef(repo, ttpAbsPath)
			if err != nil {
				return "", fmt.Errorf("found TTP but failed to convert path to ref: %w", err)
			}
			return ref, nil
		}
	}

	return "", fmt.Errorf("no TTP found with UUID: %s", targetUUID)
}

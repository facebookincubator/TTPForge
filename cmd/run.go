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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(RunTTPCmd())
}

// RunTTPCmd runs an input TTP.
func RunTTPCmd() *cobra.Command {
	var argsList []string
	ttpCfg := blocks.TTPExecutionConfig{}
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the TTP found in the specified YAML file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// find the TTP file
			relativeTTPPath := args[0]
			var fullTTPPath string
			var foundRepo repos.Repo
			for _, repo := range Conf.repos {
				foundPath, err := repo.FindTTP(relativeTTPPath)
				if err != nil {
					return fmt.Errorf("failure in TTP search process: %v", err)
				}
				if foundPath != "" {
					fullTTPPath = foundPath
					foundRepo = repo
					break
				}
			}
			if fullTTPPath == "" {
				return errors.New("could not find TTP in any configured repositories")
			}

			// load TTP and process argument values
			// based on the TTPs argument value specifications
			c := blocks.TTPExecutionConfig{
				Repo: foundRepo,
			}
			ttp, err := blocks.LoadTTP(fullTTPPath, nil, &c, argsList)
			if err != nil {
				return fmt.Errorf("could not load TTP at %v: %v", relativeTTPPath, err)
			}

			if _, err := ttp.RunSteps(c); err != nil {
				return fmt.Errorf("failed to run TTP at %v: %v", fullTTPPath, err)
			}
			return nil
		},
	}
	runCmd.PersistentFlags().BoolVar(&ttpCfg.NoCleanup, "no-cleanup", false, "Disable cleanup (useful for debugging)")
	runCmd.PersistentFlags().UintVar(&ttpCfg.CleanupDelaySeconds, "cleanup-delay-seconds", 0, "Wait this long after TTP execution before starting cleanup")
	runCmd.Flags().StringArrayVarP(&argsList, "arg", "a", []string{}, "variable input mapping for args to be used in place of inputs defined in each ttp file")

	return runCmd
}

func findDirectPathToTTP(ttpPath string) (string, error) {
	// if an existing absolute or relative TTP path was specified
	// (meaning repository search is not necessary) then we just use that
	// This would happen for example if you ran `ttpforge run ~/src/myrepo/ttp.yaml`
	fsys := afero.NewOsFs()
	exists, err := afero.Exists(fsys, ttpPath)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	fullPath, err := filepath.Abs(ttpPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

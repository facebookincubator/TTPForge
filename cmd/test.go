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
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/preprocess"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type testCase struct {
	Name   string            `yaml:"name"`
	Args   map[string]string `yaml:"args"`
	DryRun bool              `yaml:"dry_run"`
}

type testSection struct {
	Cases []testCase `yaml:"tests"`
}

func buildTestCommand(cfg *Config) *cobra.Command {
	var timeoutSeconds int
	runCmd := &cobra.Command{
		Use:   "test [repo_name//path/to/ttp]",
		Short: "Test the TTP found in the specified YAML file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			// find the TTP file
			ttpRef := args[0]
			_, ttpAbsPath, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
			if err != nil {
				return fmt.Errorf("failed to resolve TTP reference %v: %w", ttpRef, err)
			}

			// preprocess to separate out the `tests:` section from the `steps:`
			// section and avoid YAML parsing errors associated with template syntax
			contents, err := afero.ReadFile(afero.NewOsFs(), ttpAbsPath)
			if err != nil {
				return fmt.Errorf("failed to read TTP file %v: %w", ttpAbsPath, err)
			}
			preprocessResult, err := preprocess.Parse(contents)
			if err != nil {
				return err
			}

			// load the test cases
			var ts testSection
			err = yaml.Unmarshal(preprocessResult.PreambleBytes, &ts)
			if err != nil {
				return fmt.Errorf("failed to parse `test:` section of TTP file %v: %w", ttpAbsPath, err)
			}

			// look up the path of this binary (ttpforge)
			selfPath, err := os.Executable()
			if err != nil {
				return fmt.Errorf("could not resolve self path (path to current ttpforge binary): %w", err)
			}

			if len(ts.Cases) == 0 {
				logging.L().Warnf("No tests defined in TTP file %v; exiting...", ttpAbsPath)
				return nil
			}

			// run all cases
			logging.DividerThick()
			logging.L().Infof("EXECUTING %v TEST CASE(S)", len(ts.Cases))
			for tcIdx, tc := range ts.Cases {
				logging.DividerThin()
				logging.L().Infof("RUNNING TEST CASE #%d: %q", tcIdx+1, tc.Name)
				logging.DividerThin()
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, selfPath)
				cmd.Args = append(cmd.Args, "run", ttpAbsPath)
				for argName, argVal := range tc.Args {
					cmd.Args = append(cmd.Args, "--arg")
					cmd.Args = append(cmd.Args, argName+"="+argVal)
				}
				if tc.DryRun {
					cmd.Args = append(cmd.Args, "--dry-run")
				}
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("test case %q failed: %w", tc.Name, err)
				}
			}
			logging.DividerThick()
			logging.L().Info("ALL TESTS COMPLETED SUCCESSFULLY!")
			return nil
		},
	}
	runCmd.PersistentFlags().IntVar(&timeoutSeconds, "time-out-seconds", 10, "Timeout allowed for each test case")

	return runCmd
}

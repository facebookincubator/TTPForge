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

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/preprocess"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type testCase struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Args        map[string]string `yaml:"args"`
	DryRun      bool              `yaml:"dry_run"`
}

// We want to verify everything but the steps themselves against
// our schema. The steps will, in turn, be validated
// by the subsequent invocation of the `ttpforge run` command.
type ttpNonStepFields struct {
	blocks.PreambleFields `yaml:",inline"`
	Cases                 []testCase `yaml:"tests"`
}

// Note - this command cannot be unit tested
// because it calls os.Executable() and actually re-executes
// the same binary ("itself", though with a different command)
// as a subprocess
func buildTestCommand(cfg *Config) *cobra.Command {
	var timeoutSeconds int
	runCmd := &cobra.Command{
		Use:   "test [repo_name//path/to/ttp]",
		Short: "Test the TTP found in the specified YAML file.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			for _, ttpRef := range args {
				// find the TTP file
				_, ttpAbsPath, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
				if err != nil {
					return fmt.Errorf("failed to resolve TTP reference %v: %w", ttpRef, err)
				}
				if err := runTestsForTTP(ttpAbsPath, timeoutSeconds); err != nil {
					return fmt.Errorf("test(s) for TTP %v failed: %w", ttpRef, err)
				}
			}
			return nil
		},
	}
	runCmd.PersistentFlags().IntVar(&timeoutSeconds, "time-out-seconds", 10, "Timeout allowed for each test case")

	return runCmd
}

func runTestsForTTP(ttpAbsPath string, timeoutSeconds int) error {
	logging.DividerThick()
	logging.L().Infof("TESTING TTP FILE:")
	logging.L().Info(ttpAbsPath)
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

	// load the test cases and preamble fields - we don't load
	// the steps themselves because that process will
	// be tested when we call `ttpforge run`
	var ttpf ttpNonStepFields
	err = yaml.Unmarshal(preprocessResult.PreambleBytes, &ttpf)
	if err != nil {
		return fmt.Errorf("failed to parse TTP file %v: %w", ttpAbsPath, err)
	}

	// validate as many fields as we can prior to actually
	// invoking `ttpforge run`
	if err := ttpf.PreambleFields.Validate(); err != nil {
		return fmt.Errorf("invalid TTP file %v: %w", ttpAbsPath, err)
	}

	// look up the path of this binary (ttpforge)
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not resolve self path (path to current ttpforge binary): %w", err)
	}

	var testCases []testCase
	if len(ttpf.Cases) == 0 {
		if len(ttpf.ArgSpecs) == 0 {
			// since this TTP doesn't accept arguments, it doesn't need a test case
			// to validate its steps - so we can add an implicit dry run case
			testCases = append(testCases, testCase{
				Name:        "auto_generated_dry_run",
				Description: "Auto-generated dry run test case",
				DryRun:      true,
			})
		} else {
			logging.L().Warnf("No tests defined in TTP file %v; exiting...", ttpAbsPath)
			return nil
		}
	} else {
		testCases = append(testCases, ttpf.Cases...)
	}

	// run all cases
	logging.DividerThick()
	logging.L().Infof("EXECUTING %v TEST CASE(S)", len(testCases))
	for tcIdx, tc := range testCases {
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
	logging.DividerThin()
	logging.L().Info("ALL TESTS COMPLETED SUCCESSFULLY!")
	return nil
}

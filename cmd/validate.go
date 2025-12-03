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

	"github.com/facebookincubator/ttpforge/pkg/validation"
	"github.com/spf13/cobra"
)

func buildValidateCommand(cfg *Config) *cobra.Command {
	var runTests bool
	var timeoutSeconds int

	validateCmd := &cobra.Command{
		Use:   "validate [repo_name//path/to/ttp]",
		Short: "Validate the structure and syntax of a TTP YAML file",
		Long: `Validate performs comprehensive validation on a TTP YAML file
Unlike --dry-run, this command:
- Does not require values to be provided for all arguments
- Does not require OS/platform compatibility
- Performs extensive structural and best practice checks
- Reports errors, warnings, and informational messages

This is useful for CI/CD validation, linting, and checking TTP syntax
without needing to provide all runtime arguments or match platform requirements.`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTTPRef(cfg, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			ttpRef := args[0]
			foundRepo, ttpAbsPath, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
			if err != nil {
				return fmt.Errorf("failed to resolve TTP reference %v: %w", ttpRef, err)
			}

			fmt.Printf("Validating TTP: %s\n", ttpAbsPath)

			result := validation.ValidateTTP(ttpAbsPath, foundRepo.GetFs(), foundRepo)
			result.Print()

			if result.HasErrors() {
				return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
			}

			if runTests {
				if err := runTestsForTTP(ttpAbsPath, timeoutSeconds); err != nil {
					return fmt.Errorf("test(s) for TTP %v failed: %w", ttpRef, err)
				}
			}

			return nil
		},
	}

	validateCmd.Flags().BoolVar(&runTests, "run-tests", false, "Run tests defined in the TTP file after validation")
	validateCmd.Flags().IntVar(&timeoutSeconds, "time-out-seconds", 10, "Timeout allowed for each test case (only used with --run-tests)")

	return validateCmd
}

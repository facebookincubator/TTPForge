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

// BuildRootCommand constructs a fully-initialized root cobra
// command including all flags and sub-commands.
// This function is called from main(), but
// otherwise is principally used for tests.
//
// The cfg parameter is used to control certain aspect of execution
// in unit tests. Note that this should usually just be set to nil,
// and many of the fields you could set may be overwritten when
// cfg.init is subsequently called.
//
// **Parameters:**
//
// cfg: a Config struct used to control certain aspects of execution
//
// **Returns:**
//
// *cobra.Command: The initialized root cobra command
func BuildRootCommand(testCfg *TestConfig) *cobra.Command {
	cfg := &Config{
		testCfg: testCfg,
	}

	// setup root command and flags
	rootCmd := &cobra.Command{
		Use:   "ttpforge",
		Short: "Execute TTPs.",
		Long: `
TTPForge is a Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.CalledAs() != "init" {
				err := cfg.init()
				if err != nil {
					return fmt.Errorf("failed to load TTPForge configuration file: %w", err)
				}
			}
			return nil
		},
		// we will print our own errors with pretty formatting
		SilenceErrors: true,
	}
	// shared flags across commands - mostly logging
	rootCmd.PersistentFlags().StringVarP(&cfg.cfgFile, "config", "c", "", "Config file")
	rootCmd.PersistentFlags().BoolVar(&logConfig.Stacktrace, "stack-trace", false, "Show stacktrace when logging error")
	rootCmd.PersistentFlags().BoolVar(&logConfig.NoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&logConfig.Verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&logConfig.LogFile, "logfile", "l", "", "Enable logging to file.")

	// add sub commands
	rootCmd.AddCommand(buildInitCommand())
	rootCmd.AddCommand(buildCreateCommand())
	rootCmd.AddCommand(buildListCommand(cfg))
	rootCmd.AddCommand(buildEnumCommand(cfg))
	rootCmd.AddCommand(buildShowCommand(cfg))
	rootCmd.AddCommand(buildRunCommand(cfg))
	rootCmd.AddCommand(buildTestCommand(cfg))
	rootCmd.AddCommand(buildInstallCommand(cfg))
	rootCmd.AddCommand(buildRemoveCommand(cfg))
	rootCmd.AddCommand(buildMoveCommand(cfg))
	return rootCmd
}

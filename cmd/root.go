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
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"gopkg.in/yaml.v3"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Execute sets up runtime configuration for the root command
// and adds formatted error handling
func Execute() error {
	rootCmd := BuildRootCommand()
	if err := rootCmd.Execute(); err != nil {
		// we want our own log formatting (for pretty colors)
		// so we don't use cobra.CheckErr
		logging.L().Errorf("failed to run command:\n\t%v", err)
		return err
	}
	return nil
}

// BuildRootCommand constructs a fully-initialized root cobra
// command including all flags and sub-commands.
// This function is principally used for tests.
func BuildRootCommand() *cobra.Command {

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
				return initConfig()
			}
			return nil
		},
		// we will print our own errors with pretty formatting
		SilenceErrors: true,
	}
	// shared flags across commands - mostly logging
	rootCmd.PersistentFlags().StringVarP(&Conf.cfgFile, "config", "c", "", "Config file")
	rootCmd.PersistentFlags().BoolVar(&logConfig.Stacktrace, "stack-trace", false, "Show stacktrace when logging error")
	rootCmd.PersistentFlags().BoolVar(&logConfig.NoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&logConfig.Verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&logConfig.LogFile, "logfile", "l", "", "Enable logging to file.")

	// add sub commands
	rootCmd.AddCommand(buildInitCommand())
	rootCmd.AddCommand(buildListCommand())
	rootCmd.AddCommand(buildShowCommand())
	rootCmd.AddCommand(buildRunCommand())
	rootCmd.AddCommand(buildInstallCommand())
	rootCmd.AddCommand(buildRemoveCommand())
	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// this eventually needs to be de-global'd but for now we'll at east zero it out
	// between test cases
	Conf = &Config{
		cfgFile: Conf.cfgFile,
	}
	// find config file
	if Conf.cfgFile == "" {
		defaultConfigFilePath, err := getDefaultConfigFilePath()
		if err != nil {
			return fmt.Errorf("could not lookup default config file path: %v", err)
		}
		exists, err := afero.Exists(afero.NewOsFs(), defaultConfigFilePath)
		if err != nil {
			return fmt.Errorf("could not check existence of file %v: %v", defaultConfigFilePath, err)
		}
		if exists {
			Conf.cfgFile = defaultConfigFilePath
		} else {
			logging.L().Warn("No config file specified and default configuration file not found!")
			logging.L().Warn("You probably want to run `ttpforge init`!")
			logging.L().Warn("However, if you know what you are doing, then carry on :)")
		}
	}

	// load config file if we found one
	if Conf.cfgFile != "" {
		cfgContents, err := os.ReadFile(Conf.cfgFile)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(cfgContents, Conf); err != nil {
			return err
		}
	}
	var err error
	if Conf.repoCollection, err = Conf.loadRepoCollection(); err != nil {
		return err
	}

	// setup logging
	return logging.InitLog(logConfig)
}

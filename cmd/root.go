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
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"gopkg.in/yaml.v3"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Config stores the variables from the TTPForge global config file
type Config struct {
	RepoSpecs []repos.Spec `yaml:"repos"`

	repoCollection repos.RepoCollection
	cfgFile        string
}

var (
	// Conf refers to the configuration used throughout TTPForge.
	Conf = &Config{}

	logConfig logging.Config
)

// ExecOptions is used to control some high-level behaviors
// like ~/.ttpforge/config.yaml auto-initialization.
// These toggles are passed from main.go so that
// it doesn't happen accidentally in unit tests
type ExecOptions struct {
	AutoInitConfig bool
}

// Execute sets up runtime configuration for the root command
// and adds formatted error handling
func Execute(eo ExecOptions) error {
	autoInitConfig = eo.AutoInitConfig
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
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
	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// find config file
	if Conf.cfgFile == "" {
		if !autoInitConfig {
			err := errors.New("config auto-init disabled: you must specify a config file manually")
			cobra.CheckErr(err)
		}
		// check default config location
		defaultConfigPath, err := ensureDefaultConfig()
		cobra.CheckErr(err)
		Conf.cfgFile = defaultConfigPath
	}

	// load config file
	logging.L().Debugf("Using config file: %s", Conf.cfgFile)
	cfgContents, err := os.ReadFile(Conf.cfgFile)
	cobra.CheckErr(err)
	err = yaml.Unmarshal(cfgContents, Conf)
	cobra.CheckErr(err)

	// expand config-relative paths
	cfgFileAbsPath, err := filepath.Abs(Conf.cfgFile)
	cobra.CheckErr(err)
	cfgDir := filepath.Dir(cfgFileAbsPath)
	fsys := afero.NewOsFs()
	for specIdx, curSpec := range Conf.RepoSpecs {
		Conf.RepoSpecs[specIdx].Path = filepath.Join(cfgDir, curSpec.Path)
	}
	rc, err := repos.NewRepoCollection(fsys, Conf.RepoSpecs, true)
	cobra.CheckErr(err)
	Conf.repoCollection = rc

	err = logging.InitLog(logConfig)
	cobra.CheckErr(err)
}

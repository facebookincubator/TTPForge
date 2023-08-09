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
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config maintains variables used throughout the TTPForge.
type Config struct {
	Verbose       bool     `mapstructure:"verbose"`
	Logfile       string   `mapstructure:"logfile"`
	NoColor       bool     `mapstructure:"nocolor"`
	InventoryPath []string `mapstructure:"inventory"`

	RepoSpecs []repos.Spec `mapstructure:"repos"`

	repos   []repos.Repo
	cfgFile string
}

var (
	// Logger is used to facilitate logging throughout TTPForge.
	Logger *zap.Logger
	// Conf refers to the configuration used throughout TTPForge.
	Conf = &Config{}

	rootCmd = &cobra.Command{
		Use:   "ttpforge",
		Short: "Execute TTPs.",
		Long: `
TTPForge is a Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
	}
)

// ExecOptions is used to control some high-level behaviors
// like ~/.ttpforge/config.yaml auto-initialization.
// These toggles are passed from main.go so that
// it doesn't happen accidentally in unit tests
type ExecOptions struct {
	AutoInitConfig bool
}

// Execute adds child commands to the root
// command and sets flags appropriately.
func Execute(eo ExecOptions) {
	autoInitConfig = eo.AutoInitConfig
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	Logger = logging.Logger
	// cobra.OnInitialize(initConfig)

	// These flags are set using Cobra only, so we populate
	// the Conf.* variables directly reference the unset values
	// in the struct Config above.
	rootCmd.PersistentFlags().StringVarP(&Conf.cfgFile, "config", "c", "", "Config file")
	rootCmd.PersistentFlags().StringArrayVar(&Conf.InventoryPath, "inventory", []string{"."}, "list of paths to search for ttps")
	// Notice here that the values from the command line are not
	// populated in this instance. This is because we are using viper in
	// addition to cobra to manage these values - Cobra will look for these
	// values on the command line. If the values are not present,
	// then Viper uses values found in the config file (if present).
	_ = rootCmd.PersistentFlags().Bool("nocolor", false, "disable colored output")
	_ = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose logging")
	_ = rootCmd.PersistentFlags().StringP("logfile", "l", "", "Enable logging to file.")

	err := viper.BindPFlag("nocolor", rootCmd.PersistentFlags().Lookup("nocolor"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("inventory", rootCmd.PersistentFlags().Lookup("inventory"))
	cobra.CheckErr(err)

	// Errors caught by this are not actioned upon.
	// The decision making for that is documented
	// [here](https://github.com/facebookincubator/TTPForge/issues/55)
	if err := rootCmd.PersistentFlags().Parse(os.Args); err != nil {
		logging.Logger.Sugar().Debugw("failed to parse os args", os.Args, err)
	}

	verbose, err := strconv.ParseBool(rootCmd.PersistentFlags().Lookup("verbose").Value.String())
	cobra.CheckErr(err)

	if verbose {
		logging.ToggleDebug()
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
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
	viper.SetConfigFile(Conf.cfgFile)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logging.Logger.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
	} else {
		cobra.CheckErr(err)
	}

	if err := viper.Unmarshal(Conf, func(config *mapstructure.DecoderConfig) {
		config.IgnoreUntaggedFields = true
	}); err != nil {
		cobra.CheckErr(err)
	}

	// ensure specified TTP repositories are present
	cfgDir := filepath.Dir(Conf.cfgFile)
	fsys := afero.NewOsFs()
	for _, repoSpec := range Conf.RepoSpecs {
		repoSpec.Path = filepath.Join(cfgDir, repoSpec.Path)
		repo, err := repoSpec.Load(fsys)
		Conf.repos = append(Conf.repos, repo)
		cobra.CheckErr(err)
	}

	err := logging.InitLog(Conf.NoColor, Conf.Logfile, Conf.Verbose)
	cobra.CheckErr(err)
	Logger = logging.Logger
}

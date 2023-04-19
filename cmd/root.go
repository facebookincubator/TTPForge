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
	"strconv"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config maintains variables used throughout the TTPForge.
type Config struct {
	Verbose       bool     `mapstructure:"verbose"`
	Logfile       string   `mapstructure:"logfile"`
	NoColor       bool     `mapstructure:"nocolor"`
	InventoryPath []string `mapstructure:"inventory"`
	StackTrace    bool     `mapstructure:"stacktrace"`
	cfgFile       string
	saveConfig    string
}

var (
	// Logger is used to facilitate logging throughout TTPForge.
	Logger *zap.Logger
	conf   = &Config{}

	rootCmd = &cobra.Command{
		Use:   "ttpforge",
		Short: "Execute TTPs.",
		Long: `
TTPForge is a Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if conf.saveConfig != "" {
				// https://github.com/facebookincubator/ttpforge/issues/4
				if err := WriteConfigToFile(conf.saveConfig); err != nil {
					logging.Logger.Error("failed to write config values", zap.Error(err))
				}
			}
		},
	}
)

// WriteConfigToFile writes the configuration data to a YAML file at the specified
// filepath. It uses the yaml.Marshal function to convert the configuration struct
// into YAML format, and then writes the resulting bytes to the file.
//
// This function is a custom alternative to Viper's built-in WriteConfig method
// to provide better control over the formatting of the output YAML file.
//
// Params:
//   - filepath: The path of the file where the configuration data will be saved.
//
// Returns:
//   - error: An error object if any issues occur during the marshaling or file
//     writing process, otherwise nil.
func WriteConfigToFile(filepath string) error {
	yamlBytes, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, yamlBytes, 0644)
}

// Execute adds child commands to the root
// command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	Logger = logging.Logger
	cobra.OnInitialize(initConfig)

	// Create separate Viper instance to handle command-line flags
	// without affecting the config file
	var flagsViper = viper.New()

	// These flags are set using Cobra only, so we populate the conf.* variables directly
	// reference the unset values in the struct Config above.
	rootCmd.PersistentFlags().StringVarP(&conf.cfgFile, "config", "c", "config.yaml", "Config file (default is config.yaml)")
	rootCmd.PersistentFlags().StringVar(&conf.saveConfig, "save-config", "", "Writes values used in execution to the specified location")
	rootCmd.PersistentFlags().BoolVar(&conf.StackTrace, "stacktrace", false, "Show stacktrace when logging error")
	rootCmd.PersistentFlags().StringArrayVar(&conf.InventoryPath, "inventory", []string{"."}, "list of paths to search for ttps")
	// Notice here that the values from the command line are not populated in this instance.
	// This is because we are using viper in addition to cobra to manage these values -
	// Cobra will look for these values on the command line. If the values are not present,
	// then Viper uses values found in the config file (if present).
	_ = rootCmd.PersistentFlags().Bool("nocolor", false, "disable colored output")
	_ = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose logging")
	_ = rootCmd.PersistentFlags().StringP("logfile", "l", "", "Enable logging to file.")

	err := flagsViper.BindPFlag("nocolor", rootCmd.PersistentFlags().Lookup("nocolor"))
	cobra.CheckErr(err)
	err = flagsViper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)
	err = flagsViper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	cobra.CheckErr(err)
	err = flagsViper.BindPFlag("inventory", rootCmd.PersistentFlags().Lookup("inventory"))
	cobra.CheckErr(err)
	err = flagsViper.BindPFlag("stacktrace", rootCmd.PersistentFlags().Lookup("stacktrace"))
	cobra.CheckErr(err)

	verbose, err := strconv.ParseBool(rootCmd.PersistentFlags().Lookup("verbose").Value.String())
	cobra.CheckErr(err)

	if verbose {
		logging.ToggleDebug()
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if conf.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(conf.cfgFile)
	} else {
		// Search config in current directory with name ".cobra" (without extension).
		viper.AddConfigPath(".")
		// Look for config folder
		homedir, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(homedir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logging.Logger.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
	} else {
		cobra.CheckErr(err)
	}

	if err := viper.Unmarshal(conf, func(config *mapstructure.DecoderConfig) {
		config.IgnoreUntaggedFields = true
	}); err != nil {
		cobra.CheckErr(err)
	}

	if err := logging.InitLog(conf.NoColor, conf.Logfile, conf.Verbose, conf.StackTrace); err != nil {
		Logger = logging.Logger
		cobra.CheckErr(err)
	}
}

func getStringFlagOrDefault(cmd *cobra.Command, flag string) *string {
	value, _ := cmd.Flags().GetString(flag)
	if value == "" {
		viperValue := viper.GetString(flag)
		return &viperValue
	}
	return &value
}

func checkRequiredFlags(cmd *cobra.Command, requiredFlags []string) error {
	for _, flag := range requiredFlags {
		value := getStringFlagOrDefault(cmd, flag)
		if *value == "" {
			return fmt.Errorf("required flag '%s' not set", flag)
		}
	}

	return nil
}

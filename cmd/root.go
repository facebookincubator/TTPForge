package cmd

import (
	"os"
	"strconv"

	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config maintains variables used throughout the TTP Forge.
type Config struct {
	Verbose       bool     `mapstructure:"verbose"`
	Logfile       string   `mapstructure:"logfile"`
	NoColor       bool     `mapstructure:"nocolor"`
	InventoryPath []string `mapstructure:"inventory"`
	StackTrace    bool
	cfgFile       string
	saveConfig    string
}

var (
	// Logger is used to facilitate logging throughout TTP Forge.
	Logger       *zap.Logger
	conf         = &Config{}
	deferredInit []func()

	rootCmd = &cobra.Command{
		Use:   "forge",
		Short: "Execute TTPs.",
		Long: `
TTP-Forge

Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if conf.saveConfig != "" {
				viper.SetConfigFile(conf.saveConfig)
			}
			if err := viper.WriteConfig(); err != nil {
				logging.Logger.Error("failed to write config values", zap.Error(err))
			}
		},
	}
)

// Execute adds child commands to the root
// command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	Logger = logging.Logger
	cobra.OnInitialize(initConfig)

	// These flags are set using Cobra only, so we populate the conf.* variables directly
	// reference the unset values in the struct Config above.
	rootCmd.PersistentFlags().StringVarP(&conf.cfgFile, "config", "c", ".config.yaml", "Config file (default is .config.yaml)")
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

	err := viper.BindPFlag("nocolor", rootCmd.PersistentFlags().Lookup("nocolor"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("inventory", rootCmd.PersistentFlags().Lookup("inventory"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("stacktrace", rootCmd.PersistentFlags().Lookup("stacktrace"))
	cobra.CheckErr(err)

	rootCmd.PersistentFlags().Parse(os.Args)

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

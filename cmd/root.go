package cmd

import (
	"strconv"

	"github.com/facebookincubator/TTP-Runner/logging"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config maintains variables used throughout the TTP Forge.
type Config struct {
	Verbose        bool     `mapstructure:"verbose"`
	Logfile        string   `mapstructure:"logfile"`
	NoColor        bool     `mapstructure:"nocolor"`
	TTPSearchPaths []string `mapstructure:"ttp_search_paths"`
	StackTrace     bool
	cfgFile        string
	saveConfig     string
}

var (
	// Logger is used to facilitate logging throughout TTP Forge.
	Logger       *zap.Logger
	conf         = &Config{}
	deferredInit []func()

	rootCmd = &cobra.Command{
		Use:   "ttp-runner",
		Short: "Run TTPs for a Purple Team Engagement",
		Long: `
TTP-Runner

Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if conf.saveConfig != "" {
				viper.SetConfigFile(conf.saveConfig)
				if err := viper.WriteConfig(); err != nil {
					logging.Logger.Error("failed to write config values", zap.Error(err))
				}
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
	// reference the unset values in the struct Config above
	rootCmd.PersistentFlags().StringVarP(&conf.cfgFile, "config", "c", "", "Config file (default is .config.yaml)")
	rootCmd.PersistentFlags().StringVar(&conf.saveConfig, "save-config", "", "Writes values used in execution to the specified location")
	// Notice here that the values from the command line are not populated in this instance
	// this is because we are using viper in addition to cobra to manage these values
	// ie. Cobra will look for these values on the command line, if not present, then
	//     Viper will use the values it can find in the config file, if present
	_ = rootCmd.PersistentFlags().Bool("nocolor", false, "disable colored output")
	_ = rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose logging")
	_ = rootCmd.PersistentFlags().StringP("logfile", "l", "", "Enable logging to file.")

	err := viper.BindPFlag("nocolor", rootCmd.PersistentFlags().Lookup("nocolor"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))
	cobra.CheckErr(err)

	verbose, err := strconv.ParseBool(rootCmd.PersistentFlags().Lookup("verbose").Value.String())
	cobra.CheckErr(err)

	if verbose {
		logging.ToggleDebug()
	}

	for _, funcInit := range deferredInit {
		funcInit()
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Use config file from the flag - default is handled there
	viper.SetConfigFile(conf.cfgFile)
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

package cmd

import (
	"github.com/facebookincubator/TTP-Runner/logging"
	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config maintains variables used throughout the TTP Runner.
type Config struct {
	// to use viper with a structure we use the `mapstructure` to
	// indicate that viper should find the key (in this instance Verbose)
	// using the string "verbose" when looking in the yaml file
	Verbose bool   `mapstructure:"verbose"`
	Logfile string `mapstructure:"logfile"`
	Nocolor bool   `mapstructure:"nocolor"`
	// omit the `mapstructure` to prevent viper from writing
	// values to config and make use of the cobra cli itself
	// see the init() method for the difference in usage
	cfgFile    string
	saveConfig string
}

var (
	conf = &Config{}

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
				err := viper.WriteConfig()
				if err != nil {

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
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		logging.Logger.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(conf, func(config *mapstructure.DecoderConfig) {
		config.IgnoreUntaggedFields = true
	}); err != nil {
		cobra.CheckErr(err)
	}

	if err := logging.InitLog(conf.Nocolor, conf.Logfile, conf.Verbose); err != nil {
		cobra.CheckErr(err)
	}
}

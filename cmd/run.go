package cmd

import (
	"os"
	"path/filepath"

	"github.com/facebookincubator/TTP-Runner/blocks"
	"github.com/facebookincubator/TTP-Runner/pkg/logging"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func executeYAML(yamlFile string) error {
	logging.Logger = Logger

	ttp, err := blocks.LoadTTP(yamlFile)
	if err != nil {
		Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return err
	}

	inventory := viper.GetStringSlice("inventory")
	Logger.Sugar().Debugw("Inventory path gathered", "paths", inventory)

	for _, path := range inventory {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		abs, err := blocks.FetchAbs(path, dir)
		if err != nil {
			return err
		}
		blocks.InventoryPath = append(blocks.InventoryPath, abs)
	}

	if len(inventory) > 0 {
		blocks.InventoryPath = inventory
	} else {
		Logger.Sugar().Warn("No inventory path specified, using current directory")
	}

	// Use the directory of the YAML file as the full path reference.
	dir := filepath.Dir(yamlFile)
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	// Set the working directory for the TTP.
	ttp.WorkDir = dir

	if err := ttp.RunSteps(); err != nil {
		Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return err
	}

	return nil
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the forgery using the file specified in args.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := executeYAML(args[0]); err != nil {
			Logger.Sugar().Errorw("failed to execute TTP", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

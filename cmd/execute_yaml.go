package cmd

import (
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// ExecuteYAML is the top-level TTP execution function
// exported so that we can test it
// the returned TTP is also required to assert against in tests
func ExecuteYAML(yamlFile string) (*blocks.TTP, error) {
	logging.Logger = Logger

	ttp, err := blocks.LoadTTP(yamlFile)
	if err != nil {
		Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	inventory := viper.GetStringSlice("inventory")
	Logger.Sugar().Debugw("Inventory path gathered", "paths", inventory)

	for _, path := range inventory {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		abs, err := blocks.FetchAbs(path, dir)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	// Set the working directory for the TTP.
	ttp.WorkDir = dir

	if err := ttp.RunSteps(); err != nil {
		Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return &ttp, nil
}

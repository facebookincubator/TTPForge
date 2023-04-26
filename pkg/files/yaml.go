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

package files

import (
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
	ttp, err := blocks.LoadTTP(yamlFile)
	if err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	inventory := viper.GetStringSlice("inventory")
	logging.Logger.Sugar().Debugw("Inventory path gathered", "paths", inventory)

	for _, path := range inventory {
		exists, err := PathExistsInInventory(path)
		if err != nil {
			return nil, err
		}
		if exists {
			abs, err := filepath.Abs(path)
			if err != nil {
				return nil, err
			}
			blocks.InventoryPath = append(blocks.InventoryPath, abs)
		}
	}

	if len(inventory) > 0 {
		blocks.InventoryPath = inventory
	} else {
		logging.Logger.Sugar().Warn("No inventory path specified, using current directory")
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
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return &ttp, nil
}

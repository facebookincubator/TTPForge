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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
)

// ExecuteYAML is the top-level function for executing a TTP defined in a YAML file. It is exported for testing purposes,
// and the returned TTP is required for assertion checks in tests.
//
// Parameters:
//
// yamlFile: A string representing the path to the YAML file containing the TTP definition.
// inventoryPaths: A slice of strings representing the inventory paths to search for the TTP.
//
// Returns:
//
// *blocks.TTP: A pointer to a TTP struct containing the executed TTP and its related information.
// error: An error if the TTP execution fails or if the TTP file cannot be found.
//
// Example:
//
// yamlFilePath := "/path/to/your/ttp.yaml"
// inventoryPaths := []string{"/path/to/your/inventory"}
//
// ttp, err := ExecuteYAML(yamlFilePath, inventoryPaths)
//
// if err != nil {
// log.Fatalf("failed to execute TTP: %v", err)
// }
//
// log.Printf("TTP %s executed successfully\n", ttp.Name)
func ExecuteYAML(yamlFile string, c blocks.TTPExecutionConfig) (*blocks.TTP, error) {

	inventoryPaths := c.InventoryPaths
	if len(inventoryPaths) == 0 {
		logging.Logger.Sugar().Warn("No inventory path specified, using current directory")
	}

	// see if the relative path exists in current dir
	var absPathToTTP string
	_, err := os.Stat(yamlFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("unexpected error when checking TTP existence: %v", err)
		}
	} else {
		absPathToTTP, err = filepath.Abs(yamlFile)
		if err != nil {
			return nil, err
		}
	}

	// TTP not found in current dir - search inventoryPath
	if absPathToTTP == "" {
		for _, inventoryPath := range inventoryPaths {
			absInventoryPath, err := blocks.FindFilePath(inventoryPath, filepath.Dir(inventoryPath), nil)
			if err != nil {
				return nil, err
			}

			exists, err := os.Stat(absInventoryPath)
			if err != nil {
				logging.Logger.Sugar().Errorw("failed to locate the path to the TTP", "path", inventoryPath, zap.Error(err))
				return nil, err
			}

			if exists != nil {
				absPathToTTP = absInventoryPath
				break
			}
		}
	}

	ttp, err := blocks.LoadTTP(absPathToTTP)
	if err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	if err := ttp.RunSteps(c); err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return ttp, nil
}

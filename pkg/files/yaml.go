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
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
)

// ExecuteYAML is the top-level TTP execution function
// exported so that we can test it
// the returned TTP is also required to assert against in tests
func ExecuteYAML(yamlFile string, inventoryPaths []string) (*blocks.TTP, error) {
	ttp, err := blocks.LoadTTP(yamlFile)
	if err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	if len(inventoryPaths) == 0 {
		logging.Logger.Sugar().Warn("No inventory path specified, using current directory")
	}

	var ttpDir string
	for _, inventoryPath := range inventoryPaths {
		// Remove ttp from invPath and remove the specific ttp YAML file from yamlFile.
		// Then bring them both together to get the directory of the TTP to run.
		ttpd := filepath.Join(filepath.Dir(inventoryPath), filepath.Dir(yamlFile))
		exists, err := os.Stat(ttpd)
		if err != nil {
			logging.Logger.Sugar().Errorw("failed to locate the path to the TTP", "path", inventoryPath, zap.Error(err))
			return nil, err
		}

		if exists != nil {
			ttpDir = ttpd
			break
		}
	}

	if ttpDir == "" {
		return nil, fmt.Errorf("failed to find the TTP directory")
	}

	// Set the working directory for the TTP.
	ttp.WorkDir = ttpDir

	if err := ttp.RunSteps(); err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return &ttp, nil
}

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
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type yamlRunner struct {
	NoCleanup bool
	WorkDir   string
}

// ExecuteYAML is the top-level TTP execution function
// exported so that we can test it
// the returned TTP is also required to assert against in tests
func (r *yamlRunner) ExecuteYAML(yamlFile string) (*blocks.TTP, error) {
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

	// handle temporary workdir creation if needed
	if r.WorkDir == "" {
		// no pattern, that would be an indicator of compromise :P
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			return nil, err
		}

		// NoCleanup means we want to save TTP results, don't delete them
		if !r.NoCleanup {
			defer func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					Logger.Sugar().Errorf("Failed to delete workdir: %v", tmpDir)
				}
			}()
		}
		r.WorkDir = tmpDir
	}

	// initialize the working directory for execution
	err = r.copyTTPtoWorkDir(yamlFile)
	if err != nil {
		return nil, err
	}

	// actually enter the directory for execution
	origDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(r.WorkDir)
	if err != nil {
		return nil, err
	}
	defer os.Chdir(origDir)

	if err := ttp.RunSteps(r.NoCleanup); err != nil {
		Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return &ttp, nil
}

// this function will populate r.WorkDir with a temporary
// working directory if it was not already populated
// returns a cleanup function to run if needed
func (r *yamlRunner) copyTTPtoWorkDir(yamlFile string) error {

	dir := filepath.Dir(yamlFile)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// seriously you need a library for this wtf golang
	err = cp.Copy(absDir, r.WorkDir)
	return err
}

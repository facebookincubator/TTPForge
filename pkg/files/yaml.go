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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
)

// ExecuteYAML is the top-level function for executing a TTP defined
// in a YAML file. It is exported for testing purposes,
// and the returned TTP is required for assertion checks in tests.
//
// **Parameters:**
//
// yamlFile: A string representing the path to the YAML file containing
// the TTP definition.
// inventoryPaths: A slice of strings representing the inventory paths
// to search for the TTP.
//
// **Returns:**
//
// *blocks.TTP: A pointer to a TTP struct containing the executed TTP
// and its related information.
//
// error: An error if the TTP execution fails or if the TTP file cannot be found.
func ExecuteYAML(yamlFile string, c blocks.TTPExecutionContext) (*blocks.TTP, error) {

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

	ttp, err := blocks.LoadTTP(absPathToTTP, nil)
	if err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	// Parse CLI inputs.
	for _, arg := range c.CliInputs {
		keyVal := strings.Split(arg, "=")
		if len(keyVal) != 2 {
			log.Fatalf("Invalid argument format for '%s'. Expected format is 'key=value'.", arg)
		}
		ttp.InputMap[keyVal[0]] = keyVal[1]
	}

	if err := ttp.RunSteps(c); err != nil {
		logging.Logger.Sugar().Errorw("failed to run TTP", zap.Error(err))
		return nil, err
	}

	return ttp, nil
}

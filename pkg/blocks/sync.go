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

package blocks

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileSystem is a global fs.FS instance used for file system operations.
var FileSystem fs.FS

// LoadTTP reads a TTP file and creates a TTP instance based on its contents.
// If the file is empty or contains invalid data, it returns an error.
//
// Parameters:
//
// ttpFilePath: the absolute or relative path to the TTP file.
//
// Returns:
//
// ttp: Pointer to the created TTP instance, or nil if the file is empty or invalid.
// err: An error if the file contains invalid data or cannot be read.
func LoadTTP(ttpFilePath string) (*TTP, error) {
	var ttp TTP
	file, err := os.Open(ttpFilePath)
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(contents, &ttp)
	if err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(ttpFilePath)
	if err != nil {
		return nil, err
	}
	ttp.WorkDir = filepath.Dir(absPath)

	// TODO: refactor directory handling - this is in-elegant
	// but has less bugs than previous way
	for _, step := range ttp.Steps {
		step.SetDir(ttp.WorkDir)
		if cleanups := step.GetCleanup(); cleanups != nil {
			for _, c := range cleanups {
				c.SetDir(ttp.WorkDir)
			}
		}
	}

	return &ttp, nil
}

// Actors represents the various actors involved in the execution of TTPs.
// It contains Outputs and Environment, which are maps for storing output
// data and environment variables.
type Actors struct {
	Outputs     map[string]any    // Outputs stores the results generated during the execution of TTPs.
	Environment map[string]string // Environment stores the environment variables required for TTPs execution.
}

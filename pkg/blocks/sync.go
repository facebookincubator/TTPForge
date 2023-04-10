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

	"gopkg.in/yaml.v3"
)

// FileSystem is a global fs.FS instance used for file system operations.
var FileSystem fs.FS

// LoadTTP reads a TTP file and creates a TTP instance based on its contents.
// If the file is empty or contains invalid data, it returns an error.
//
// Parameters:
//
// filename: A string representing the path to the TTP file.
//
// Returns:
//
// loadedTTPs: The created TTP instance, or an empty TTP if the file is empty or invalid.
// err: An error if the file contains invalid data or cannot be read.
func LoadTTP(filename string) (loadedTTPs TTP, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &loadedTTPs)
	if err != nil {
		return
	}

	return
}

// Actors represents the various actors involved in the execution of TTPs.
// It contains Outputs and Environment, which are maps for storing output
// data and environment variables.
type Actors struct {
	Outputs     map[string]any    // Outputs stores the results generated during the execution of TTPs.
	Environment map[string]string // Environment stores the environment variables required for TTPs execution.
}

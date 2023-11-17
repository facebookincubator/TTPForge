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

package checks

import (
	"fmt"

	"github.com/spf13/afero"
)

// PathExists is a condition that verifies that a file exists at a given path
// It can also verify the contents of the file against a checksum
type PathExists struct {
	Path     string    `yaml:"path_exists"`
	Checksum *Checksum `yaml:"checksum"`
}

// Verify checks the condition and returns an error if it fails
func (c *PathExists) Verify(ctx VerificationContext) error {
	fsys := ctx.FileSystem

	// basic existence check
	exists, err := afero.Exists(fsys, c.Path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %q does not exist", c.Path)
	}

	// verify the checksum if provided
	if c.Checksum != nil {
		contentBytes, err := afero.ReadFile(fsys, c.Path)
		if err != nil {
			return err
		}
		return c.Checksum.Verify(contentBytes)
	}
	return nil
}

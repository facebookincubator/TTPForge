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
	"os"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// PathExists is a condition that verifies that a file exists at a given path
// It can also verify:
// - File contents against a checksum
// - File contains/doesn't contain specific text
// - File permissions match expected values
type PathExists struct {
	Path               string    `yaml:"path_exists"`
	Checksum           *Checksum `yaml:"checksum,omitempty"`
	ContentContains    string    `yaml:"content_contains,omitempty"`
	ContentNotContains string    `yaml:"content_not_contains,omitempty"`
	ContentRegex       string    `yaml:"content_regex,omitempty"`
	Permissions        string    `yaml:"permissions,omitempty"`
}

// IsNil checks if the condition is empty or uninitialized
func (c *PathExists) IsNil() bool {
	return c.Path == ""
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

	// read file content once if needed for multiple checks
	var contentStr string
	needsContent := c.Checksum != nil || c.ContentContains != "" ||
		c.ContentNotContains != "" || c.ContentRegex != ""

	if needsContent {
		contentBytes, err := afero.ReadFile(fsys, c.Path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", c.Path, err)
		}
		contentStr = string(contentBytes)

		// verify the checksum if provided
		if c.Checksum != nil {
			if err := c.Checksum.Verify(contentBytes); err != nil {
				return err
			}
		}

		// check if content contains expected string
		if c.ContentContains != "" {
			if !strings.Contains(contentStr, c.ContentContains) {
				return fmt.Errorf("file %q does not contain %q",
					c.Path, c.ContentContains)
			}
		}

		// check if content does not contain specified string
		if c.ContentNotContains != "" {
			if strings.Contains(contentStr, c.ContentNotContains) {
				return fmt.Errorf("file %q contains %q but should not",
					c.Path, c.ContentNotContains)
			}
		}

		// check if content matches regex pattern
		if c.ContentRegex != "" {
			matched, err := regexp.MatchString(c.ContentRegex, contentStr)
			if err != nil {
				return fmt.Errorf("invalid regex pattern %q: %w",
					c.ContentRegex, err)
			}
			if !matched {
				return fmt.Errorf("file %q content does not match regex %q",
					c.Path, c.ContentRegex)
			}
		}
	}

	// check file permissions if specified
	if c.Permissions != "" {
		info, err := fsys.Stat(c.Path)
		if err != nil {
			return fmt.Errorf("failed to stat file %q: %w", c.Path, err)
		}

		actualPerm := info.Mode().Perm()
		var expectedPerm os.FileMode

		// Parse permissions string (e.g., "0755", "0644")
		_, err = fmt.Sscanf(c.Permissions, "%o", &expectedPerm)
		if err != nil {
			return fmt.Errorf("invalid permissions format %q: %w",
				c.Permissions, err)
		}

		if actualPerm != expectedPerm {
			return fmt.Errorf("file %q has permissions %04o, expected %04o",
				c.Path, actualPerm, expectedPerm)
		}
	}

	return nil
}

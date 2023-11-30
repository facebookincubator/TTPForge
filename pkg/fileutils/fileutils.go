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

package fileutils

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandTilde expands a tilde to the user's home directory
func ExpandTilde(path string) (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homedir, path[2:]), nil
	}
	return path, nil
}

// AbsPath is a thin wrapper around filepath.Abs
// that we use because it also calls ExpandTilde
func AbsPath(path string) (string, error) {
	tmp, err := ExpandTilde(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(tmp)
}

// IsAbs is a thin wrapper around filepath.IsAbs
// that we use because it also calls ExpandTilde
func IsAbs(path string) (bool, error) {
	tmp, err := ExpandTilde(path)
	if err != nil {
		return false, err
	}
	return filepath.IsAbs(tmp), nil
}

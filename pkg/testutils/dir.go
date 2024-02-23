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

package testutils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-cmp/cmp"
)

// MakeTempTestDir is used to populate a temporary directory
// with the file/directory contents specified in filesMap.
// It is useful for setting up temporary test environments
//
// **Parameters:**
//
// filesMap: A pointer to a TTPExecutionConfig that represents the execution configuration for the TTP.
//
// **Returns:**
//
// * the created temporary directory path
// * an error if any part of the process failed
func MakeTempTestDir(filesMap map[string][]byte) (string, error) {
	tempDir, err := os.MkdirTemp("", "ttpforge-testing")
	if err != nil {
		return "", err
	}
	for relPath, contents := range filesMap {
		if filepath.IsAbs(relPath) {
			return "", fmt.Errorf("cannot process path %v: this function does not support absolute paths", relPath)
		}
		path := filepath.Join(tempDir, relPath)
		dirPath := filepath.Dir(path)
		err := os.MkdirAll(dirPath, 0700)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(path, contents, 0644)
		if err != nil {
			return "", err
		}
	}
	return tempDir, nil
}

// AreDirsEqual recursively compares two directories for equality
// NOTE: filepath.Wale guarantees lexical order traversal (elements already ordered) thus allowing us to use a slice
// as opposed to a map.
func AreDirsEqual(source string, dest string) (bool, error) {
	var files1, files2 []string
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files1 = append(files1, string(content))
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	err = filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files2 = append(files2, string(content))
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return cmp.Equal(files1, files2), nil
}

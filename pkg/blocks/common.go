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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
)

// FetchAbs returns the absolute path of a file given its path and the
// working directory. It handles cases where the path starts with "~/",
// is an absolute path, or is a relative path from the working directory.
// It logs any errors and returns them.
//
// **Parameters:**
//
// path: A string representing the path to the file.
//
// workdir: A string representing the working directory.
//
// **Returns:**
//
// fullpath: A string representing the absolute path to the file.
//
// error: An error if the path cannot be resolved to an absolute path.
func FetchAbs(path string, workdir string) (fullpath string, err error) {
	if path == "" {
		err = errors.New("empty path provided")
		logging.L().Errorw("failed to get fullpath", zap.Error(err))
		return path, err
	}

	var basePath string
	switch {
	case strings.HasPrefix(path, "~/"):
		basePath, err = os.UserHomeDir()
		if err != nil {
			logging.L().Errorw("failed to get home dir", zap.Error(err))
			return path, err
		}
		path = path[2:]
	case filepath.IsAbs(path):
		basePath = ""
	default:
		basePath = workdir

		// Remove the common prefix between path and workdir, if any
		relPath, err := filepath.Rel(workdir, path)
		if err == nil {
			path = relPath
		}
	}

	fullpath, err = filepath.Abs(filepath.Join(basePath, path))
	if err != nil {
		logging.L().Errorw("failed to get fullpath", zap.Error(err))
		return path, err
	}

	logging.L().Debugw("Full path: ", "fullpath", fullpath)
	return fullpath, nil
}

// FindFilePath checks if a file exists given its path, the working directory,
// and an optional fs.StatFS. It handles cases where the path starts with "../",
// "~/", or is a relative path. It also checks a list of paths in InventoryPath
// for the file. It logs any errors and returns them.
//
// **Parameters:**
//
// path: A string representing the path to the file.
//
// workdir: A string representing the working directory.
//
// system: An optional fs.StatFS that can be used to check if the file exists.
//
// **Returns:**
//
// string: A string representing the path to the file, or an empty string
// if the file does not exist.
//
// error: An error if the file cannot be found or if other errors occur.
func FindFilePath(path string, workdir string, system fs.StatFS) (string, error) {
	logging.L().Debugw("Attempting to find file path", "path", path, "workdir", workdir)

	// Check if file exists using provided fs.StatFS
	if system != nil {
		fsPath := filepath.Join(workdir, path)
		if _, err := system.Stat(fsPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				logging.L().Errorw("file not found using provided fs.StatFS", "path", path, zap.Error(err))
				return "", err
			}
			logging.L().Errorw("error checking provided fs.StatFS for file existence", "path", path, zap.Error(err))
			return "", err
		}
		logging.L().Debugw("File found using provided fs.StatFS", "path", path)
		return fsPath, nil

	}

	// Handle home directory representation in Windows.
	if strings.HasPrefix(path, "~/") || (runtime.GOOS == "windows" && strings.HasPrefix(path, "%USERPROFILE%")) {
		if runtime.GOOS == "windows" {
			path = strings.Replace(path, "%USERPROFILE%", "~", 1)
		}
	} else {
		// Convert path to lowercase for Windows, which employs
		// case-insensitive file systems.
		if runtime.GOOS == "windows" {
			path = strings.ToLower(path)
		}
	}

	// Resolve the input path to an absolute path
	absPath, err := FetchAbs(path, workdir)
	if err != nil {
		logging.L().Errorw("failed to fetch absolute path", "path", path, "workdir", workdir, zap.Error(err))
		return "", err
	}

	// Check if the absolute path exists
	if _, err := os.Stat(absPath); !errors.Is(err, fs.ErrNotExist) {
		logging.L().Debugw("File found in absolute path", "absPath", absPath)
		return absPath, nil
	}

	// If the file is not found in any of the locations, return an error
	err = fmt.Errorf("invalid path %s provided", path)
	logging.L().Errorw("file not found in any location", "path", path, zap.Error(err))
	return "", err
}

// FetchEnv converts an environment variable map into a slice of strings that
// can be used as an argument when running a command.
//
// **Parameters:**
//
// environ: A map of environment variable names to values.
//
// **Returns:**
//
// []string: A slice of strings representing the environment variables
// and their values.
func FetchEnv(environ map[string]string) []string {
	var envSlice []string

	for k, v := range environ {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	return envSlice
}

// Contains checks if a key exists in a map.
//
// **Parameters:**
//
// key: A string representing the key to search for.
//
// search: A map of keys and values.
//
// **Returns:**
//
// bool: A boolean value indicating if the key was found in the map.
func Contains(key string, search map[string]any) bool {
	if _, ok := search[key]; ok {
		return true
	}

	return false
}

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

// ResolveSymlinksInPath resolves symlinks in the existing portion of a path
// without requiring the entire path to exist
func ResolveSymlinksInPath(absPath string) (string, error) {
	// Find the deepest existing directory by walking up the path
	currentPath := absPath
	var nonExistingParts []string

	for {
		// Check if current path exists
		if _, err := os.Stat(currentPath); err == nil {
			// Path exists, try to resolve symlinks
			resolved, err := filepath.EvalSymlinks(currentPath)
			if err != nil {
				// If symlink resolution fails, use the original path
				resolved = currentPath
			}

			// Rebuild the full path with resolved base and non-existing parts
			for i := len(nonExistingParts) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, nonExistingParts[i])
			}
			return resolved, nil
		}

		// Path doesn't exist, move up one level
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// We've reached the root, can't go further up
			return absPath, nil
		}

		nonExistingParts = append(nonExistingParts, filepath.Base(currentPath))
		currentPath = parent
	}
}

// ExpandPath expands tilde and environment variables in a path
func ExpandPath(path string) (string, error) {
	// First expand tilde
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homedir, path[2:])
	}

	// Then expand environment variables
	return os.ExpandEnv(path), nil
}

// containsShellVariable checks if a path contains shell variable syntax
func ContainsShellVariable(path string) bool {
	// Check for bash/powershell style: $VAR or ${VAR}
	if strings.Contains(path, "$") {
		return true
	}
	// Check for Windows cmd style: %VAR%
	if strings.Contains(path, "%") {
		return true
	}
	return false
}

// AbsPath converts a path to an absolute path and resolves symlinks.
// Expands tilde and environment variables before conversion.
func AbsPath(path string) (string, error) {
	expanded, err := ExpandPath(path)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}

	// Then resolve symlinks
	return ResolveSymlinksInPath(absPath)
}

// IsAbs is a thin wrapper around filepath.IsAbs
// that we use because it also calls ExpandPath
func IsAbs(path string) (bool, error) {
	tmp, err := ExpandPath(path)
	if err != nil {
		return false, err
	}
	return filepath.IsAbs(tmp), nil
}

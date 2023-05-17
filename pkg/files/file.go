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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/spf13/afero"
)

// CreateDirIfNotExists checks if a directory exists at the given path and creates it if it does not exist.
// It returns an error if the directory could not be created.
//
// Parameters:
//
// fsys: An afero.Fs object representing the file system to operate on.
// path: A string representing the path to the directory to check and create if necessary.
//
// Returns:
//
// error: An error if the directory could not be created.
//
// Example:
//
// fsys := afero.NewOsFs()
// dirPath := "path/to/directory"
// err := CreateDirIfNotExists(fsys, dirPath)
//
//	if err != nil {
//	    log.Fatalf("failed to create directory: %v", err)
//	}
func CreateDirIfNotExists(fsys afero.Fs, path string) error {
	_, err := fsys.Stat(path)

	switch {
	case os.IsNotExist(err):
		if err := fsys.MkdirAll(path, 0755); err != nil {
			return err
		}
	case err == nil:
		isDir, err := afero.IsDir(fsys, path)
		if err != nil {
			return err
		}
		if !isDir {
			return fmt.Errorf("%s is a file, not a directory", path)
		}
	default:
		return err
	}

	return nil
}

// ExpandHomeDir expands the tilde character in a path to the user's home directory.
// The function takes a string representing a path and checks if the first character is a tilde (~).
// If it is, the function replaces the tilde with the user's home directory. The path is returned
// unchanged if it does not start with a tilde or if there's an error retrieving the user's home
// directory.
//
// Borrowed from https://github.com/l50/goutils/blob/e91b7c4e18e23c53e35d04fa7961a5a14ca8ef39/fileutils.go#L283-L318
//
// Parameters:
//
// path: The string containing a path that may start with a tilde (~) character.
//
// Returns:
//
// string: The expanded path with the tilde replaced by the user's home directory, or the
//
//	original path if it does not start with a tilde or there's an error retrieving
//	the user's home directory.
//
// Example:
//
// pathWithTilde := "~/Documents/myfile.txt"
// expandedPath := ExpandHomeDir(pathWithTilde)
// log.Printf("Expanded path: %s", expandedPath)
func ExpandHomeDir(path string) string {
	path = strings.TrimSpace(path)
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(homeDir, path[1:])
}

// PathExistsInInventory checks if a relative file path exists in any of the inventory directories specified in the
// inventoryPaths parameter. The function uses afero.Fs to operate on a filesystem.
//
// Parameters:
//
// fsys: An afero.Fs object representing the filesystem to search.
// relPath: A string representing the relative path of the file to search for in the inventory directories.
// inventoryPaths: A []string containing the inventory directory paths to search.
//
// Returns:
//
// bool: A boolean value indicating whether the file exists in any of the inventory directories (true) or not (false).
// error: An error if there is an issue checking the file's existence.
//
// Example:
//
// fsys := afero.NewOsFs()
// relFilePath := "templates/exampleTTP.yaml.tmpl"
// inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}
// exists, err := PathExistsInInventory(fsys, relFilePath, inventoryPaths)
//
// if err != nil {
// log.Fatalf("failed to check file existence: %v", err)
// }
//
// if exists {
// log.Printf("File %s found in the inventory directories\n", relFilePath)
// } else {
//
// log.Printf("File %s not found in the inventory directories\n", relFilePath)
// }
func PathExistsInInventory(fsys afero.Fs, relPath string, inventoryPaths []string) (bool, error) {
	for _, inventoryPath := range inventoryPaths {
		inventoryPath = ExpandHomeDir(inventoryPath)
		absPath := filepath.Join(inventoryPath, relPath)

		if _, err := fsys.Stat(absPath); err == nil {
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}

	return false, nil
}

// TemplateExists checks if a template file exists in a 'templates' folder located in the parent directory of any of the
// inventory directories specified in the inventoryPaths parameter. If the template file is found, it returns the full path
// to the template file, otherwise, it returns an empty string.
//
// Parameters:
//
// fsys: An afero.Fs object representing the file system to operate on.
// templatePath: A string representing the path of the template file to search for in the 'templates' folder of the parent
//
//	directory of each inventory directory.
//
// inventoryPaths: A []string containing the inventory directory paths to search.
//
// Returns:
//
// fullPath: A string containing the full path to the template file if it exists in the 'templates' folder of the parent
//
//	directory of any of the inventory directories, or an empty string if not found.
//
// error: An error if there is an issue checking the template file's existence.
//
// Example:
//
// fsys := afero.NewOsFs()
// templatePath := "bash"
// inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}
// fullPath, err := TemplateExists(fsys, templatePath, inventoryPaths)
//
//	if err != nil {
//	    log.Fatalf("failed to check template existence: %v", err)
//	}
//
//	if fullPath != "" {
//	    log.Printf("Template %s found in the parent directory of the inventory directories\n", templatePath)
//	} else {
//
//	    log.Printf("Template %s not found in the parent directory of the inventory directories\n", templatePath)
//	}
func TemplateExists(fsys afero.Fs, relTemplatePath string, inventoryPaths []string) (string, error) {
	for _, inventoryPath := range inventoryPaths {
		fsys := afero.NewOsFs()
		iofs := afero.NewIOFS(fsys)
		fullPath, err := blocks.FindFilePath(relTemplatePath, filepath.Dir(inventoryPath), iofs)
		if err != nil {
			// see comment above - callers expect this error to be suppressed
			if errors.Is(err, fs.ErrNotExist) {
				return "", nil
			} else {
				return "", err
			}
		}

		// parentDir := filepath.Dir(inventoryPath)
		// fullPath := filepath.Join(templatesDir, relTemplatePath)
		if _, err := fsys.Stat(fullPath); err == nil {
			return fullPath, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
	}

	return "", nil
}

// TTPExists checks if a TTP file exists in any of the inventory directories specified in the inventoryPaths parameter.
// If the TTP file is found, it returns true, otherwise, it returns false.
//
// Parameters:
//
// fsys: An afero.Fs representing the file system to operate on.
// ttpName: A string representing the name of the TTP file to search for in the inventory directories.
// inventoryPaths: A []string containing the inventory directory paths to search.
//
// Returns:
//
// bool: A boolean value indicating whether the TTP file exists in any of the inventory directories (true) or not (false).
// error: An error if there is an issue checking the TTP file's existence.
//
// Example:
//
// ttpName := "exampleTTP"
// inventoryPaths := []string{"path/to/inventory1", "path/to/inventory2"}
// fsys := afero.NewOsFs()
// exists, err := TTPExists(fsys, ttpName, inventoryPaths)
//
// if err != nil {
// log.Fatalf("failed to check TTP existence: %v", err)
// }
//
// if exists {
// log.Printf("TTP %s found in the inventory directories\n", ttpName)
// } else {
//
// log.Printf("TTP %s not found in the inventory directories\n", ttpName)
// }
func TTPExists(fsys afero.Fs, ttpName string, inventoryPaths []string) (bool, error) {
	ttpPath := filepath.Join("ttps", ttpName+".yaml")
	return PathExistsInInventory(fsys, ttpPath, inventoryPaths)
}

// MkdirAllFS is a filesystem-agnostic version of os.MkdirAll. It creates a directory named path,
// along with any necessary parents, and returns nil, or else returns an error.
// The permission bits perm are used for all directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing and returns nil.
//
// Parameters:
//
// fsys: An afero.Fs object representing the file system to operate on.
// path: A string representing the path to the directory to create, including any necessary parent directories.
// perm: An os.FileMode representing the permission bits for the created directories.
//
// Returns:
//
// error: An error if the directory could not be created.
//
// Example:
//
// fsys := afero.NewOsFs()
// dirPath := "path/to/directory"
// err := MkdirAllFS(fsys, dirPath, 0755)
//
// if err != nil {
// log.Fatalf("failed to create directory: %v", err)
// }
func MkdirAllFS(fsys afero.Fs, path string, perm os.FileMode) error {
	return fsys.MkdirAll(path, perm)
}

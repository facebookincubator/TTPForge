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
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// CreateDirIfNotExists checks if a directory exists at the given path and creates it if it does not exist.
// It returns an error if the directory could not be created.
//
// Parameters:
//
// path: A string representing the path to the directory to check and create if necessary.
//
// Returns:
//
// error: An error if the directory could not be created.
func CreateDirIfNotExists(path string) error {
	fileInfo, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it does not exist
			err = os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Check if the path is a directory
		if !fileInfo.IsDir() {
			return fmt.Errorf("%s is a file, not a directory", path)
		}
	}

	return nil
}

// PathExistsInInventory checks if a relative file path exists in any of the inventory directories specified in the
// configuration file. If the file is found in any of the inventory directories, it returns true, otherwise, it returns false.
//
// Parameters:
//
// relPath: A string representing the relative path of the file to search for in the inventory directories.
//
// Returns:
//
// bool: A boolean value indicating whether the file exists in any of the inventory directories (true) or not (false).
// error: An error if there is an issue checking the file's existence.
//
// Example:
//
// relFilePath := "templates/exampleTTP.yaml.tmpl"
// exists, err := PathExistsInInventory(relFilePath)
//
//	if err != nil {
//	  log.Fatalf("failed to check file existence: %v", err)
//	}
//
//	if exists {
//	  log.Printf("File %s found in the inventory directories\n", relFilePath)
//	} else {
//
//	  log.Printf("File %s not found in the inventory directories\n", relFilePath)
//	}
func PathExistsInInventory(relPath string) (bool, error) {
	inventory := viper.GetStringSlice("inventory")

	for _, invPath := range inventory {
		absPath := filepath.Join(invPath, relPath)
		if _, err := os.Stat(absPath); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// TemplateExists checks if a template file exists in any of the inventory directories specified in the configuration
// file. If the template file is found, it returns true, otherwise, it returns false.
//
// Parameters:
//
// templateName: A string representing the name of the template file to search for in the inventory directories.
//
// Returns:
//
// bool: A boolean value indicating whether the template file exists in any of the inventory directories (true) or not (false).
// error: An error if there is an issue checking the template file's existence.
//
// Example:
//
// templateName := "exampleTTP"
// exists, err := TemplateExists(templateName)
//
//	if err != nil {
//	  log.Fatalf("failed to check template existence: %v", err)
//	}
//
//	if exists {
//	  log.Printf("Template %s found in the inventory directories\n", templateName)
//	} else {
//
//	  log.Printf("Template %s not found in the inventory directories\n", templateName)
//	}
func TemplateExists(templateName string) (bool, error) {
	templatePath := filepath.Join("templates", templateName+"TTP.yaml.tmpl")
	return PathExistsInInventory(templatePath)
}

// TTPExists checks if a TTP file exists in any of the inventory directories specified in the configuration file.
// If the TTP file is found, it returns true, otherwise, it returns false.
//
// Parameters:
//
// ttpName: A string representing the name of the TTP file to search for in the inventory directories.
//
// Returns:
//
// bool: A boolean value indicating whether the TTP file exists in any of the inventory directories (true) or not (false).
// error: An error if there is an issue checking the TTP file's existence.
//
// Example:
//
// ttpName := "exampleTTP"
// exists, err := TTPExists(ttpName)
//
//	if err != nil {
//	  log.Fatalf("failed to check TTP existence: %v", err)
//	}
//
//	if exists {
//	  log.Printf("TTP %s found in the inventory directories\n", ttpName)
//	} else {
//
//	  log.Printf("TTP %s not found in the inventory directories\n", ttpName)
//	}
func TTPExists(ttpName string) (bool, error) {
	ttpPath := filepath.Join("ttps", ttpName+".yaml")
	return PathExistsInInventory(ttpPath)
}

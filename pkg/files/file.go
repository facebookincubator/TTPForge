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
	"os"
	"path/filepath"
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
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// TemplateExists checks if a template file with the given name exists in the "templates" directory.
// It returns true if the template file exists, and false otherwise.
//
// Parameters:
//
// templateName: A string representing the name of the template file to check for existence.
//
// Returns:
//
// bool: A boolean value indicating whether the template file exists.
func TemplateExists(templateName string) bool {
	templatePath := filepath.Join("templates", templateName+"TTP.yaml.tmpl")
	if _, err := os.Stat(templatePath); err != nil {
		return false
	}

	return true
}

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

package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/files"
)

func TestCreateDirIfNotExists(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "tempFile")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temporary file

	tests := []struct {
		name       string
		path       string
		shouldFail bool
	}{
		{
			name: "creates a directory",
			path: "testDir",
		},
		{
			name: "does not create an existing directory",
			path: "testDir",
		},
		{
			name:       "handles invalid path",
			path:       "/nonexistent/testDir",
			shouldFail: true,
		},
		{
			name:       "returns error if path is a non-directory file",
			path:       tempFile.Name(),
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := files.CreateDirIfNotExists(tc.path)
			if err != nil && !tc.shouldFail {
				t.Fatalf("expected no error, got: %v", err)
			}

			if err == nil && tc.shouldFail {
				t.Fatal("expected an error, but got none")
			}
		})
	}

	os.RemoveAll("testDir")
}

func TestTemplateExists(t *testing.T) {
	// Navigate to the root of the repository
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatalf("failed to change into the templates directory: %v", err)
	}

	// Create test template file
	testTemplatePath := "testTemplate"
	f, err := os.Create(filepath.Join("templates", testTemplatePath+"TTP.yaml.tmpl"))
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}
	f.Close()

	tests := []struct {
		name           string
		templateName   string
		expectedResult bool
	}{
		{
			name:           "finds an existing template",
			templateName:   testTemplatePath,
			expectedResult: true,
		},
		{
			name:           "does not find a nonexistent template",
			templateName:   "nonexistentTemplate.yaml.tmpl",
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := files.TemplateExists(tc.templateName)
			if result != tc.expectedResult {
				t.Fatalf("expected %v, got: %v", tc.expectedResult, result)
			}
		})
	}

	// Clean up test template file
	if err := os.Remove(filepath.Join("templates", testTemplatePath+"TTP.yaml.tmpl")); err != nil {
		t.Fatalf("Failed to remove test template: %v", err)
	}
}

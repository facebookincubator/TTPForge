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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/files"
	"github.com/spf13/viper"
)

func TestCreateDirIfNotExists(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "tempFile")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	// Clean up the temporary file
	defer os.Remove(tempFile.Name())

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

func TestPathExistsInInventory(t *testing.T) {
	testDir, err := os.MkdirTemp("", "inventory")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	viper.Set("inventory", []string{testDir})

	filePath := filepath.Join(testDir, "test.txt")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if _, err := io.WriteString(file, "test content"); err != nil {
		t.Fatalf("failed to write to test file: %v", err)
	}

	file.Close()

	tests := []struct {
		name     string
		relPath  string
		expected bool
	}{
		{
			name:     "file exists in inventory",
			relPath:  "test.txt",
			expected: true,
		},
		{
			name:     "file does not exist in inventory",
			relPath:  "nonexistent.txt",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := files.PathExistsInInventory(tc.relPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, exists)
			}
		})
	}
}

func TestTemplateExists(t *testing.T) {
	testDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	templateDir := filepath.Join(testDir, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}

	templatePath := filepath.Join(templateDir, "exampleTTP.yaml.tmpl")
	templateFile, err := os.Create(templatePath)
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}
	if _, err := io.WriteString(templateFile, "test template content"); err != nil {
		t.Fatalf("failed to write to test template: %v", err)
	}
	templateFile.Close()

	tests := []struct {
		name        string
		template    string
		shouldExist bool
	}{
		{
			name:        "template exists",
			template:    "exampleTTP",
			shouldExist: true,
		},
		{
			name:        "template does not exist",
			template:    "nonexistentTTP",
			shouldExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := files.TemplateExists(tc.template)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.shouldExist {
				t.Fatalf("expected %v, got %v", tc.shouldExist, exists)
			}
		})
	}
}

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
	"fmt"
	"io"
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

	defer os.RemoveAll("testDir")
}

func createTestInventory(t *testing.T, dir string) {
	t.Helper()

	lateralMovementDir := filepath.Join(dir, "ttps", "lateral-movement", "ssh")
	if err := os.MkdirAll(lateralMovementDir, 0755); err != nil {
		t.Fatalf("failed to create lateral movement dir: %v", err)
	}

	privEscalationDir := filepath.Join(dir, "ttps", "privilege-escalation", "credential-theft", "hello-world")
	if err := os.MkdirAll(privEscalationDir, 0755); err != nil {
		t.Fatalf("failed to create privilege escalation dir: %v", err)
	}

	testFiles := []struct {
		path     string
		contents string
	}{
		{
			path:     filepath.Join(lateralMovementDir, "rogue-ssh-key.yaml"),
			contents: fmt.Sprintln("---\nname: test-rogue-ssh-key-contents"),
		},
		{
			path:     filepath.Join(privEscalationDir, "hello-world.yaml"),
			contents: fmt.Sprintln("---\nname: test-priv-esc-key-contents"),
		},
	}

	for _, file := range testFiles {
		f, err := os.Create(file.path)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		if _, err := io.WriteString(f, file.contents); err != nil {
			t.Fatalf("failed to write to test file: %v", err)
		}
		f.Close()
	}
}

func TestPathExistsInInventory(t *testing.T) {
	testDir, err := os.MkdirTemp("", "inventory")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	createTestInventory(t, testDir)
	// Specify test inventory path
	inventoryPaths := []string{filepath.Join(testDir, "ttps")}

	tests := []struct {
		name     string
		relPath  string
		expected bool
	}{
		{
			name:     "file exists in inventory",
			relPath:  "lateral-movement/ssh/rogue-ssh-key.yaml",
			expected: true,
		},
		{
			name:     "file exists in inventory",
			relPath:  "privilege-escalation/credential-theft/hello-world/hello-world.yaml",
			expected: true,
		},
		{
			name:     "file does not exist in inventory",
			relPath:  "non/nonexistent.txt",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := files.PathExistsInInventory(tc.relPath, inventoryPaths)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, exists)
			}
		})
	}
}

// Borrowed from: https://github.com/l50/goutils/blob/e91b7c4e18e23c53e35d04fa7961a5a14ca8ef39/fileutils_test.go#L294-L340
func TestExpandHomeDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home directory: %v", err)
	}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "EmptyPath",
			input:    "",
			expected: "",
		},
		{
			name:     "NoTilde",
			input:    "/path/without/tilde",
			expected: "/path/without/tilde",
		},
		{
			name:     "TildeOnly",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "TildeWithSlash",
			input:    "~/path/with/slash",
			expected: filepath.Join(homeDir, "path/with/slash"),
		},
		{
			name:     "TildeWithoutSlash",
			input:    "~path/without/slash",
			expected: filepath.Join(homeDir, "path/without/slash"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := files.ExpandHomeDir(tc.input)
			if actual != tc.expected {
				t.Errorf("test failed: ExpandHomeDir(%q) = %q; expected %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestTemplateExists(t *testing.T) {
	testDir, err := os.MkdirTemp("", "inventory")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	createTestInventory(t, testDir)
	// Specify test inventory path
	inventoryPaths := []string{filepath.Join(testDir, "ttps")}

	bashTemplateDir := filepath.Join(testDir, "templates", "bash")
	if err := os.MkdirAll(bashTemplateDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}

	templateFile, err := os.Create(filepath.Join(bashTemplateDir, "bashTTP.yaml.tmpl"))
	if err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	if _, err := io.WriteString(templateFile, "test template content"); err != nil {
		t.Fatalf("failed to write to test template: %v", err)
	}
	defer templateFile.Close()

	tests := []struct {
		name        string
		template    string
		shouldExist bool
	}{
		{
			name:        "template exists",
			template:    "bash",
			shouldExist: true,
		},
		{
			name:        "template does not exist",
			template:    "nonexistent",
			shouldExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := files.TemplateExists(filepath.Join("templates", tc.template), inventoryPaths)

			if exists != tc.shouldExist {
				t.Fatalf("expected %v, got %v", tc.shouldExist, exists)
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.shouldExist {
				t.Fatalf("expected %v, got %v", tc.shouldExist, exists)
			}
		})
	}
}

func TestTTPExists(t *testing.T) {
	testDir, err := os.MkdirTemp("", "inventory")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	createTestInventory(t, testDir)
	inventoryPaths := []string{testDir}

	ttpDir := filepath.Join(testDir, "ttps")
	if err := os.MkdirAll(ttpDir, 0755); err != nil {
		t.Fatalf("failed to create ttps directory: %v", err)
	}

	ttpPath := filepath.Join(ttpDir, "exampleTTP.yaml")
	ttpFile, err := os.Create(ttpPath)
	if err != nil {
		t.Fatalf("failed to create test ttp: %v", err)
	}
	if _, err := io.WriteString(ttpFile, "test ttp content"); err != nil {
		t.Fatalf("failed to write to test ttp: %v", err)
	}
	ttpFile.Close()

	tests := []struct {
		name        string
		ttpName     string
		shouldExist bool
	}{
		{
			name:        "TTP exists",
			ttpName:     "exampleTTP",
			shouldExist: true,
		},
		{
			name:        "TTP does not exist",
			ttpName:     "nonexistentTTP",
			shouldExist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exists, err := files.TTPExists(tc.ttpName, inventoryPaths)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.shouldExist {
				t.Fatalf("expected %v, got %v", tc.shouldExist, exists)
			}
		})
	}
}

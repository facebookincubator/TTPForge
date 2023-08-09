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
	"github.com/spf13/afero"
)

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

	bashTmplDir := filepath.Join(dir, "templates", "bash")
	if err := os.MkdirAll(bashTmplDir, 0755); err != nil {
		t.Fatalf("failed to create bash template dir: %v", err)
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

func TestCreateDirIfNotExists(t *testing.T) {
	tempFile, err := os.CreateTemp("", "tempFile")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
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

	fsys := afero.NewOsFs()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := files.CreateDirIfNotExists(fsys, tc.path)
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

func TestTemplateExists(t *testing.T) {
	testDir, err := os.MkdirTemp("", "inventory")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	createTestInventory(t, testDir)

	testCases := []struct {
		name           string
		expected       string
		relPath        string
		inventoryPaths []string
		notEmpty       bool
	}{
		{
			name:           "The bash templates are identified using the inventory",
			expected:       filepath.Join(testDir, "templates", "bash"),
			inventoryPaths: []string{filepath.Join(testDir, "ttps")},
			relPath:        "templates/bash",
			notEmpty:       true,
		},
		{
			name:           "Non-existent templates are handled properly",
			notEmpty:       false,
			inventoryPaths: []string{filepath.Join(testDir, "ttps")},
			relPath:        "templates/NOTREAL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullPath, err := files.TemplateExists(afero.NewOsFs(), tc.relPath, tc.inventoryPaths)
			if err != nil {
				t.Errorf("failed to check file existence: %v", err)
			}

			switch {
			case tc.notEmpty && fullPath == "":
				t.Errorf("test failed: TemplateExists(%v, %q, %v) returned an empty string; expected a non-empty string", afero.NewOsFs(), tc.relPath, tc.inventoryPaths)
			case !tc.notEmpty && fullPath != "":
				t.Errorf("test failed: TemplateExists(%v, %q, %v) = %v; expected an empty string", afero.NewOsFs(), tc.relPath, tc.inventoryPaths, fullPath)
			case tc.notEmpty && fullPath != tc.expected:
				t.Errorf("test failed: TemplateExists(%v, %q, %v) = %v; expected %v", afero.NewOsFs(), tc.relPath, tc.inventoryPaths, fullPath, tc.expected)
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

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("failed to change into test directory: %v", err)
	}

	tests := []struct {
		name        string
		ttpName     string
		shouldExist bool
	}{
		{
			name:        "TTP exists",
			ttpName:     "lateral-movement/ssh/rogue-ssh-key",
			shouldExist: true,
		},
		{
			name:        "TTP exists",
			ttpName:     "privilege-escalation/credential-theft/hello-world/hello-world",
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
			exists, err := files.TTPExists(afero.NewOsFs(), tc.ttpName, inventoryPaths)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if exists != tc.shouldExist {
				t.Fatalf("expected %v, got %v", tc.shouldExist, exists)
			}
		})
	}
}

func TestMkdirAllFS(t *testing.T) {
	testDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	fsys := afero.NewBasePathFs(afero.NewOsFs(), testDir)

	tests := []struct {
		name       string
		path       string
		shouldFail bool
	}{
		{
			name: "creates directory with parents",
			path: filepath.Join("nested", "dir"),
		},
		{
			name: "does not create an existing directory",
			path: ".",
		},
		{
			name:       "handles invalid path",
			path:       "../nonexistent/dir",
			shouldFail: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := files.MkdirAllFS(fsys, tc.path, 0755)
			if err != nil && !tc.shouldFail {
				t.Fatalf("expected no error, got: %v", err)
			}

			if err == nil && tc.shouldFail {
				t.Fatal("expected an error, but got none")
			}

			if !tc.shouldFail {
				exists, err := afero.Exists(fsys, tc.path)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !exists {
					t.Fatalf("directory %s should have been created", tc.path)
				}
			}
		})
	}
}

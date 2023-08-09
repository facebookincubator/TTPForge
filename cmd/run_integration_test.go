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

package cmd_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/cmd"
	mageutils "github.com/l50/goutils/v2/dev/mage"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const yamlContent = `---
name: paramtest
description: Test variadiac parameter handling
args:
  - name: user
  - name: password
  - name: optional_step_one
    type: bool
  - name: optional_step_two
    type: bool
    default: false
  - name: optional_step_three
    type: bool
    default: true
steps:
  - name: "paramtest"
    inline: |
      set -e

      user="$(echo {{ .Args.user }} | tr -d '\n\t\r')"
      password="$(echo {{ .Args.password }} | tr -d '\n\t\r')"

      go run variadicParameterExample.go \
        --user $user \
        --password $password
{{ if .Args.optional_step_one }}
  - name: optional_step_one
    inline: echo "optional step one"
{{ end }}
{{ if .Args.optional_step_two }}
  - name: optional_step_two
    inline: echo "optional step two"
{{ end }}
{{ if .Args.optional_step_three }}
  - name: optional_step_three
    inline: echo "optional step three"
{{ end }}`

const goMod = `
module github.com/facebookincubator/ttpforge

go 1.20

require (
        github.com/spf13/cobra v1.7.0
)`

const goContent = `package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	variadicParamsExampleCmd = &cobra.Command{
		Use:   "variadicParamsExample",
		Short: "Execute variadic parameters example",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("User input: %s\n", user)
			fmt.Printf("Password input: %s\n", password)
		},
	}

	password string
	user     string
)

func init() {
	variadicParamsExampleCmd.Flags().StringVar(&user,
		"user", "", "Email address for the variadicParamsExample user")

	variadicParamsExampleCmd.Flags().StringVar(&password,
		"password", "", "Password for the variadicParamsExample user")
}

func main() {
	if err := variadicParamsExampleCmd.Execute(); err != nil {
		fmt.Errorf("%s failed to run: %v", variadicParamsExampleCmd.Short, err)
		os.Exit(1)
	}
}`

func configTestEnvironment(t *testing.T) (string, string) {
	// Create a temporary directory
	testDir, err := os.MkdirTemp("", "run-integration-test")
	assert.NoError(t, err, "failed to create temporary directory")

	// Create ttp dir
	ttpDir := filepath.Join(testDir, "ttps")
	if err := os.MkdirAll(ttpDir, 0755); err != nil {
		t.Fatalf("failed to create ttps directory: %v", err)
	}

	// config for the test
	testConfigYAML := `---
inventory:
  - ` + ttpDir + `
logfile: ""
nocolor: false
verbose: false
`

	// Write the config to a temporary file
	testConfigYAMLPath := filepath.Join(testDir, "config.yaml")

	if err := os.WriteFile(testConfigYAMLPath, []byte(testConfigYAML), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create test TTP file
	if err := os.WriteFile(filepath.Join(testDir, "paramtest.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create yaml file: %v", err)
	}

	// Create the test Go file
	goFilePath := filepath.Join(testDir, "variadicParameterExample.go")
	if err := os.WriteFile(goFilePath, []byte(goContent), 0644); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	// Create the test go.mod file
	goModPath := filepath.Join(testDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create test go.mod file: %v", err)
	}

	return testDir, testConfigYAMLPath
}

func captureOutput(f func()) (string, error) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)
	errC := make(chan error) // error channel
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			log.Printf("failed to copy buffer: %v", err)
			errC <- err // send error back to the main Goroutine
		} else {
			outC <- buf.String()
		}
		close(outC)
		close(errC)
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC
	err := <-errC

	if err != nil {
		return "", fmt.Errorf("an error occurred in the goroutine: %w", err)
	}

	return out, nil
}

func TestRunCommandVariadicArgs(t *testing.T) {
	testDir, testConfigYAMLPath := configTestEnvironment(t)
	defer os.RemoveAll(testDir)

	testCases := []struct {
		name            string
		setFlags        func(*cobra.Command)
		user            string
		password        string
		optionalStepOne bool
		optionalStepTwo bool
		expected        string
		err             bool
	}{
		{
			name: "Should successfully run command with correct arguments",
			setFlags: func(newRunTTPCmd *cobra.Command) {
				_ = newRunTTPCmd.Flags().Set("config", testConfigYAMLPath)
				_ = newRunTTPCmd.Flags().Set("no-cleanup", "true")
			},
			user:            "testUser",
			password:        "testPassword",
			optionalStepOne: false,
			optionalStepTwo: true,
			expected:        "User input: testUser\nPassword input: testPassword\noptional step two\noptional step three\n",
			err:             false,
		},
		{
			name:     "Should fail to run command without arguments",
			setFlags: func(newRunTTPCmd *cobra.Command) {},
			user:     "",
			password: "",
			err:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newRunTTPCmd := cmd.RunTTPCmd()

			// Set flags for the test case
			tc.setFlags(newRunTTPCmd)
			_ = newRunTTPCmd.Flags().Set("arg", fmt.Sprintf("user=%s", tc.user))
			_ = newRunTTPCmd.Flags().Set("arg", fmt.Sprintf("password=%s", tc.password))
			_ = newRunTTPCmd.Flags().Set("arg", fmt.Sprintf("optional_step_one=%v", tc.optionalStepOne))
			_ = newRunTTPCmd.Flags().Set("arg", fmt.Sprintf("optional_step_two=%v", tc.optionalStepTwo))

			// Add the path to the TTP script file as an argument
			newRunTTPCmd.SetArgs([]string{filepath.Join(testDir, "paramtest.yaml")})

			// Change directory
			if err := os.Chdir(testDir); err != nil {
				t.Fatalf("failed to change into test directory: %v", err)
			}

			// Run this to generate go.sum
			if err := mageutils.Tidy(); err != nil {
				t.Fatalf("failed to run go mod tidy: %v", err)
			}

			// Capture command output and error
			output, err := captureOutput(func() {
				if err := newRunTTPCmd.Execute(); err != nil {
					fmt.Println(err)
				}
			})

			// Run error assertion
			assert.NoError(t, err)

			// Error assertion
			if tc.err {
				assert.NotEqual(t, tc.expected, output)
			} else {
				assert.Equal(t, tc.expected, output, "The output of the executed TTP script does not match the expected output")
			}
		})
	}
}

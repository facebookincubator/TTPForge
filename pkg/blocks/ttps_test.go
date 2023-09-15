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

package blocks_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/targets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v3"
)

// TestAmbiguousStepType verifies that we error
// out appropriately when an ambiguously-typed
// step is provided. Should probably live in step_test.go
// eventually but for current code structure
// it is better to have it live here.
func TestAmbiguousStepType(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "Ambiguous Inline+File Step",
			content: `name: test
description: this is a test
steps:
  - name: ambiguous
    inline: foo
    file: bar`,
		},
		{
			name: "Ambiguous Edit+SubTTP Step",
			content: `name: test
description: this is a test
steps:
  - name: ambiguous
    edit_file: hello
    ttp: world`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			assert.Error(t, err, "steps with ambiguous types should yield an error when parsed")
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "File scenario",
			content: `
---
name: Hello World
description: |
  Print hello world
steps:
  - name: hello
    file: ./ttps/privilege-escalation/credential-theft/hello-world/hello-world.sh
    cleanup:
      name: cleanup
      inline: |
        echo "cleaned up!"
        - name: hello_inline
        inline: |
          ./ttps/privilege-escalation/credential-theft/hello-world/hello-world.sh
`,
			wantError: false,
		},
		{
			name: "Basic Mitre scenario",
			content: `
---
name: Leverage mdfind to search for aws credentials on disk.
description: |
  This TTP runs a search using mdfind to search for AKIA strings in files,
  which would likely indicate that the file is an aws key.
mitre:
  tactics:
    - TA0006 Credential Access
  techniques:
    - T1552 Unsecured Credentials
  subtechniques:
    - "T1552.001 Unsecured Credentials: Credentials In Files"
steps:
  - name: mdfind_aws_keys
    inline: |
      echo -e "Searching for aws keys on disk using mdfind..."
      mdfind "kMDItemTextContent == '*AKIA*' || kMDItemDisplayName == '*AKIA*' -onlyin ~"
      echo "[+] TTP Done!"
`,
			wantError: false,
		},
		{
			name: "OS and Arch",
			content: `
---
name: os-and-arch-target
description: |
  Example demonstrating how to create a TTP that works for specific
  operating systems and architectures.

  This TTP also demonstrates how to create a TTP that can
  be used for specific cloud providers and regions.
targets:
  os:
    - linux
    - macos
  arch:
    - x86_64
    - arm64
  cloud:
    - provider: "aws"
      region: "us-west-1"
    - provider: "gcp"
      region: "us-west1-a"
    - provider: "azure"
      region: "eastus"

steps:
  - name: friendly-message
    inline: |
      set -e

      echo -e "You are running a TTP that works for the following operating systems: {{ .Targets.os }}"
      echo -e "and architectures: {{ .Targets.arch }}."
      echo -e "It can be used for the following cloud providers"
      echo -e "at the specified regions as well: {{ .Targets.cloud }}."
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTTP_ValidateSteps(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Valid steps",
			content: `
name: test
description: this is a test
steps:
  - name: step1
    inline: |
      echo "step1"
  - name: step2
    inline: |
      echo "step2"
`,
			wantError: false,
		},
		{
			name: "Valid targets and steps",
			content: `
---
name: enumerate-creds-lazagne
description: |
  Employ [LaZagne](https://github.com/AlessandroZ/LaZagne) for
  extracting credentials stored on disk and in memory of a target system.
args:
  - name: lazagne-path

steps:
  - name: setup
    inline: |
      if ! command -v python3 &> /dev/null; then
          echo "Error: Python3 is not installed on the current system, cannot run LaZagne"
          exit 1
      fi

      if ! command -v pip3 &> /dev/null; then
          echo "Error: pip3 is not installed on the current system, cannot run LaZagne"
          exit 1
      fi

      if ! command -v git &> /dev/null; then
          echo "Error: git is not installed on the current system, cannot run LaZagne"
          exit 1
      fi

      if [[ -d "{{ .Args.lazagne-path }}" ]]; then
          echo "Info: LaZagne already present on the current system"
      else
          git clone https://github.com/AlessandroZ/LaZagne.git {{ .Args.lazagne-path }}
      fi

      echo "Info: Ensuring the latest LaZagne dependencies are installed and up-to-date"
      cd {{ .Args.lazagne-path }} && pip3 install -r requirements.txt

  - name: run-lazagne
    inline: |
      set -e

      # Determine the operating system
      OS=$(uname)
      if [[ "$OS" == "Darwin" ]]; then
          export TARGET_OS="Mac"
      elif [[ "$OS" == "Linux" ]]; then
          export TARGET_OS="Linux"
      else
          echo "Unsupported operating system."
          exit 1
      fi

      echo "Running LaZagne"
      cd {{ .Args.lazagne-path }} && python3 ${TARGET_OS}/laZagne.py all

    cleanup:
      inline: |
        set -e

        echo "Uninstalling Python packages..."
        cd {{ .Args.lazagne-path }} && pip3 uninstall -y -r requirements.txt

        if [[ -d "{{ .Args.lazagne-path }}" ]]; then
            echo "Cleaning up LaZagne repository..."
            rm -rf {{ .Args.lazagne-path }}
        fi
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = ttp.ValidateSteps(blocks.TTPExecutionContext{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCleanupAfterStepFailure(t *testing.T) {
	content := `name: test
description: verifies that cleanups run after step failures
steps:
    - name: step1
      inline: echo "step1"
      cleanup:
        inline: echo "cleanup1"
    - name: step2
      inline: echo "step2"
      cleanup:
        inline: echo "cleanup2"
    - name: step3
      inline: THIS WILL FAIL ON PURPOSE
      cleanup:
        inline: echo "cleanup3"
    - name: step4
      inline: echo "step4"
      cleanup:
        inline: echo "cleanup4"`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(blocks.TTPExecutionConfig{})
	assert.Error(t, err, "should get an error from step failure")

	require.Equal(t, 2, len(stepResults.ByIndex))
	assert.Equal(t, "step1\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "step2\n", stepResults.ByIndex[1].Stdout)

	require.Equal(t, 2, len(stepResults.ByName))
	assert.Equal(t, "step1\n", stepResults.ByName["step1"].Stdout)
	assert.Equal(t, "step2\n", stepResults.ByName["step2"].Stdout)

	require.NotNil(t, stepResults.ByIndex[0].Cleanup)
	require.NotNil(t, stepResults.ByIndex[1].Cleanup)
	assert.Equal(t, "cleanup1\n", stepResults.ByIndex[0].Cleanup.Stdout)
	assert.Equal(t, "cleanup2\n", stepResults.ByIndex[1].Cleanup.Stdout)

	require.NotNil(t, stepResults.ByName["step1"].Cleanup)
	require.NotNil(t, stepResults.ByName["step2"].Cleanup)
	assert.Equal(t, "cleanup1\n", stepResults.ByName["step1"].Cleanup.Stdout)
	assert.Equal(t, "cleanup2\n", stepResults.ByName["step2"].Cleanup.Stdout)
}

func TestTemplatingArgsAndConditionalExec(t *testing.T) {
	content := `name: test_variable_expansion
description: tests args + step result variable expansion functionality
args:
- name: arg1
- name: do_optional_step_1
  default: false
- name: do_optional_step_2
  default: false
steps:
  - name: mandatory_step
    inline: echo "arg value is {{ .Args.arg1 }}"
{{ if .Args.do_optional_step_1 }}
  - name: optional_step_1
    inline: echo "optional step 1"
{{ end }}
{{ if .Args.do_optional_step_2 }}
  - name: optional_step_2
    inline: echo "optional step 2"
{{ end }}`

	execCfg := blocks.TTPExecutionConfig{
		Args: map[string]any{
			"arg1":               "victory",
			"do_optional_step_2": true,
		},
	}
	ttp, err := blocks.RenderTemplatedTTP(content, &execCfg)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(execCfg)
	require.NoError(t, err)

	require.Equal(t, 2, len(stepResults.ByIndex))
	assert.Equal(t, "arg value is victory\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "optional step 2\n", stepResults.ByIndex[1].Stdout)

	require.Equal(t, 2, len(stepResults.ByName))
	assert.Equal(t, "arg value is victory\n", stepResults.ByName["mandatory_step"].Stdout)
	assert.Equal(t, "optional step 2\n", stepResults.ByName["optional_step_2"].Stdout)
}

func TestVariableExpansionArgsAndStepResults(t *testing.T) {
	content := `name: test_variable_expansion
description: tests args + step result variable expansion functionality
args:
- name: arg1
steps:
  - name: step1
    inline: echo {\"foo\":{\"bar\":\"baz\"}}
    outputs:
      first:
        filters:
        - json_path: foo.bar
  - name: step2
    inline: echo "first output is baz"
  - name: step3
    inline: echo "arg value is {{ .Args.arg1 }}"`

	execCfg := blocks.TTPExecutionConfig{
		Args: map[string]any{
			"arg1": "victory",
		},
	}
	ttp, err := blocks.RenderTemplatedTTP(content, &execCfg)
	require.NoError(t, err)

	stepResults, err := ttp.RunSteps(execCfg)
	require.NoError(t, err)

	require.Equal(t, 3, len(stepResults.ByIndex))
	assert.Equal(t, "{\"foo\":{\"bar\":\"baz\"}}\n", stepResults.ByIndex[0].Stdout)
	assert.Equal(t, "first output is baz\n", stepResults.ByIndex[1].Stdout)
	assert.Equal(t, "arg value is victory\n", stepResults.ByIndex[2].Stdout)

	require.Equal(t, 3, len(stepResults.ByName))
	assert.Equal(t, "{\"foo\":{\"bar\":\"baz\"}}\n", stepResults.ByName["step1"].Stdout)
	assert.Equal(t, "first output is baz\n", stepResults.ByName["step2"].Stdout)
	assert.Equal(t, "arg value is victory\n", stepResults.ByName["step3"].Stdout)
}

func TestMitreAttackMapping(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Valid MITRE Mapping",
			content: `
name: TestTTP
description: Test description
mitre:
  tactics:
    - Initial Access
    - Execution
  techniques:
    - Spearphishing Link
  subtechniques:
    - Attachment
`,
			wantError: false,
		},
		{
			name: "Invalid MITRE Mapping - Missing Tactic",
			content: `
name: TestTTP
description: Test description
mitre:
  techniques:
    - Spearphishing Link
`,
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ttp.MitreAttackMapping.Tactics)
			}
		})
	}
}

func TestUnmarshalTargetsAndCloud(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		wantError   bool
		expectedTTP blocks.TTP
	}{
		{
			name: "Valid targets with cloud",
			content: `
---
name: cloud-target
description: Test cloud targeting
targets:
  os:
    - linux
    - windows
  arch:
    - x86_64
    - arm64
  cloud:
    - provider: "aws"
      region: "us-west-1"
    - provider: "gcp"
      region: "us-central1"
    - provider: "azure"
      region: "eastus"
`,
			wantError: false,
			expectedTTP: blocks.TTP{
				Name:        "cloud-target",
				Description: "Test cloud targeting",
				TargetSpec: targets.TargetSpec{
					OS:   []string{"linux", "windows"},
					Arch: []string{"x86_64", "arm64"},
					Cloud: []targets.Cloud{
						{
							Provider: "aws",
							Region:   "us-west-1",
						},
						{
							Provider: "gcp",
							Region:   "us-central1",
						},
						{
							Provider: "azure",
							Region:   "eastus",
						},
					},
				},
				Steps:    nil,
				ArgSpecs: nil,
				WorkDir:  "",
			},
		},
		{
			name: "Valid cloud-only targets",
			content: `
---
name: cloud
description: Cloud target
targets:
  cloud:
    - provider: "aws"
      region: "us-west-1"
`,
			wantError: false,
			expectedTTP: blocks.TTP{
				Name:        "cloud",
				Description: "Cloud target",
				TargetSpec: targets.TargetSpec{
					Cloud: []targets.Cloud{
						{
							Provider: "aws",
							Region:   "us-west-1",
						},
					},
				},
				Steps:    nil,
				ArgSpecs: nil,
				WorkDir:  "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttp blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttp)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			require.Equal(t, tc.expectedTTP, ttp, "Parsed Cloud-only TTP struct does not match expected")
		})
	}
}

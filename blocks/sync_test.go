package blocks_test

import (
	"os"
	"testing"

	"github.com/facebookincubator/TTP-Runner/blocks"
)

func TestLoadTTP(t *testing.T) {
	// Create a temporary file with valid TTP content.
	validTTPData := `
---
name: test_load_yaml
description: "Test loading and execution of YAML-based TTPs."
steps:
  - name: testing_exec
    file: ./tests/test.sh
    args:
      - blah
  - name: testing_exec_two
    file: ./tests/test.sh
    args:
      - blah
      - steps.testing_exec.output
`
	invalidTTPData := `
---
name: invalid_ttp
description: "Test detection and handling of invalid YAML-based TTPs."
steps:
  - name: "step1"
  - name: "step2"
`
	tmpFile1, err := os.CreateTemp(os.TempDir(), "valid_*.ttp")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := os.CreateTemp(os.TempDir(), "invalid_*.ttp")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile2.Name())

	if err := os.WriteFile(tmpFile1.Name(), []byte(validTTPData), 0644); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpFile1.Name(), err)
	}

	if err := os.WriteFile(tmpFile2.Name(), []byte(invalidTTPData), 0644); err != nil {
		t.Fatalf("failed to write to %s: %v", tmpFile2.Name(), err)
	}

	testCases := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"Valid TTP", tmpFile1.Name(), false},
		{"Invalid TTP", tmpFile2.Name(), true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := blocks.LoadTTP(tc.filename); (err != nil) != tc.wantErr {
				t.Errorf("error running LoadTTP() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

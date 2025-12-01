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

package validation

import (
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// Result holds the results of validation
type Result struct {
	Errors   []string
	Warnings []string
	Infos    []string
}

func (vr *Result) AddError(msg string) {
	vr.Errors = append(vr.Errors, msg)
}

func (vr *Result) AddWarning(msg string) {
	vr.Warnings = append(vr.Warnings, msg)
}

func (vr *Result) AddInfo(msg string) {
	vr.Infos = append(vr.Infos, msg)
}

func (vr *Result) HasErrors() bool {
	return len(vr.Errors) > 0
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[1;31m"
	colorYellow = "\033[1;33m"
	colorCyan   = "\033[1;36m"
	colorGreen  = "\033[1;32m"
)

func (vr *Result) Print() {
	if len(vr.Errors) > 0 {
		fmt.Printf("\n%sERROR (%d):%s\n", colorRed, len(vr.Errors), colorReset)
		for _, err := range vr.Errors {
			fmt.Printf("  %s✗%s %s\n", colorRed, colorReset, err)
		}
	}

	if len(vr.Warnings) > 0 {
		fmt.Printf("\n%sWARNING (%d):%s\n", colorYellow, len(vr.Warnings), colorReset)
		for _, warn := range vr.Warnings {
			fmt.Printf("  %s⚠%s %s\n", colorYellow, colorReset, warn)
		}
	}

	if len(vr.Infos) > 0 {
		fmt.Printf("\n%sINFO (%d):%s\n", colorCyan, len(vr.Infos), colorReset)
		for _, info := range vr.Infos {
			fmt.Printf("  %sℹ%s %s\n", colorCyan, colorReset, info)
		}
	}

	if !vr.HasErrors() {
		fmt.Printf("\n%s✓ TTP structure is valid!%s\n", colorGreen, colorReset)
	}
}

// ValidateTTP performs comprehensive validation using all checks
func ValidateTTP(ttpFilePath string, fsys afero.Fs, repo repos.Repo) *Result {
	result := &Result{}

	// Read file once and cache the content
	ttpBytes, err := readTTPBytesForValidation(ttpFilePath, fsys)
	if err != nil {
		result.AddError(fmt.Sprintf("Failed to read TTP file: %v", err))
		return result
	}

	ttpContent := string(ttpBytes)

	// Check required fields first - this always runs regardless of YAML parsing success
	ValidateRequiredFields(ttpContent, result)

	// Run structural validation checks
	ValidateStructure(ttpContent, result)

	// Attempt to parse YAML
	var ttpMap map[string]any
	err = yaml.Unmarshal(ttpBytes, &ttpMap)
	if err != nil {
		// If YAML parsing fails, it's likely due to templates
		result.AddWarning(fmt.Sprintf("YAML parsing had issues: %v - skipping full validation", err))
	} else {
		// YAML parsed successfully - run all structure-based validations
		ValidatePreamble(ttpMap, ttpBytes, result)
		ValidateRequirements(ttpMap, result)
		ValidateArgs(ttpMap, result)
		ValidateTemplateReferences(ttpMap, result)
	}

	// Run blocks package validation
	ValidateIntegration(ttpFilePath, ttpBytes, repo, result)

	return result
}

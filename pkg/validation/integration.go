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
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/parseutils"
	"github.com/facebookincubator/ttpforge/pkg/platforms"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// isTemplateRelatedError checks if an error is related to template rendering or variables
func isTemplateRelatedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "template") ||
		strings.Contains(errStr, "{{") ||
		strings.Contains(errStr, "<no value>")
}

// readTTPBytesForValidation reads TTP file contents
func readTTPBytesForValidation(ttpFilePath string, fsys afero.Fs) ([]byte, error) {
	var file afero.File
	var err error

	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	file, err = fsys.Open(ttpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return contents, nil
}

// renderTemplatedTTPForValidation renders templates with dummy values for args
// This allows template control structures ({{ if }}, {{ range }}) to be evaluated
// while preventing validation errors from <no value> substitutions
func renderTemplatedTTPForValidation(ttpStr string, rp blocks.RenderParameters) (*blocks.TTP, error) {
	// First, extract arg definitions from the YAML
	dummyArgs := extractDummyArgsFromTTP(ttpStr)

	// Merge dummy args with any provided args
	if rp.Args == nil {
		rp.Args = make(map[string]any)
	}
	for k, v := range dummyArgs {
		if _, exists := rp.Args[k]; !exists {
			rp.Args[k] = v
		}
	}

	// Now render the template with dummy args
	logging.L().Debugf("Rendering template with dummy args: %+v", rp.Args)
	tmpl, err := template.New("ttp").Funcs(sprig.TxtFuncMap()).Parse(ttpStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, rp)
	if err != nil {
		logging.L().Warnf("Template execution failed (this may be expected if args are not provided): %v", err)
		// Fallback to parsing raw YAML
		var ttp blocks.TTP
		err = yaml.Unmarshal([]byte(ttpStr), &ttp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode TTP YAML after template error: %w", err)
		}
		return &ttp, nil
	}

	logging.L().Debugf("Rendered YAML:\n%s", result.String())

	var ttp blocks.TTP
	err = yaml.Unmarshal(result.Bytes(), &ttp)
	if err != nil {
		return nil, fmt.Errorf("YAML unmarshal failed: %w", err)
	}

	return &ttp, nil
}

// extractDummyArgsFromTTP extracts arg definitions and creates dummy values
// This prevents {{.Args.foo}} from becoming <no value> during validation
func extractDummyArgsFromTTP(ttpStr string) map[string]any {
	dummyArgs := make(map[string]any)

	// Use parseutils to parse the TTP header (which includes args but excludes steps)
	// This handles template control structures in steps gracefully
	ttp, err := parseutils.ParseTTP([]byte(ttpStr), "validation")
	if err != nil {
		logging.L().Debugf("Failed to parse TTP for arg extraction: %v", err)
		return dummyArgs
	}

	// Create dummy values from parsed args
	for _, arg := range ttp.Args {
		// Check if there's a default value (highest priority)
		if arg.Default != nil {
			dummyArgs[arg.Name] = arg.Default
			continue
		}

		// Check if there are choices (use first choice)
		if len(arg.Choices) > 0 {
			dummyArgs[arg.Name] = arg.Choices[0]
			continue
		}

		// Generate dummy value based on type (lowest priority)
		dummyArgs[arg.Name] = generateDummyValueForType(arg.Type)
	}

	return dummyArgs
}

// generateDummyValueForType creates a dummy value based on arg type
func generateDummyValueForType(argType string) any {
	switch argType {
	case "int", "integer":
		return 1
	case "bool", "boolean":
		return true
	case "path":
		return "/tmp/dummy_path"
	case "string", "":
		return "dummy_value"
	default:
		return "dummy_value"
	}
}

// ValidateIntegration attempts validation using the blocks package
// Template errors are converted to warnings
// Note: Preamble validation is now done in ValidatePreamble, so this
// function focuses on step-level validation
//
// Note: The blocks package validation methods log errors to stderr using logging.L().Error().
// These ERROR logs will appear during validation but the actual validation results are
// captured in the Result object and displayed in the structured output at the end.
func ValidateIntegration(ttpFilePath string, ttpBytes []byte, repo repos.Repo, result *Result) {
	rp := blocks.RenderParameters{
		Args:     map[string]any{},
		Platform: platforms.GetCurrentPlatformSpec(),
	}

	ttp, err := renderTemplatedTTPForValidation(string(ttpBytes), rp)
	if err != nil {
		if isTemplateRelatedError(err) {
			result.AddWarning(fmt.Sprintf("TTP rendering with dummy args (templates may need real values): %v", err))
		} else {
			result.AddError(fmt.Sprintf("TTP rendering: %v", err))
		}
		// If we can't parse the full TTP, try to validate individual steps
		validateIndividualSteps(ttpBytes, repo, result)
		return
	}

	execCtx := blocks.NewTTPExecutionContext()
	execCtx.Cfg.Repo = repo

	absPath, err := filepath.Abs(ttpFilePath)
	if err == nil {
		ttp.WorkDir = filepath.Dir(absPath)
		execCtx.Vars.WorkDir = ttp.WorkDir
	}

	for idx, step := range ttp.Steps {
		stepCopy := step

		if err := stepCopy.Validate(execCtx); err != nil {
			if isTemplateRelatedError(err) {
				result.AddWarning(fmt.Sprintf("Step #%d (%s) template validation: %v", idx+1, step.Name, err))
			} else {
				result.AddError(fmt.Sprintf("Step #%d (%s) validation: %v", idx+1, step.Name, err))
			}
		}
	}
}

// extractStepsFromYAML extracts the steps array from TTP YAML bytes
func extractStepsFromYAML(ttpBytes []byte) ([]any, error) {
	var ttpMap map[string]any
	if err := yaml.Unmarshal(ttpBytes, &ttpMap); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	stepsVal, ok := ttpMap["steps"]
	if !ok {
		return nil, fmt.Errorf("no steps found")
	}

	stepsList, isList := stepsVal.([]any)
	if !isList {
		return nil, fmt.Errorf("steps is not a list")
	}

	return stepsList, nil
}

// getStepName returns a human-readable name for a step
func getStepName(stepMap map[string]any, idx int) string {
	stepName := fmt.Sprintf("#%d", idx+1)
	if nameVal, ok := stepMap["name"]; ok && nameVal != nil {
		stepName = fmt.Sprintf("#%d (%v)", idx+1, nameVal)
	}
	return stepName
}

// validateSingleStep validates a single step and reports errors/warnings to the result
func validateSingleStep(stepVal any, idx int, execCtx blocks.TTPExecutionContext, result *Result) {
	stepMap, isMap := stepVal.(map[string]any)
	if !isMap {
		return
	}

	stepName := getStepName(stepMap, idx)

	// Try to unmarshal this individual step into a block Step
	stepYAML, err := yaml.Marshal(stepVal)
	if err != nil {
		return
	}

	var step blocks.Step
	if err := yaml.Unmarshal(stepYAML, &step); err != nil {
		// Step failed to unmarshal - try to detect which action type was intended
		// and provide a more helpful error message
		errStr := err.Error()
		detailedErr := detectAndValidateActionType(stepMap, stepName, execCtx)
		if detailedErr != "" {
			result.AddError(fmt.Sprintf("Step %s: %v", stepName, detailedErr))
		} else if !strings.Contains(errStr, "no name specified") {
			// Only report the unmarshaling error if it's not about missing name
			// (missing name is already caught by ValidateRequiredFields)
			result.AddError(fmt.Sprintf("Step %s: %v", stepName, err))
		}
		return
	}

	// Step unmarshaled successfully, validate it
	if err := step.Validate(execCtx); err != nil {
		if isTemplateRelatedError(err) {
			result.AddWarning(fmt.Sprintf("Step %s template validation: %v", stepName, err))
		} else {
			result.AddError(fmt.Sprintf("Step %s validation: %v", stepName, err))
		}
	}
}

// validateIndividualSteps attempts to validate steps one by one
// This is a fallback when the full TTP can't be parsed
func validateIndividualSteps(ttpBytes []byte, repo repos.Repo, result *Result) {
	stepsList, err := extractStepsFromYAML(ttpBytes)
	if err != nil {
		return // Can't extract steps
	}

	execCtx := blocks.NewTTPExecutionContext()
	execCtx.Cfg.Repo = repo

	// Validate each step individually
	for idx, stepVal := range stepsList {
		validateSingleStep(stepVal, idx, execCtx, result)
	}
}

// detectAndValidateActionType tries to detect which action type the user intended
// and validates it directly to provide better error messages
func detectAndValidateActionType(stepMap map[string]any, _ string, execCtx blocks.TTPExecutionContext) string {
	stepYAML, err := yaml.Marshal(stepMap)
	if err != nil {
		return ""
	}

	// Try each action type and call its Validate() method directly
	actionCandidates := []struct {
		name   string
		action blocks.Action
		field  string
	}{
		{"create_file", blocks.NewCreateFileStep(), "create_file"},
		{"copy_path", blocks.NewCopyPathStep(), "copy_path"},
		{"http_request", blocks.NewHTTPRequestStep(), "http_request"},
		{"fetch_uri", blocks.NewFetchURIStep(), "fetch_uri"},
		{"edit_file", blocks.NewEditStep(), "edit_file"},
		{"kill_process", blocks.NewKillProcessStep(), "kill_process"},
		{"remove_path", blocks.NewRemovePathAction(), "remove_path"},
		{"print_str", blocks.NewPrintStrAction(), "print_str"},
		{"cd", blocks.NewChangeDirectoryStep(), "cd"},
		{"file", blocks.NewFileStep(), "file"},
		{"ttp", blocks.NewSubTTPStep(), "ttp"},
		{"inline", blocks.NewBasicStep(), "inline"},
		{"expect", blocks.NewExpectStep(), "responses"},
	}

	// Check which action field is present in the step
	for _, candidate := range actionCandidates {
		if _, ok := stepMap[candidate.field]; ok {
			// Try to unmarshal into this specific action type
			err := yaml.Unmarshal(stepYAML, candidate.action)
			if err != nil {
				continue
			}

			// Call Validate() to get the specific error message
			if validationErr := candidate.action.Validate(execCtx); validationErr != nil {
				return validationErr.Error()
			}

			// If validation passed but step still failed to parse, it might be due to IsNil()
			if candidate.action.IsNil() {
				return fmt.Sprintf("%s action has empty or missing required field", candidate.name)
			}
		}
	}

	// Special case for kill_process which can use kill_process_name or kill_process_id
	if _, hasName := stepMap["kill_process_name"]; hasName {
		action := blocks.NewKillProcessStep()
		if err := yaml.Unmarshal(stepYAML, action); err == nil {
			if validationErr := action.Validate(execCtx); validationErr != nil {
				return validationErr.Error()
			}
		}
	}
	if _, hasID := stepMap["kill_process_id"]; hasID {
		action := blocks.NewKillProcessStep()
		if err := yaml.Unmarshal(stepYAML, action); err == nil {
			if validationErr := action.Validate(execCtx); validationErr != nil {
				return validationErr.Error()
			}
		}
	}

	return "" // Couldn't determine the action type
}

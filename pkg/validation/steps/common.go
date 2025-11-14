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

package steps

import (
	"fmt"
	"strings"
)

// ValidationIssue represents a single validation issue found during step checking
type ValidationIssue struct {
	Level   string // "error", "warning", "info"
	Message string
}

// StepChecker is the interface that step-specific validators implement
type StepChecker interface {
	// Check validates a specific step and returns any validation issues found
	Check(stepMap map[string]any, stepName string) []ValidationIssue
}

// Registry holds all registered step checkers
var registry = make(map[string]StepChecker)

// Register adds a step checker to the registry for a given action type
func Register(actionType string, checker StepChecker) {
	registry[actionType] = checker
}

// GetChecker retrieves a step checker for a given action type
func GetChecker(actionType string) (StepChecker, bool) {
	checker, ok := registry[actionType]
	return checker, ok
}

// GetAllActionTypes returns all registered action types
func GetAllActionTypes() []string {
	actionTypes := make([]string, 0, len(registry))
	for actionType := range registry {
		actionTypes = append(actionTypes, actionType)
	}
	return actionTypes
}

// GetValidActions returns all valid action types recognized by TTPForge
func GetValidActions() []string {
	return []string{
		"inline",
		"create_file",
		"edit_file",
		"copy_path",
		"remove_path",
		"http_request",
		"fetch_uri",
		"cd",
		"expect",
		"print_str",
		"kill_process",
		"file",
		"ttp",
	}
}

// isEmptyStringValue checks if a value is empty, handling different types appropriately
// Returns true if the value is nil, an empty string, or a non-string type (unexpected)
func isEmptyStringValue(val any) (isEmpty bool, isWrongType bool) {
	if val == nil {
		return true, false
	}

	str, isString := val.(string)
	if !isString {
		return true, true
	}

	return strings.TrimSpace(str) == "", false
}

// validateRequiredStringField checks that a required field exists, is a string, and is non-empty
func validateRequiredStringField(stepMap map[string]any, fieldName, stepName string) []ValidationIssue {
	var issues []ValidationIssue

	fieldVal, ok := stepMap[fieldName]
	if !ok {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': %s requires %s", stepName, fieldName, fieldName),
		})
		return issues
	}

	isEmpty, isWrongType := isEmptyStringValue(fieldVal)
	if isWrongType {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': %s must be a string, got %T", stepName, fieldName, fieldVal),
		})
	} else if isEmpty {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': %s cannot be empty", stepName, fieldName),
		})
	}

	return issues
}

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
	"regexp"
	"slices"
	"strings"
)

var (
	namingPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Result interface defines the methods needed for validation result collection
type Result interface {
	AddError(msg string)
	AddWarning(msg string)
	AddInfo(msg string)
}

// ValidateSteps validates step list structure and delegates to step-specific validators
func ValidateSteps(ttpMap map[string]any, result Result) {
	stepsVal, ok := ttpMap["steps"]
	if !ok {
		return
	}

	stepsList, isList := stepsVal.([]any)
	if !isList {
		return
	}

	validActionsMap := getValidActionsMap()
	validExecutorsMap := getValidExecutorsMap()

	seenNames := make(map[string]bool)
	for i, step := range stepsList {
		stepMap, isMap := step.(map[string]any)
		if !isMap {
			result.AddError(fmt.Sprintf("Step %d must be a dictionary", i+1))
			continue
		}

		// Handle steps with or without names
		nameVal, hasName := stepMap["name"]
		var stepName string
		if !hasName {
			// Use a placeholder name for validation and reporting
			stepName = fmt.Sprintf("<unnamed step %d>", i+1)
			result.AddError(fmt.Sprintf("Step %d missing 'name' field", i+1))
			// Continue validation with placeholder name instead of skipping
		} else {
			stepName = fmt.Sprintf("%v", nameVal)
			if seenNames[stepName] {
				result.AddWarning(fmt.Sprintf("Duplicate step name: %s", stepName))
			}
			seenNames[stepName] = true
		}

		validateStepAction(stepMap, stepName, validActionsMap, result)
		ValidateCleanupForStep(stepMap, stepName, validActionsMap, result)

		_, hasInline := stepMap["inline"]
		_, hasFile := stepMap["file"]
		if hasInline || hasFile {
			if executor, hasExecutor := stepMap["executor"]; hasExecutor {
				execStr := fmt.Sprintf("%v", executor)
				if !validExecutorsMap[execStr] {
					result.AddWarning(fmt.Sprintf("Step '%s': non-standard executor '%s'", stepName, execStr))
				}
			} else if hasInline {
				result.AddInfo(fmt.Sprintf("Step '%s': No executor specified, will default to bash", stepName))
			}
		}

		if outputvar, ok := stepMap["outputvar"]; ok {
			outputvarStr := fmt.Sprintf("%v", outputvar)
			if !namingPattern.MatchString(outputvarStr) {
				result.AddWarning(fmt.Sprintf("Step '%s': Output variable name should be lowercase: %s", stepName, outputvarStr))
			}
		}
	}
}

func validateStepAction(stepMap map[string]any, stepName string, validActions map[string]bool, result Result) {
	actionFields := []string{}
	for field := range stepMap {
		if validActions[field] {
			actionFields = append(actionFields, field)
		}
	}

	if _, hasName := stepMap["kill_process_name"]; hasName {
		if !slices.Contains(actionFields, "kill_process") {
			actionFields = append(actionFields, "kill_process")
		}
	}
	if _, hasID := stepMap["kill_process_id"]; hasID {
		if !slices.Contains(actionFields, "kill_process") {
			actionFields = append(actionFields, "kill_process")
		}
	}

	if len(actionFields) == 0 {
		result.AddError(fmt.Sprintf("Step '%s' has no valid action", stepName))
		return
	}

	if len(actionFields) > 1 {
		result.AddError(fmt.Sprintf("Step '%s' has ambiguous type (multiple actions: %s)", stepName, strings.Join(actionFields, ", ")))
		return
	}

	action := actionFields[0]
	checker, ok := GetChecker(action)
	if !ok {
		return
	}

	// Get issues from checker and add them to result
	checkerIssues := checker.Check(stepMap, stepName)
	for _, issue := range checkerIssues {
		switch issue.Level {
		case "error":
			result.AddError(issue.Message)
		case "warning":
			result.AddWarning(issue.Message)
		case "info":
			result.AddInfo(issue.Message)
		}
	}
}

func getValidActionsMap() map[string]bool {
	validActions := GetValidActions()
	actionsMap := make(map[string]bool, len(validActions))
	for _, action := range validActions {
		actionsMap[action] = true
	}
	return actionsMap
}

func getValidExecutorsMap() map[string]bool {
	validExecutors := []string{
		"bash",
		"sh",
		"powershell",
		"cmd",
		"python",
		"python3",
	}
	executorsMap := make(map[string]bool, len(validExecutors))
	for _, executor := range validExecutors {
		executorsMap[executor] = true
	}
	return executorsMap
}

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

package blocks

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/repos"
)

const contextVariablePrefix = "$forge."

// TTPExecutionConfig - pass this into RunSteps to control TTP execution
type TTPExecutionConfig struct {
	DryRun              bool
	NoCleanup           bool
	CleanupDelaySeconds uint
	Args                map[string]any
	Repo                repos.Repo
}

// TTPExecutionContext - holds config and context for the currently executing TTP
type TTPExecutionContext struct {
	Cfg         TTPExecutionConfig
	StepResults *StepResultsRecord
}

// ExpandVariables takes a string containing the following types of variables
// and expands all of them to their appropriate values:
//
// * Step outputs: ($forge.steps.bar.outputs.baz)
//
// **Parameters:**
//
// inStrs: the list of strings that have variables expanded
//
// **Returns:**
//
// []string: the corresponding strings with variables expanded
// error: an error if there is a problem
func (c TTPExecutionContext) ExpandVariables(inStrs []string) ([]string, error) {
	re := regexp.MustCompile(
		`\$*` + regexp.QuoteMeta(contextVariablePrefix) + `[\w\.]*`,
	)
	var expandedStrs []string
	for _, inStr := range inStrs {
		var failedMatch string
		var failedMatchError error
		expandedStr := re.ReplaceAllStringFunc(inStr, func(match string) string {
			result, err := c.processMatch(match)
			if err != nil {
				failedMatch = match
				failedMatchError = err
			}
			return result
		})
		if failedMatchError != nil {
			return nil, fmt.Errorf("invalid variable expression %v: %v", failedMatch, failedMatchError)
		}
		expandedStrs = append(expandedStrs, expandedStr)
	}
	return expandedStrs, nil
}

func (c TTPExecutionContext) processStepsVariable(path string) (string, error) {
	tokens := strings.Split(path, ".")
	if len(tokens) < 2 {
		return "", fmt.Errorf("invalid step result reference: %v", "steps."+path)
	}

	stepName := tokens[0]
	stepResult, ok := c.StepResults.ByName[stepName]
	if !ok {
		return "", fmt.Errorf("invalid step name in variable path: %v", "steps."+path)
	}

	fieldSelector := tokens[1]
	switch fieldSelector {
	case "stdout":
		if len(tokens) != 2 {
			return "", fmt.Errorf("invalid step result reference (should end at stdout): %v", "steps."+path)
		}
		return stepResult.Stdout, nil
	case "outputs":
		if len(tokens) != 3 {
			return "", fmt.Errorf("step output reference %v should be exactly one level deep (e.g. steps.foo.outputs.bar)", "steps."+path)
		}
		key := tokens[2]
		val, ok := stepResult.Outputs[key]
		if !ok {
			return "", fmt.Errorf("key %v not found in output of step %v", key, stepName)
		}
		return val, nil
	}
	return "", fmt.Errorf("invalid step result field selector: %v", fieldSelector)
}

func (c TTPExecutionContext) processMatch(match string) (string, error) {
	if strings.HasPrefix(match, "$$") {
		return strings.TrimPrefix(match, "$"), nil
	}
	variableSpecifier := strings.TrimPrefix(match, contextVariablePrefix)
	tokens := strings.Split(variableSpecifier, ".")
	for _, token := range tokens {
		// happens if we have a something like {{steps.wut.}} or {{.steps.wut}}
		if token == "" {
			return "", errors.New("leading or trailing '.' in variable expression")
		}
	}
	if len(tokens) < 2 {
		return "", fmt.Errorf("invalid variable expression: %v", match)
	}

	prefix := tokens[0]
	path := strings.Join(tokens[1:], ".")
	if prefix == "steps" {
		return c.processStepsVariable(path)
	}
	return "", fmt.Errorf("invalid variable prefix: %v", prefix)
}

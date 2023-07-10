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
)

// TTPExecutionConfig - pass this into RunSteps to control TTP execution
type TTPExecutionConfig struct {
	CliInputs      []string
	NoCleanup      bool
	Args           map[string]string
	TTPSearchPaths []string
}

// TTPExecutionContext - holds config and context for the currently executing TTP
type TTPExecutionContext struct {
	Cfg         TTPExecutionConfig
	StepResults *StepResultsRecord
}

func (c TTPExecutionContext) processStepsVariable(path string) (string, error) {
	tokens := strings.Split(path, ".")
	stepName := tokens[0]
	if stepResult, ok := c.StepResults.ByName[stepName]; ok {
		return stepResult.Stdout, nil
	}
	return "", fmt.Errorf("invalid step name in variable expression: %v", "steps."+path)
}

func (c TTPExecutionContext) processMatch(match string) (string, error) {
	variableSpecifier := strings.TrimLeft(strings.TrimRight(match, "}"), "{")
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

	scope := tokens[0]
	path := strings.Join(tokens[1:], ".")
	switch scope {
	case "steps":
		return c.processStepsVariable(path)
	}
	return "", errors.New("invalid scope")
}

func (c TTPExecutionContext) ExpandVariables(inStrs []string) ([]string, error) {
	re := regexp.MustCompile(`\{\{([^\{\}]*)\}\}`)
	var expandedStrs []string
	for _, inStr := range inStrs {
		var invalidMatches []string
		expandedStr := re.ReplaceAllStringFunc(inStr, func(match string) string {
			result, err := c.processMatch(match)
			if err != nil {
				invalidMatches = append(invalidMatches, match)
			}
			return result
		})
		if len(invalidMatches) > 0 {
			return nil, fmt.Errorf("invalid variable expressions: %v", strings.Join(invalidMatches, ","))
		}
		expandedStrs = append(expandedStrs, expandedStr)
	}
	return expandedStrs, nil
}

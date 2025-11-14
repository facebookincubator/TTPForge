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

func init() {
	Register("expect", &ExpectChecker{})
}

// ExpectChecker validates expect steps
type ExpectChecker struct{}

// Check validates an expect step
func (c *ExpectChecker) Check(stepMap map[string]any, stepName string) []ValidationIssue {
	var issues []ValidationIssue

	expectVal, ok := stepMap["expect"]
	if !ok {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': expect action requires 'expect' field", stepName),
		})
		return issues
	}

	expectMap, isMap := expectVal.(map[string]any)
	if !isMap {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': expect field must be a dictionary", stepName),
		})
		return issues
	}

	if inlineVal, ok := expectMap["inline"]; !ok || strings.TrimSpace(fmt.Sprintf("%v", inlineVal)) == "" {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': expect.inline must be provided", stepName),
		})
	}

	if responsesVal, ok := expectMap["responses"]; !ok {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': expect.responses must be provided", stepName),
		})
	} else {
		responsesList, isList := responsesVal.([]any)
		if !isList || len(responsesList) == 0 {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Message: fmt.Sprintf("Step '%s': expect.responses must be provided and non-empty", stepName),
			})
		}
	}

	return issues
}

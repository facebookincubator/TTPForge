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
)

func init() {
	Register("copy_path", &CopyPathChecker{})
}

// CopyPathChecker validates copy_path steps
type CopyPathChecker struct{}

// Check validates a copy_path step
func (c *CopyPathChecker) Check(stepMap map[string]any, stepName string) []ValidationIssue {
	var issues []ValidationIssue

	issues = append(issues, validateRequiredStringField(stepMap, "copy_path", stepName)...)

	if _, ok := stepMap["to"]; !ok {
		issues = append(issues, ValidationIssue{
			Level:   "error",
			Message: fmt.Sprintf("Step '%s': copy_path requires 'to' field", stepName),
		})
	}

	return issues
}

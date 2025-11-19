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
	Register("http_request", &HTTPRequestChecker{})
}

// HTTPRequestChecker validates http_request steps
type HTTPRequestChecker struct {
	validHTTPMethods map[string]bool
}

// NewHTTPRequestChecker creates a new HTTPRequestChecker with valid HTTP methods
func NewHTTPRequestChecker() *HTTPRequestChecker {
	return &HTTPRequestChecker{
		validHTTPMethods: map[string]bool{
			"GET":     true,
			"POST":    true,
			"PUT":     true,
			"DELETE":  true,
			"HEAD":    true,
			"PATCH":   true,
			"OPTIONS": true,
		},
	}
}

// Check validates an http_request step
func (c *HTTPRequestChecker) Check(stepMap map[string]any, stepName string) []ValidationIssue {
	if c.validHTTPMethods == nil {
		c.validHTTPMethods = NewHTTPRequestChecker().validHTTPMethods
	}

	var issues []ValidationIssue

	issues = append(issues, validateRequiredStringField(stepMap, "http_request", stepName)...)

	if httpType, ok := stepMap["type"]; ok {
		typeStr := fmt.Sprintf("%v", httpType)
		if !c.validHTTPMethods[typeStr] {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				Message: fmt.Sprintf("Step '%s': Invalid HTTP method: %s", stepName, typeStr),
			})
		}
	}

	return issues
}

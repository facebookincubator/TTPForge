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
	"regexp"
)

// requiredFieldPatterns defines regex patterns for detecting required top-level fields
var requiredFieldPatterns = map[string]*regexp.Regexp{
	"name":        regexp.MustCompile(`(?m)^name:`),
	"uuid":        regexp.MustCompile(`(?m)^uuid:`),
	"api_version": regexp.MustCompile(`(?m)^api_version:`),
	"steps":       regexp.MustCompile(`(?m)^steps:`),
}

// ValidateRequiredFields checks that all required top-level fields are present
// This uses regex patterns to ensure consistency and runs regardless of YAML parsing success
func ValidateRequiredFields(content string, result *Result) {
	// Check each required field
	for fieldName, pattern := range requiredFieldPatterns {
		matches := pattern.FindAllString(content, -1)
		if len(matches) == 0 {
			result.AddError(fmt.Sprintf("Missing required field: %s", fieldName))
		} else if len(matches) > 1 {
			result.AddError(fmt.Sprintf("The top-level key '%s:' should occur exactly once", fieldName))
		}
	}
}

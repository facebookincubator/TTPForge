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
	"regexp"
)

var (
	nameKeyPattern       = regexp.MustCompile(`(?m)^name:`)
	uuidKeyPattern       = regexp.MustCompile(`(?m)^uuid:`)
	apiVersionKeyPattern = regexp.MustCompile(`(?m)^api_version:`)
	stepsKeyPattern      = regexp.MustCompile(`(?m)^steps:`)
)

// ValidateRequiredFields checks that all required top-level fields are present
// This uses regex patterns to ensure consistency and runs regardless of YAML parsing success
func ValidateRequiredFields(content string, result *Result) {
	// Check for required field: name
	nameMatches := nameKeyPattern.FindAllString(content, -1)
	if len(nameMatches) == 0 {
		result.AddError("Missing required field: name")
	} else if len(nameMatches) > 1 {
		result.AddError("The top-level key 'name:' should occur exactly once")
	}

	// Check for required field: uuid
	uuidMatches := uuidKeyPattern.FindAllString(content, -1)
	if len(uuidMatches) == 0 {
		result.AddError("Missing required field: uuid")
	} else if len(uuidMatches) > 1 {
		result.AddError("The top-level key 'uuid:' should occur exactly once")
	}

	// Check for required field: api_version
	apiVersionMatches := apiVersionKeyPattern.FindAllString(content, -1)
	if len(apiVersionMatches) == 0 {
		result.AddError("Missing required field: api_version")
	} else if len(apiVersionMatches) > 1 {
		result.AddError("The top-level key 'api_version:' should occur exactly once")
	}

	// Check for required field: steps
	stepsMatches := stepsKeyPattern.FindAllString(content, -1)
	if len(stepsMatches) == 0 {
		result.AddError("Missing required field: steps")
	} else if len(stepsMatches) > 1 {
		result.AddError("The top-level key 'steps:' should occur exactly once")
	}
}

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
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/args"
)

var namingPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateArgs validates argument definitions from the parsed preamble.
func ValidateArgs(argSpecs []args.Spec, result *Result) {
	seenArgs := make(map[string]bool)
	validTypes := args.GetValidArgTypes()
	validTypesMap := make(map[string]bool)
	for _, t := range validTypes {
		validTypesMap[t] = true
	}

	for i, spec := range argSpecs {
		if spec.Name == "" {
			result.AddError(fmt.Sprintf("Argument %d missing 'name' field", i+1))
			continue
		}

		if seenArgs[spec.Name] {
			result.AddError(fmt.Sprintf("Duplicate argument name: %s", spec.Name))
		}
		seenArgs[spec.Name] = true

		if !namingPattern.MatchString(spec.Name) {
			result.AddWarning(fmt.Sprintf("Argument name should be lowercase with underscores: %s", spec.Name))
		}

		if spec.Type != "" {
			if !validTypesMap[spec.Type] {
				result.AddError(fmt.Sprintf("Invalid argument type: %s (valid: %s)", spec.Type, strings.Join(validTypes, ", ")))
			}
		} else {
			result.AddInfo(fmt.Sprintf("Argument '%s' has no type specified (defaults to string)", spec.Name))
		}

		if spec.Default == nil {
			result.AddInfo(fmt.Sprintf("Argument '%s' has no default value", spec.Name))
		}
	}
}

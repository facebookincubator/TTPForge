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

var (
	namingPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// ValidateArgs validates argument definitions
func ValidateArgs(ttpMap map[string]any, result *Result) {
	argsVal, ok := ttpMap["args"]
	if !ok {
		return
	}

	argsList, isList := argsVal.([]any)
	if !isList {
		result.AddError("'args' must be a list")
		return
	}

	seenArgs := make(map[string]bool)
	for i, arg := range argsList {
		argMap, isMap := arg.(map[string]any)
		if !isMap {
			result.AddError(fmt.Sprintf("Argument %d must be a dictionary", i+1))
			continue
		}

		nameVal, hasName := argMap["name"]
		if !hasName {
			result.AddError(fmt.Sprintf("Argument %d missing 'name' field", i+1))
			continue
		}

		argName := fmt.Sprintf("%v", nameVal)

		if seenArgs[argName] {
			result.AddError(fmt.Sprintf("Duplicate argument name: %s", argName))
		}
		seenArgs[argName] = true

		if !namingPattern.MatchString(argName) {
			result.AddWarning(fmt.Sprintf("Argument name should be lowercase with underscores: %s", argName))
		}

		if argType, ok := argMap["type"]; ok {
			typeStr := fmt.Sprintf("%v", argType)
			validTypes := args.GetValidArgTypes()
			validTypesMap := make(map[string]bool)
			for _, t := range validTypes {
				validTypesMap[t] = true
			}
			if !validTypesMap[typeStr] {
				result.AddError(fmt.Sprintf("Invalid argument type: %s (valid: %s)", typeStr, strings.Join(validTypes, ", ")))
			}
		} else {
			result.AddInfo(fmt.Sprintf("Argument '%s' has no type specified (defaults to string)", argName))
		}

		if _, ok := argMap["description"]; !ok {
			result.AddWarning(fmt.Sprintf("Argument '%s' should have a description", argName))
		}

		if _, ok := argMap["default"]; !ok {
			result.AddInfo(fmt.Sprintf("Argument '%s' has no default value", argName))
		}
	}
}

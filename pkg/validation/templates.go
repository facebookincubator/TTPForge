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

	"gopkg.in/yaml.v3"
)

var (
	templateVarPattern = regexp.MustCompile(`\{\{\.Args\.([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	stepVarsPattern    = regexp.MustCompile(`\{\[\{\.StepVars\.([a-zA-Z_][a-zA-Z0-9_]*)\}\]\}`)
)

// ValidateTemplateReferences validates that template variables reference defined args/outputvars
// and that all defined args are actually used
func ValidateTemplateReferences(ttpMap map[string]any, result *Result) {
	definedArgs := make(map[string]bool)
	if argsVal, ok := ttpMap["args"]; ok {
		if argsList, isList := argsVal.([]any); isList {
			for _, arg := range argsList {
				if argMap, isMap := arg.(map[string]any); isMap {
					if nameVal, ok := argMap["name"]; ok {
						definedArgs[fmt.Sprintf("%v", nameVal)] = true
					}
				}
			}
		}
	}

	// Collect defined outputvars from steps (in order)
	definedOutputVars := make(map[string]bool)
	if stepsVal, ok := ttpMap["steps"]; ok {
		if stepsList, isList := stepsVal.([]any); isList {
			for _, step := range stepsList {
				if stepMap, isMap := step.(map[string]any); isMap {
					if outputvarVal, ok := stepMap["outputvar"]; ok {
						definedOutputVars[fmt.Sprintf("%v", outputvarVal)] = true
					}
				}
			}
		}
	}

	yamlBytes, err := yaml.Marshal(ttpMap)
	if err != nil {
		return
	}
	yamlStr := string(yamlBytes)

	// Validate {{.Args.varname}} references and track which args are used
	usedArgs := make(map[string]bool)
	argsMatches := templateVarPattern.FindAllStringSubmatch(yamlStr, -1)
	seenVars := make(map[string]bool)
	for _, match := range argsMatches {
		if len(match) > 1 {
			varName := match[1]
			usedArgs[varName] = true
			if !seenVars[varName] && !definedArgs[varName] {
				result.AddError(fmt.Sprintf("Template variable '{{.Args.%s}}' references undefined argument '%s'", varName, varName))
				seenVars[varName] = true
			}
		}
	}

	// Validate {[{.StepVars.varname}]} references
	stepVarsMatches := stepVarsPattern.FindAllStringSubmatch(yamlStr, -1)
	seenStepVars := make(map[string]bool)
	for _, match := range stepVarsMatches {
		if len(match) > 1 {
			varName := match[1]
			if !seenStepVars[varName] && !definedOutputVars[varName] {
				result.AddError(fmt.Sprintf("Template variable '{[{.StepVars.%s}]}' references undefined outputvar '%s'", varName, varName))
				seenStepVars[varName] = true
			}
		}
	}

	// Check for defined but unused arguments
	for argName := range definedArgs {
		if !usedArgs[argName] {
			result.AddWarning(fmt.Sprintf("Argument '%s' is defined but never used in the TTP", argName))
		}
	}
}

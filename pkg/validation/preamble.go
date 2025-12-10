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

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/preprocess"
	"gopkg.in/yaml.v3"
)

// ValidatePreamble validates description, MITRE ATT&CK, and preamble field values
// using both basic structure checks and semantic validation from the blocks package
// Note: Required field presence checks are handled in ValidateRequiredFields
func ValidatePreamble(ttpMap map[string]any, ttpBytes []byte, result *Result) {
	// Validate api_version value
	if apiVer, ok := ttpMap["api_version"]; ok {
		verStr := fmt.Sprintf("%v", apiVer)
		if verStr != "1.0" && verStr != "2.0" && verStr != "1" && verStr != "2" {
			result.AddWarning(fmt.Sprintf("Unusual api_version: %v (typically 1.0 or 2.0)", apiVer))
		}
	}

	// Validate uuid format
	if uuid, ok := ttpMap["uuid"]; ok {
		uuidStr := fmt.Sprintf("%v", uuid)
		uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidPattern.MatchString(uuidStr) {
			result.AddError(fmt.Sprintf("Invalid UUID format: %s (expected format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)", uuidStr))
		}
	}

	// Validate name value
	if name, ok := ttpMap["name"]; ok {
		nameStr := fmt.Sprintf("%v", name)
		if strings.TrimSpace(nameStr) == "" {
			result.AddError("'name' cannot be empty")
		}
		if regexp.MustCompile(`^\d`).MatchString(nameStr) {
			result.AddWarning(fmt.Sprintf("TTP name should not start with a number: %s", nameStr))
		}
	}

	// Validate authors value
	if authors, ok := ttpMap["authors"]; ok {
		authorsList, isList := authors.([]any)
		if !isList {
			result.AddWarning("'authors' should be a list")
		} else if len(authorsList) == 0 {
			result.AddWarning("'authors' list cannot be empty")
		} else {
			for i, author := range authorsList {
				authorStr := fmt.Sprintf("%v", author)
				if author == nil || strings.TrimSpace(authorStr) == "" {
					result.AddWarning(fmt.Sprintf("'authors' entry %d cannot be empty", i))
				}
			}
		}
	}

	// Validate description
	if desc, ok := ttpMap["description"]; ok {
		descStr := fmt.Sprintf("%v", desc)
		if strings.TrimSpace(descStr) == "" || len(strings.TrimSpace(descStr)) < 10 {
			result.AddWarning("'description' should be more detailed")
		}
	} else {
		result.AddWarning("Consider adding a 'description' field")
	}

	// Validate MITRE ATT&CK
	if mitre, ok := ttpMap["mitre"]; ok {
		_, isMap := mitre.(map[string]any)
		if !isMap {
			result.AddError("'mitre' must be a dictionary")
			return
		}
	} else {
		result.AddInfo("Consider adding MITRE ATT&CK mappings under 'mitre' field")
	}

	// Semantic validation using blocks package
	preprocessResult, err := preprocess.Parse(ttpBytes)
	if err != nil {
		result.AddError(fmt.Sprintf("Preamble preprocessing failed: %v", err))
		return
	}

	type PreambleContainer struct {
		blocks.PreambleFields `yaml:",inline"`
	}
	var preamble PreambleContainer
	err = yaml.Unmarshal(preprocessResult.PreambleBytes, &preamble)
	if err != nil {
		result.AddError(fmt.Sprintf("Failed to unmarshal YAML preamble: %v", err))
		return
	}

	// Validate using blocks package (strict=false means UUID format check is skipped)
	err = preamble.Validate(false)
	if err != nil {
		result.AddError(err.Error())
	}
}

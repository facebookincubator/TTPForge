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

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"gopkg.in/yaml.v3"
)

// ValidateRequirements validates requirements structure and semantics
// using both basic structure checks and validation from the blocks package
func ValidateRequirements(ttpMap map[string]any, result *Result) {
	req, ok := ttpMap["requirements"]
	if !ok {
		result.AddInfo("No requirements specified - TTP will run on all platforms")
		return
	}

	reqMap, isMap := req.(map[string]any)
	if !isMap {
		result.AddError("'requirements' must be a dictionary")
		return
	}

	// Basic structure check
	if platformsVal, ok := reqMap["platforms"]; ok {
		_, isList := platformsVal.([]any)
		if !isList {
			result.AddError("'platforms' in requirements must be a list")
			return
		}
	}

	// Semantic validation using blocks package
	// Convert the map to a strongly-typed struct for validation
	var reqConfig blocks.RequirementsConfig
	reqBytes, err := yaml.Marshal(reqMap)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(reqBytes, &reqConfig)
	if err != nil {
		result.AddError(fmt.Sprintf("Failed to parse requirements: %v", err))
		return
	}

	// Validate using blocks package (validates platforms OS/arch values)
	err = reqConfig.Validate()
	if err != nil {
		result.AddError(err.Error())
	}
}

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
	"gopkg.in/yaml.v3"
)

// ValidateStructure checks top-level structure after YAML parsing
// Note: Required field presence checks are handled in ValidateRequiredFields
func ValidateStructure(content string, result *Result) {
	// Parse YAML using Node API to preserve structure
	var node yaml.Node
	err := yaml.Unmarshal([]byte(content), &node)
	if err != nil {
		// If YAML is invalid, let the main validation handle it
		return
	}

	// The root node should be a document node containing a mapping node
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return
	}

	mappingNode := node.Content[0]

	// In a mapping node, keys and values alternate: [key1, value1, key2, value2, ...]
	if len(mappingNode.Content) < 2 {
		return
	}

	// Get the last key (second-to-last element in Content)
	lastKeyNode := mappingNode.Content[len(mappingNode.Content)-2]
	lastKeyName := lastKeyNode.Value

	// Check if "steps" exists and if it's not the last key
	hasSteps := false
	for i := 0; i < len(mappingNode.Content); i += 2 {
		if mappingNode.Content[i].Value == "steps" {
			hasSteps = true
			break
		}
	}

	if hasSteps && lastKeyName != "steps" {
		result.AddError("The top-level key 'steps:' should always be the last top-level key in the file")
	}

	// Validate steps structure
	for i := 0; i < len(mappingNode.Content); i += 2 {
		keyNode := mappingNode.Content[i]
		if keyNode.Value == "steps" {
			valueNode := mappingNode.Content[i+1]
			if valueNode.Kind != yaml.SequenceNode {
				result.AddError("'steps' must be a list")
			} else if len(valueNode.Content) == 0 {
				result.AddError("'steps' cannot be empty")
			}
			break
		}
	}
}

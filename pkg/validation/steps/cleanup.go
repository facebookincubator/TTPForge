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

// ValidateCleanupForStep validates cleanup actions for a single step
func ValidateCleanupForStep(stepMap map[string]any, stepName string, validActions map[string]bool, result Result) {
	cleanup, hasCleanup := stepMap["cleanup"]
	if !hasCleanup {
		needsCleanup := false
		for action := range validActions {
			if _, ok := stepMap[action]; ok {
				needsCleanup = true
				break
			}
		}
		if needsCleanup {
			result.AddInfo(fmt.Sprintf("Step '%s': Consider adding cleanup", stepName))
		}
		return
	}

	if cleanupStr, ok := cleanup.(string); ok && cleanupStr == "default" {
		return
	}

	cleanupMap, isMap := cleanup.(map[string]any)
	if !isMap {
		result.AddError(fmt.Sprintf("Step '%s': Cleanup must be 'default' or a dictionary", stepName))
		return
	}

	actionFound := false
	for action := range cleanupMap {
		if validActions[action] {
			actionFound = true
			break
		}
	}

	if !actionFound {
		result.AddError(fmt.Sprintf("Step '%s': Invalid cleanup action", stepName))
	}
}

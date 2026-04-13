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
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidatePreamble validates preamble field values using the canonical
// blocks.PreambleFields struct parsed via parseutils.ParsePreambleOnly.
func ValidatePreamble(preamble *blocks.PreambleFields, result *Result) {
	// Validate api_version value
	if preamble.APIVersion != "" {
		ver := preamble.APIVersion
		if ver != "1.0" && ver != "2.0" && ver != "1" && ver != "2" {
			result.AddWarning(fmt.Sprintf("Unusual api_version: %v (typically 1.0 or 2.0)", ver))
		}
	}

	// Validate uuid format
	if preamble.UUID != "" {
		if !uuidPattern.MatchString(preamble.UUID) {
			result.AddError(fmt.Sprintf("Invalid UUID format: %s (expected format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)", preamble.UUID))
		}
	}

	// Validate name value
	if preamble.Name != "" {
		if strings.TrimSpace(preamble.Name) == "" {
			result.AddError("'name' cannot be empty")
		}
		if regexp.MustCompile(`^\d`).MatchString(preamble.Name) {
			result.AddWarning(fmt.Sprintf("TTP name should not start with a number: %s", preamble.Name))
		}
	}

	// Validate authors
	if len(preamble.Authors) == 0 {
		// not an error — authors are optional
	} else {
		for i, author := range preamble.Authors {
			if strings.TrimSpace(author) == "" {
				result.AddWarning(fmt.Sprintf("'authors' entry %d cannot be empty", i))
			}
		}
	}

	// Validate description
	if strings.TrimSpace(preamble.Description) == "" || len(strings.TrimSpace(preamble.Description)) < 10 {
		result.AddWarning("'description' should be more detailed")
	}

	// Validate MITRE ATT&CK
	if preamble.MitreAttackMapping == nil {
		result.AddInfo("Consider adding MITRE ATT&CK mappings under 'mitre' field")
	}

	// Semantic validation using blocks package (validates requirements, mitre tactics, etc.)
	err := preamble.Validate(false)
	if err != nil {
		result.AddError(err.Error())
	}
}

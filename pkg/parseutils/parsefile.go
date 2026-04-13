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

package parseutils

import (
	"bytes"
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"gopkg.in/yaml.v3"
)

// ParsePreamble parses the preamble of a TTP YAML file into the canonical
// blocks.PreambleFields struct. It truncates at the steps: boundary to avoid
// parsing issues with Go templates in the steps section.
func ParsePreamble(data []byte, filename string) (*blocks.PreambleFields, error) {
	stepsIndex := bytes.Index(data, []byte("\nsteps:"))
	if stepsIndex != -1 {
		data = data[:stepsIndex+1]
	}
	var preamble blocks.PreambleFields
	if err := yaml.Unmarshal(data, &preamble); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preamble from %s: %w", filename, err)
	}
	return &preamble, nil
}

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

package checks

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// CommonCheckFields are common fields across all check types
type CommonCheckFields struct {
	Msg string `yaml:"msg"`
}

// Check is wrapper struct around a Condition.
// This wrapping setup is used so that we only
// need to implement the UnmarshalYAML method once (in Check)
// instead of having to implement it in each individual condition type.
// This is similar to what we do with ParseAction
// for decoding the actions associated with steps.
type Check struct {
	CommonCheckFields
	condition Condition
}

// Verify wraps the Verify method from the underlying condition
func (c *Check) Verify(ctx VerificationContext) error {
	if ctx.FileSystem == nil {
		ctx.FileSystem = afero.NewOsFs()
	}
	return c.condition.Verify(ctx)
}

// UnmarshalYAML implements custom deserialization
// process to ensure that the check is decoded
// into the correct struct type
func (c *Check) UnmarshalYAML(node *yaml.Node) error {

	// Decode all of the shared fields.
	// Use of this auxiliary type prevents infinite recursion
	var ccf CommonCheckFields
	err := node.Decode(&ccf)
	if err != nil {
		return err
	}
	c.CommonCheckFields = ccf

	if c.Msg == "" {
		return errors.New("no msg specified for check")
	}

	candidateTypeInstances := []Condition{
		&PathExists{},
	}
	for _, candidateTypeInstance := range candidateTypeInstances {
		err := node.Decode(candidateTypeInstance)
		if err == nil {
			if c.condition != nil {
				// Must catch conditions with ambiguous types, such as:
				// - path_exists: foo
				//   command_succeeds: bar
				//
				// This is a problem because we can't tell into
				// which concrete type we should decode
				return fmt.Errorf("check %q has ambiguous type", c.Msg)
			}
			c.condition = candidateTypeInstance
		}
	}
	if c.condition == nil {
		return fmt.Errorf("condition with msg %q did not match any valid condition type", c.Msg)
	}
	return nil
}

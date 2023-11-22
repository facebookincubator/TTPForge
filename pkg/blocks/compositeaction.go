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

package blocks

import "errors"

// CompositeAction is an action that executes multiple actions
type CompositeAction struct {
	actionDefaults `yaml:"-"`
	actions        []Action
}

// Execute executes all actions
func (ca *CompositeAction) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	for _, a := range ca.actions {
		if _, err := a.Execute(execCtx); err != nil {
			return nil, err
		}
	}
	return &ActResult{}, nil
}

// Validate each action and returns an error if any are thrown
func (ca *CompositeAction) Validate(execCtx TTPExecutionContext) error {
	for _, a := range ca.actions {
		if !a.CanBeUsedInCompositeAction() {
			return errors.New("cannot use action in composite")
		}
		if err := a.Validate(execCtx); err != nil {
			return err
		}
	}
	return nil
}

// CanBeUsedInCompositeAction set to true
func (ca *CompositeAction) CanBeUsedInCompositeAction() bool {
	return true
}

// IsNil provides a default implementation
// of the IsNil method from the Action interface.
// This saves us from having to declare this function for every steps
// If a specific action needs to be used in a composite action,
// it can override this step
func IsNil() bool {
	return false
}

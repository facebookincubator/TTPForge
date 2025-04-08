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

// Action is an interface that is implemented
// by all action types used in steps/cleanups
// (such as create_file, inline, etc)
type Action interface {
	IsNil() bool
	Validate(execCtx TTPExecutionContext) error
	Template(execCtx TTPExecutionContext) error
	Execute(execCtx TTPExecutionContext) (*ActResult, error)
	GetDescription() string
	GetDefaultCleanupAction() Action
	CanBeUsedInCompositeAction() bool
}

// Shared action fields struct that also provides
// default implementations for some Action
// interface methods.
// Every new action type should embed this struct
type actionDefaults struct {
	Description string `yaml:"description,omitempty"`
	OutputVar   string `yaml:"outputvar,omitempty"`
}

// IsNil provides a default implementation
// of the IsNil method from the Action interface.
func (ad *actionDefaults) IsNil() bool {
	return false
}

// GetDescription returns the description field from the action
func (ad *actionDefaults) GetDescription() string {
	return ad.Description
}

// GetDefaultCleanupAction provides a default implementation
// of the GetDefaultCleanupAction method from the Action interface.
// This saves us from having to declare this function for every steps
// If a specific action needs a default cleanup action (such as a create_file action),
// it can override this step
func (ad *actionDefaults) GetDefaultCleanupAction() Action {
	return nil
}

// CanBeUsedInCompositeAction provides a default implementation
// of the CanBeUsedInCompositeAction method from the Action interface.
// This saves us from having to declare this function for every steps
// If a specific action needs to be used in a composite action,
// it can override this step
func (ad *actionDefaults) CanBeUsedInCompositeAction() bool {
	return false
}

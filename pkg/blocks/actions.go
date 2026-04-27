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

import (
	"fmt"
	"time"
)

// DefaultExecutionTimeout is the default timeout for step execution.
const DefaultExecutionTimeout = 100 * time.Minute

// maxExecutionTimeout is the hard upper bound on a step's execution timeout,
// even when a step explicitly opts in via long_running. It guards against
// runaway typos like step_timeout: 700h.
const maxExecutionTimeout = 24 * time.Hour

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

// timedActionDefaults adds opt-in per-step deadline fields. Embed this only
// in step types whose execution honors a context deadline (basic, file).
// Step types that do not embed this will not surface step_timeout /
// long_running in their YAML schema.
type timedActionDefaults struct {
	StepTimeout string `yaml:"step_timeout,omitempty"`
	LongRunning bool   `yaml:"long_running,omitempty"`

	// resolved memoizes the parsed StepTimeout so Validate and Execute
	// don't repeat the parse + range checks. A successful resolve always
	// yields a positive duration, so zero unambiguously means "not yet
	// resolved". Errors are not cached so repeated calls report consistently.
	resolved time.Duration `yaml:"-"`
}

// resolveTimeout returns the effective execution timeout for a step.
//
// Behavior:
//   - StepTimeout empty → returns DefaultExecutionTimeout (100m).
//   - StepTimeout set → parsed via time.ParseDuration.
//   - Parsed duration must be > 0.
//   - Parsed duration must not exceed maxExecutionTimeout.
//   - Parsed duration > DefaultExecutionTimeout requires LongRunning: true,
//     so that long-timeout steps are explicitly opted in and discoverable in YAML.
//
// Checks are ordered so the first error a user sees is actionable: if the
// value is over the absolute cap, telling them to set long_running: true
// would be misleading because the value would still be rejected.
func (t *timedActionDefaults) resolveTimeout() (time.Duration, error) {
	if t.resolved != 0 {
		return t.resolved, nil
	}
	if t.StepTimeout == "" {
		t.resolved = DefaultExecutionTimeout
		return t.resolved, nil
	}
	d, err := time.ParseDuration(t.StepTimeout)
	if err != nil {
		return 0, fmt.Errorf("invalid step_timeout %q: %w", t.StepTimeout, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("step_timeout must be > 0, got %q", t.StepTimeout)
	}
	if d > maxExecutionTimeout {
		return 0, fmt.Errorf(
			"step_timeout %q exceeds the maximum allowed of %s",
			t.StepTimeout, maxExecutionTimeout,
		)
	}
	if d > DefaultExecutionTimeout && !t.LongRunning {
		return 0, fmt.Errorf(
			"step_timeout %q exceeds the default of %s; set long_running: true to opt in",
			t.StepTimeout, DefaultExecutionTimeout,
		)
	}
	t.resolved = d
	return t.resolved, nil
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

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
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Action interface {
	Execute(execCtx TTPExecutionContext) (*ActResult, error)
	Validate(execCtx TTPExecutionContext) error
	IsNil() bool
}

type CommonStepFields struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`

	// CleanupSpec is exported so that UnmarshalYAML
	// can see it - however, it should be considered
	// to be a private detail of this file
	// and not referenced elsewhere in the codebase
	CleanupSpec *yaml.Node `yaml:"cleanup,omitempty"`
}

// Step contains a TTPForge executable action
// and its associated cleanup action (if specified)
type Step struct {
	CommonStepFields

	// These are where the actual executable content
	// of the step (and its associated cleanup process)
	// live - they are not deserialized directly from YAML
	// but rather must be decoded by ParseAction
	action  Action
	cleanup Action
}

func (s *Step) UnmarshalYAML(node *yaml.Node) error {

	// Decode all of the shared fields.
	// Use of this auxiliary type prevents infinite recursion
	var csf CommonStepFields
	err := node.Decode(&csf)
	if err != nil {
		return err
	}
	s.CommonStepFields = csf

	if s.Name == "" {
		return errors.New("no name specified for step")
	}

	// figure out what kind of action is
	// associated with executing this step
	s.action, err = s.ParseAction(node)
	if err != nil {
		return err
	}

	// figure out what kind of action is
	// associated with cleaning up this step
	if csf.CleanupSpec != nil {
		s.cleanup, err = s.ParseAction(csf.CleanupSpec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Step) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	return s.action.Execute(execCtx)
}

func (s *Step) Validate(execCtx TTPExecutionContext) error {
	if err := s.action.Validate(execCtx); err != nil {
		return err
	}
	if s.cleanup != nil {
		if err := s.cleanup.Validate(execCtx); err != nil {
			return err
		}
	}
	return nil
}

func (s *Step) ParseAction(node *yaml.Node) (Action, error) {
	actionCandidates := []Action{NewBasicStep(), NewFileStep(), NewSubTTPStep(), NewEditStep(), NewFetchURIStep(), NewCreateFileStep()}
	var action Action
	for _, actionType := range actionCandidates {
		err := node.Decode(actionType)
		if err == nil && !actionType.IsNil() {
			if action != nil {
				// Must catch bad steps with ambiguous types, such as:
				// - name: hello
				//   file: bar
				//   ttp: foo
				//
				// we can't use KnownFields to solve this without a massive
				// refactor due to https://github.com/go-yaml/yaml/issues/460
				// note: we check for non-empty name earlier so s.Name will be non-empty
				return nil, fmt.Errorf("step %v has ambiguous type", s.Name)
			}
			action = actionType
		}
	}
	if action == nil {
		return nil, fmt.Errorf("step %v did not match any valid step type", s.Name)
	}
	return action, nil
}

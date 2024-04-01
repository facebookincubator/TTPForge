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
	"github.com/google/uuid"

	"github.com/facebookincubator/ttpforge/pkg/args"
)

// PreambleFields are TTP fields that can be parsed
// prior to rendering the TTP steps with `text/template`
//
// **Attributes:**
//
// Name: The name of the TTP.
// Description: A description of the TTP.
// MitreAttackMapping: A MitreAttack object containing mappings to the MITRE ATT&CK framework.
// Requirements: The Requirements to run the TTP
// ArgSpecs: An slice of argument specifications for the TTP.
type PreambleFields struct {
	APIVersion         string              `yaml:"api_version,omitempty"`
	UUID               string              `yaml:"uuid,omitempty"`
	Name               string              `yaml:"name,omitempty"`
	Description        string              `yaml:"description"`
	MitreAttackMapping *MitreAttack        `yaml:"mitre,omitempty"`
	Requirements       *RequirementsConfig `yaml:"requirements,omitempty"`
	ArgSpecs           []args.Spec         `yaml:"args,omitempty,flow"`
}

// Validate validates the preamble fields.
// It is used by both `ttpforge run` and `ttpforge test`
func (pf *PreambleFields) Validate(strict bool) error {
	if strict {
		if _, err := uuid.Parse(pf.UUID); err != nil {
			return fmt.Errorf("TTP '%s' has an invalid UUID", pf.Name)
		}
	}
	// validate MITRE mapping
	if pf.MitreAttackMapping != nil && len(pf.MitreAttackMapping.Tactics) == 0 {
		return fmt.Errorf("TTP '%s' has a MitreAttackMapping but no Tactic is defined", pf.Name)
	}

	// validate requirements
	if err := pf.Requirements.Validate(); err != nil {
		return fmt.Errorf("TTP '%s' has an invalid requirements section: %w", pf.Name, err)
	}
	return nil
}

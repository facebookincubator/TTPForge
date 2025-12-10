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

	"gopkg.in/yaml.v3"
)

// Arg is a struct that represents the information of an argument in a TTP in a YAML file.
type Arg struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Type        string `yaml:"type,omitempty"`
	Choices     []any  `yaml:"choices,omitempty"`
	Default     any    `yaml:"default,omitempty"`
}

// Mitre is a struct that represents a MITRE tactic, technique, or subtechnique in a YAML file.
type Mitre struct {
	Tactics       []string `yaml:"tactics,omitempty"`
	Techniques    []string `yaml:"techniques,omitempty"`
	Subtechniques []string `yaml:"subtechniques,omitempty"`
}

// Platform is a struct that represents a platform in a YAML file like Windows, Linux, etc.
type Platform struct {
	OS string `yaml:"os"`
}

// Requirements is a struct that represents the requirements of a TTP in a YAML file like Platform (OS), superuser, etc.
type Requirements struct {
	Platforms []Platform `yaml:"platforms,omitempty"`
	Superuser bool       `yaml:"superuser,omitempty"`
}

// TTP is a struct that represents a TTP in a YAML file.
type TTP struct {
	APIVersion   string       `yaml:"api_version"`
	UUID         string       `yaml:"uuid"`
	Name         string       `yaml:"name"`
	Authors      []string     `yaml:"authors,omitempty"`
	Description  string       `yaml:"description"`
	Requirements Requirements `yaml:"requirements,omitempty"`
	Mitre        Mitre        `yaml:"mitre,omitempty"`
	Args         []Arg        `yaml:"args,omitempty"`
}

// ParseTTP parses a YAML file and returns a map of the contents.
func ParseTTP(data []byte, filename string) (TTP, error) {
	// Find steps: in data
	stepsIndex := bytes.Index(data, []byte("\nsteps:"))
	if stepsIndex != -1 {
		data = data[:stepsIndex+1]
	}
	var ttp TTP
	if err := yaml.Unmarshal(data, &ttp); err != nil {
		return TTP{}, fmt.Errorf("Failed to unmarshal file %s: %w", filename, err)
	}
	return ttp, nil
}

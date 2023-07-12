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

package outputs

import (
	"errors"
	"fmt"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

// Parse uses provided output specifications to extract output values
// from the provided raw stdout string
//
// **Parameters:**
//
// specs: the specs for the outputs to be extracted
// inStr: the raw stdout string from the step whose outputs will be extracted
//
// **Returns:**
//
// map[string]string: the output keys and values
// error: an error if there is a problem
func Parse(specs map[string]Spec, inStr string) (map[string]string, error) {
	outputs := make(map[string]string)
	for name, spec := range specs {
		outStr, err := spec.Apply(inStr)
		if err != nil {
			return nil, err
		}
		outputs[name] = outStr
	}
	return outputs, nil
}

// Spec defines an output value for which
// a given step's stdout should be scanned
type Spec struct {
	Filters []Filter `yaml:"filters"`
}

// Filter can be used to extract an output value
// from the provided string using Apply(...)
type Filter interface {
	Apply(inStr string) (string, error)
}

// Apply applies all filters in this output spec
// to the target string in order, producing a new string
func (s *Spec) Apply(inStr string) (string, error) {
	var err error
	curStr := inStr
	for _, f := range s.Filters {
		curStr, err = f.Apply(curStr)
		if err != nil {
			return "", err
		}
	}
	return curStr, nil
}

// JSONFilter will parse a JSON string
// and extract the value at the provided path (like jq)
type JSONFilter struct {
	Path string `yaml:"json_path"`
}

// UnmarshalYAML is used to load specs from yaml files
func (s *Spec) UnmarshalYAML(node *yaml.Node) error {
	type SpecTmp struct {
		FilterNodes []yaml.Node `yaml:"filters"`
	}

	var tmp SpecTmp
	if err := node.Decode(&tmp); err != nil {
		return err
	}

	var filters []Filter
	for _, fn := range tmp.FilterNodes {
		filterTypes := []Filter{&JSONFilter{}}
		var alreadyFound bool
		for _, ft := range filterTypes {
			if err := fn.Decode(ft); err == nil {
				if alreadyFound {
					return errors.New("output spec contains filter with ambiguous type")
				}
				filters = append(filters, ft)
				alreadyFound = true
			}
		}
	}
	if len(filters) == 0 {
		return errors.New("no valid filters found in output spec")
	}
	s.Filters = filters
	return nil
}

// Apply applies this filters to the target string
// and produces a new string
func (f *JSONFilter) Apply(inStr string) (string, error) {
	result := gjson.Get(inStr, f.Path)
	if !result.Exists() {
		return "", fmt.Errorf("json path not found: %v", f.Path)
	}
	return result.String(), nil
}

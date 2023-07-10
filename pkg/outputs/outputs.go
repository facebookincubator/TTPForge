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

type OutputFilter interface {
	Apply(inStr string) (string, error)
}

type OutputSpec struct {
	Filters []OutputFilter `yaml:"filters"`
}

func (o *OutputSpec) Apply(inStr string) (string, error) {
	var err error
	curStr := inStr
	for _, f := range o.Filters {
		curStr, err = f.Apply(curStr)
		if err != nil {
			return "", err
		}
	}
	return curStr, nil
}

type JSONFilter struct {
	Path string `yaml:"json_path"`
}

func (o *OutputSpec) UnmarshalYAML(node *yaml.Node) error {
	type SpecTmp struct {
		FilterNodes []yaml.Node `yaml:"filters"`
	}

	var tmp SpecTmp
	if err := node.Decode(&tmp); err != nil {
		return err
	}

	var filters []OutputFilter
	for _, fn := range tmp.FilterNodes {
		filterTypes := []OutputFilter{&JSONFilter{}}
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
	o.Filters = filters
	return nil
}

func (f *JSONFilter) Apply(inStr string) (string, error) {
	result := gjson.Get(inStr, f.Path)
	if !result.Exists() {
		return "", fmt.Errorf("json path not found: %v", f.Path)
	}
	return result.String(), nil
}

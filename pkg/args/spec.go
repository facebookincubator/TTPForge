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

package args

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Spec struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type,omitempty"`
	Default  string `yaml:"default,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

func (spec Spec) validate() error {
	if spec.Name == "" {
		return errors.New("argument name cannot be empty")
	}
	return nil
}

func (spec Spec) checkArgType(val string) error {
	switch spec.Type {
	case "", "string":
		//string is the default
		if val == "" {
			return errors.New("value provided for string argument cannot be empty")
		}
	case "int":
		if _, err := strconv.Atoi(val); err != nil {
			return errors.New("non-integer value provided")
		}
	default:
		return fmt.Errorf("invalid type %v specified in configuration for argument %v", spec.Type, spec.Name)
	}
	return nil
}

func ParseAndValidate(specs []Spec, argsKvStrs []string) (map[string]string, error) {

	// validate the specs
	specsByName := make(map[string]Spec)
	for _, spec := range specs {
		if err := spec.validate(); err != nil {
			return nil, err
		}
		if _, ok := specsByName[spec.Name]; ok {
			return nil, fmt.Errorf("duplicate argument name: %v", spec.Name)
		} else {
			specsByName[spec.Name] = spec
		}
	}

	// validate the inputs
	processedArgs := make(map[string]string)
	for _, argKvStr := range argsKvStrs {
		argKv := strings.SplitN(argKvStr, "=", 2)
		if len(argKv) != 2 {
			return nil, fmt.Errorf("invalid argument specification string: %v", argKvStr)
		}
		argName := argKv[0]
		argVal := argKv[1]

		// passed foo=bar with no argument foo defined in specs
		spec, ok := specsByName[argName]
		if !ok {
			return nil, fmt.Errorf("received unexpected argument: %v ", argName)
		}

		// passed argument value of invalid type
		err := spec.checkArgType(argVal)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid value '%v' specified for argument '%v': %v",
				argVal,
				argName,
				err,
			)
		}

		// valid arg value - save
		processedArgs[argName] = argVal
	}
	return processedArgs, nil
}

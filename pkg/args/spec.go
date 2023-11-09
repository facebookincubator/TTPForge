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
	"regexp"
	"strconv"
	"strings"
)

// Spec defines a CLI argument for the TTP
type Spec struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type,omitempty"`
	Default string   `yaml:"default,omitempty"`
	Choices []string `yaml:"choices,omitempty"`
	Format  string   `yaml:"regexp,omitempty"`

	formatReg *regexp.Regexp
}

// ParseAndValidate checks that the provided arguments
// match the argument specifications for this TTP
//
// **Parameters:**
//
// specs: slice of argument Spec values loaded from the TTP yaml
// argKvStrs: slice of arguments in "ARG_NAME=ARG_VALUE" format
//
// **Returns:**
//
// map[string]string: the parsed and validated argument key-value pairs
// error: an error if there is a problem
func ParseAndValidate(specs []Spec, argsKvStrs []string) (map[string]any, error) {

	// validate the specs
	processedArgs := make(map[string]any)
	specsByName := make(map[string]Spec)
	for _, spec := range specs {
		if spec.Name == "" {
			return nil, errors.New("argument name cannot be empty")
		}

		err := spec.validateChoiceTypes()
		if err != nil {
			return nil, fmt.Errorf("failed to validate types of choice values: %w", err)
		}

		// set the default value, will be overwritten by passed value
		if spec.Default != "" {
			if !spec.isValidChoice(spec.Default) {
				return nil, fmt.Errorf("invalid default value: %v, allowed values: %v ", spec.Default, strings.Join(spec.Choices, ", "))
			}

			defaultVal, err := spec.convertArgToType(spec.Default)
			if err != nil {
				return nil, fmt.Errorf("default value type does not match spec: %w", err)
			}
			processedArgs[spec.Name] = defaultVal
		}

		// set Format to match whole string
		// check if first and last character are ^ and $ respectively
		// append and prepend if missing
		// if Format string is missing ^$ then we are subject to partial matches
		if spec.Format != "" {
			if spec.Type != "string" {
				return nil, fmt.Errorf("`regexp:` can only be used with string arguments")
			}
			if spec.Format[0] != '^' {
				spec.Format = "^" + spec.Format
			}
			if spec.Format[len(spec.Format)-1] != '$' {
				spec.Format = spec.Format + "$"
			}
			spec.formatReg, err = regexp.Compile(spec.Format)
			if err != nil {
				return nil, fmt.Errorf("invalid regular expression supplied to arg spec format: %w", err)
			}
		}

		if _, ok := specsByName[spec.Name]; ok {
			return nil, fmt.Errorf("duplicate argument name: %v", spec.Name)
		}
		specsByName[spec.Name] = spec
	}

	// validate the inputs
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

		if !spec.isValidChoice(argVal) {
			return nil, fmt.Errorf("received unexpected value: %v, allowed values: %v ", argVal, strings.Join(spec.Choices, ", "))
		}

		if spec.formatReg != nil && !spec.formatReg.MatchString(argVal) {
			return nil, fmt.Errorf("invalid value format: %v, expected regex format: %v ", argVal, spec.Format)
		}

		typedVal, err := spec.convertArgToType(argVal)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to process value '%v' specified for argument '%v': %v",
				argVal,
				argName,
				err,
			)
		}

		// valid arg value - save
		processedArgs[argName] = typedVal
	}

	// error if argument was not provided and no default value was specified
	for _, spec := range specs {
		if _, ok := processedArgs[spec.Name]; !ok {
			return nil, fmt.Errorf("value for required argument '%v' was not provided and no default value was specified", spec.Name)
		}
	}
	return processedArgs, nil
}

func (spec Spec) validateChoiceTypes() error {
	if len(spec.Choices) == 0 {
		return nil
	}

	for _, choice := range spec.Choices {
		_, err := spec.convertArgToType(choice)
		if err != nil {
			return err
		}
	}
	return nil
}

func (spec Spec) isValidChoice(value string) bool {
	if len(spec.Choices) == 0 {
		return true
	}

	for _, choice := range spec.Choices {
		if choice == value {
			return true
		}
	}
	return false
}

func (spec Spec) convertArgToType(val string) (any, error) {
	switch spec.Type {
	case "", "string":
		// string is the default - any string is valid
		return val, nil
	case "int":
		asInt, err := strconv.Atoi(val)
		if err != nil {
			return nil, errors.New("non-integer value provided")
		}
		return asInt, nil
	case "bool":
		asBool, err := strconv.ParseBool(val)
		if err != nil {
			return nil, errors.New("no-boolean value provided")
		}
		return asBool, nil
	default:
		return nil, fmt.Errorf("invalid type %v specified in configuration for argument %v", spec.Type, spec.Name)
	}
}

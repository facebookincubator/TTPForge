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
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/args"
	"github.com/facebookincubator/ttpforge/pkg/preprocess"
	"github.com/facebookincubator/ttpforge/pkg/targets"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// RenderTemplatedTTP is a function that utilizes Golang's `text/template` for template substitution.
// It replaces template expressions like `{{ .Args.myarg }}` with corresponding values.
// This function must be invoked prior to YAML unmarshaling, as the template syntax `{{ ... }}`
// may result in invalid YAML under specific conditions.
//
// **Parameters:**
//
// ttpStr: A string containing the TTP template to be rendered.
// execCfg: A pointer to a TTPExecutionConfig that represents the execution configuration for the TTP.
//
// **Returns:**
//
// *TTP: A pointer to the TTP object created from the template.
// error: An error if the rendering or unmarshaling process fails.
func RenderTemplatedTTP(ttpStr string, execCfg *TTPExecutionConfig) (*TTP, error) {
	tmpl, err := template.New("ttp").Parse(ttpStr)
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, execCfg)
	if err != nil {
		return nil, err
	}

	var ttp TTP
	err = yaml.Unmarshal(result.Bytes(), &ttp)
	if err != nil {
		return nil, err
	}
	return &ttp, nil
}

// LoadTTP reads a TTP file and creates a TTP instance based on its contents.
// It processes the targets and arguments present in the file, validates them,
// and populates the appropriate fields in the provided TTPExecutionConfig.
// If the file is empty or contains invalid data, it returns an error.
//
// Parameters:
// ttpFilePath: the absolute or relative path to the TTP YAML file.
// fsys: an afero.Fs that contains the specified TTP file path.
// execCfg: Configuration containing execution details which will be populated based on parsed targets and arguments.
// argsKvStrs: Key-value strings for argument specifications.
// targetsKvStrs: Key-value strings for target specifications (Currently unused but may be required for future extensions).
//
// Returns:
// ttp: Pointer to the created TTP instance, or nil if the file is empty or invalid.
// err: An error if the file contains invalid data or cannot be read.
func LoadTTP(ttpFilePath string, fsys afero.Fs, execCfg *TTPExecutionConfig, argsKvStrs []string, targetsKvStrs []string) (*TTP, error) {
	ttpBytes, err := readTTPBytes(ttpFilePath, fsys)
	if err != nil {
		return nil, err
	}

	result, err := preprocess.Parse(ttpBytes)
	if err != nil {
		return nil, err
	}

	// linting above establishes that the TTP yaml will be
	// compatible with our rendering process
	type ArgAndTargetSpecContainer struct {
		ArgSpecs   []args.Spec        `yaml:"args"`
		TargetSpec targets.TargetSpec `yaml:"targets"`
	}

	var tmpContainer ArgAndTargetSpecContainer
	err = yaml.Unmarshal(result.PreambleBytes, &tmpContainer)
	if err != nil {
		return nil, err
	}

	// Parse and validate the provided TargetSpec (if applicable)
	targetValues, err := targets.ParseAndValidateTargets(tmpContainer.TargetSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate targets: %v", err)
	}

	execCfg.Targets = targetValues

	argValues, err := args.ParseAndValidate(tmpContainer.ArgSpecs, argsKvStrs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse and validate arguments: %v", err)
	}
	execCfg.Args = argValues

	ttp, err := RenderTemplatedTTP(string(ttpBytes), execCfg)
	if err != nil {
		return nil, err
	}

	// embedded fs has no notion of workdirs
	// so we should only set workdir to the TTP's directory
	// if we are using an OsFs
	switch fsys.(type) {
	case *afero.OsFs:
		absPath, err := filepath.Abs(ttpFilePath)
		if err != nil {
			return nil, err
		}
		ttp.WorkDir = filepath.Dir(absPath)
	default:
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		ttp.WorkDir = wd
	}

	// TODO: refactor directory handling - this is in-elegant
	// but has less bugs than previous way
	for _, step := range ttp.Steps {
		step.SetDir(ttp.WorkDir)
		if cleanups := step.GetCleanup(); cleanups != nil {
			for _, c := range cleanups {
				c.SetDir(ttp.WorkDir)
			}
		}
	}

	return ttp, nil
}

func readTTPBytes(ttpFilePath string, system afero.Fs) ([]byte, error) {
	var file fs.File
	var err error
	if system == nil {
		file, err = os.Open(ttpFilePath)
	} else {
		file, err = system.Open(ttpFilePath)
	}
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

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
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/facebookincubator/ttpforge/pkg/args"
	"gopkg.in/yaml.v3"
)

var (
	stepsTopLevelKeyRegexp *regexp.Regexp
	topLevelKeyRegexp      *regexp.Regexp
)

func init() {
	stepsTopLevelKeyRegexp = regexp.MustCompile("(?m)^steps:")
	topLevelKeyRegexp = regexp.MustCompile(`(?m)^[^\s]+:`)
}

type LintResult struct {
	PreambleBytes []byte
	StepsBytes    []byte
}

func LintTTP(ttpBytes []byte) (*LintResult, error) {
	// no duplicate keys
	stepTopLevelKeyLocs := stepsTopLevelKeyRegexp.FindAllIndex(ttpBytes, -1)
	if len(stepTopLevelKeyLocs) != 1 {
		return nil, errors.New("the top-level key `steps:` should occur exactly once")
	}
	stepTopLevelKeyLoc := stepTopLevelKeyLocs[0]

	// `steps:` should always be the last top-level key
	topLevelKeyLocs := topLevelKeyRegexp.FindAllIndex(ttpBytes, -1)
	for _, loc := range topLevelKeyLocs {
		if loc[0] > stepTopLevelKeyLoc[0] {
			return nil, errors.New("the top-level key `steps:` should always be the last top-level key in the file")
		}
	}
	return &LintResult{
		PreambleBytes: ttpBytes[:stepTopLevelKeyLoc[0]],
		StepsBytes:    ttpBytes[stepTopLevelKeyLoc[0]:stepTopLevelKeyLoc[1]],
	}, nil
}

// RenderTemplatedTTP uses Golang's `text/template` to substitute template
// expressions such as `{{ .Args.myarg }}` with their appropriate values
// This function should always be called before YAML unmarshaling since
// the template syntax `{{ ... }}` may be invalid yaml in certain circumstances
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
// If the file is empty or contains invalid data, it returns an error.
//
// **Parameters:**
//
// ttpFilePath: the absolute or relative path to the TTP file.
// system: An optional fs.StatFS from which to load the TTP
//
// **Returns:**
//
// ttp: Pointer to the created TTP instance, or nil if the file is empty or invalid.
// err: An error if the file contains invalid data or cannot be read.
func LoadTTP(ttpFilePath string, system fs.StatFS, execCfg *TTPExecutionConfig, argsKvStrs []string) (*TTP, error) {

	ttpBytes, err := readTTPBytes(ttpFilePath, system)
	if err != nil {
		return nil, err
	}

	result, err := LintTTP(ttpBytes)
	if err != nil {
		return nil, err
	}

	// linting above establishes that the TTP yaml will be
	// compatible with our rendering process
	type ArgSpecContainer struct {
		ArgSpecs []args.Spec `yaml:"args"`
	}
	var tmpContainer ArgSpecContainer
	err = yaml.Unmarshal(result.PreambleBytes, &tmpContainer)
	if err != nil {
		return nil, err
	}

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
	if system == nil {
		absPath, err := filepath.Abs(ttpFilePath)
		if err != nil {
			return nil, err
		}
		ttp.WorkDir = filepath.Dir(absPath)
	} else {
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

func readTTPBytes(ttpFilePath string, system fs.StatFS) ([]byte, error) {
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

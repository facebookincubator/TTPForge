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
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var stepsTopLevelKeyRegexp *regexp.Regexp

func init() {
	stepsTopLevelKeyRegexp = regexp.MustCompile("(?m)^steps:")
}

func LintTTP(ttpBytes []byte) error {
	// `steps:` should always be the last top-level key
	stepTopLevelKeyLocs := stepsTopLevelKeyRegexp.FindAllIndex(ttpBytes, -1)
	if len(stepTopLevelKeyLocs) != 1 {
		return errors.New("the top level key `steps:` should occur exactly once")
	}
	return nil
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
func LoadTTP(ttpFilePath string, system fs.StatFS, execCfg *TTPExecutionConfig) (*TTP, error) {

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

	ttp, err := RenderTemplatedTTP(string(contents), execCfg)
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

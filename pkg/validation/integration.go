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

package validation

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/platforms"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// readTTPBytesForValidation reads TTP file contents
func readTTPBytesForValidation(ttpFilePath string, fsys afero.Fs) ([]byte, error) {
	var file afero.File
	var err error

	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	file, err = fsys.Open(ttpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return contents, nil
}

// renderTemplatedTTPForValidation is similar to blocks.RenderTemplatedTTP
// but is more lenient for validation purposes
func renderTemplatedTTPForValidation(ttpStr string, rp blocks.RenderParameters) (*blocks.TTP, error) {
	tmpl, err := template.New("ttp").Funcs(sprig.TxtFuncMap()).Parse(ttpStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, rp)
	if err != nil {
		logging.L().Warnf("Template execution failed (this may be expected if args are not provided): %v", err)
		var ttp blocks.TTP
		err = yaml.Unmarshal([]byte(ttpStr), &ttp)
		if err != nil {
			return nil, fmt.Errorf("failed to decode TTP YAML after template error: %w", err)
		}
		return &ttp, nil
	}

	var ttp blocks.TTP
	err = yaml.Unmarshal(result.Bytes(), &ttp)
	if err != nil {
		logging.L().Errorf("Failed to decode TTP YAML - received error: %v", err)
		logging.L().Error("Inspect the rendered TTP below:\n", result.String())
		return nil, fmt.Errorf("YAML unmarshal failed: %w", err)
	}

	return &ttp, nil
}

// ValidateIntegration attempts validation using the blocks package
// Template errors are converted to warnings
// Note: Preamble validation is now done in ValidatePreamble, so this
// function focuses on step-level validation
func ValidateIntegration(ttpFilePath string, ttpBytes []byte, repo repos.Repo, result *Result) {
	rp := blocks.RenderParameters{
		Args:     map[string]any{},
		Platform: platforms.GetCurrentPlatformSpec(),
	}

	ttp, err := renderTemplatedTTPForValidation(string(ttpBytes), rp)
	if err != nil {
		if strings.Contains(err.Error(), "template") {
			result.AddWarning(fmt.Sprintf("Template rendering (may be expected with empty args): %v", err))
		} else {
			result.AddError(fmt.Sprintf("TTP rendering: %v", err))
		}
		return
	}

	execCtx := blocks.NewTTPExecutionContext()
	execCtx.Cfg.Repo = repo

	absPath, err := filepath.Abs(ttpFilePath)
	if err == nil {
		ttp.WorkDir = filepath.Dir(absPath)
		execCtx.Vars.WorkDir = ttp.WorkDir
	}

	for idx, step := range ttp.Steps {
		stepCopy := step

		if err := stepCopy.Validate(execCtx); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "template") || strings.Contains(errStr, "{{") {
				result.AddWarning(fmt.Sprintf("Step #%d (%s) template validation: %v", idx+1, step.Name, err))
			} else if strings.Contains(errStr, "executable file not found in $PATH") {
				result.AddWarning(fmt.Sprintf("Step #%d (%s) executor not available on this system: %v", idx+1, step.Name, err))
			} else if strings.Contains(errStr, "exec:") && strings.Contains(errStr, "not found") {
				result.AddWarning(fmt.Sprintf("Step #%d (%s) command not available on this system: %v", idx+1, step.Name, err))
			} else {
				result.AddError(fmt.Sprintf("Step #%d (%s) validation: %v", idx+1, step.Name, err))
			}
		}
	}
}

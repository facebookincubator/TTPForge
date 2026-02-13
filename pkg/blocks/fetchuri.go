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
	"io"
	"net/http"
	"net/url"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// FetchURIStep represents a step in a process that consists of a main action,
// a cleanup action, and additional metadata.
type FetchURIStep struct {
	actionDefaults `yaml:",inline"`
	FetchURI       string   `yaml:"fetch_uri,omitempty"`
	Retries        string   `yaml:"retries,omitempty"`
	Location       string   `yaml:"location,omitempty"`
	Proxy          string   `yaml:"proxy,omitempty"`
	Overwrite      bool     `yaml:"overwrite,omitempty"`
	FileSystem     afero.Fs `yaml:"-,omitempty"`
}

// NewFetchURIStep creates a new FetchURIStep instance and returns a pointer to it.
func NewFetchURIStep() *FetchURIStep {
	return &FetchURIStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (f *FetchURIStep) IsNil() bool {
	switch {
	case f.FetchURI == "":
		return true
	case f.Location == "":
		return true
	default:
		return false
	}
}

// Validate validates the FetchURIStep. It checks that the
// Act field is valid, Location is set with
// a valid file path, and Uri is set.
//
// If Location is set, it ensures that the path exists and retrieves
// its absolute path.
func (f *FetchURIStep) Validate(execCtx TTPExecutionContext) error {
	// Validate URI exists
	if f.FetchURI == "" {
		return fmt.Errorf("require FetchURI to be set with fetchURI")
	}

	// Validate location exists
	if f.Location == "" {
		return fmt.Errorf("require Location to be set with fetchURI")
	}

	// Validate Proxy is valid URI
	if f.Proxy != "" && !execCtx.containsStepTemplating(f.Proxy) {
		err := f.validateProxy()
		if err != nil {
			return err
		}
	}

	// Retrieve the absolute path to the file, if location doesn't contain templating
	if !execCtx.containsStepTemplating(f.Location) {
		err := f.validateLocation(execCtx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (f *FetchURIStep) Template(execCtx TTPExecutionContext) error {
	var err error

	// Template URI
	f.FetchURI, err = execCtx.templateStep(f.FetchURI)
	if err != nil {
		return err
	}

	// Template retries
	f.Retries, err = execCtx.templateStep(f.Retries)
	if err != nil {
		return err
	}

	// Template and revalidate location
	if execCtx.containsStepTemplating(f.Location) {
		f.Location, err = execCtx.templateStep(f.Location)
		if err != nil {
			return err
		}
		err = f.validateLocation(execCtx)
		if err != nil {
			return err
		}
	}

	// Template and revalidate proxy
	if execCtx.containsStepTemplating(f.Proxy) {
		f.Proxy, err = execCtx.templateStep(f.Proxy)
		if err != nil {
			return err
		}
		err = f.validateProxy()
		if err != nil {
			return err
		}
	}

	return nil
}

// Execute runs the step and returns an error if one occurs.
func (f *FetchURIStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	logging.L().Info("========= Executing ==========")
	logging.L().Infof("FetchURI: %s", f.FetchURI)
	if err := f.fetchURI(execCtx); err != nil {
		logging.L().Error(zap.Error(err))
		return nil, err
	}

	logging.L().Info("========= Result ==========")
	logging.L().Infof("Fetched URI to location: %s", f.Location)
	// Send file contents to the output variable
	if f.OutputVar != "" {
		// TODO: maybe we make this step able to just send the output to a var and make the file output optional?
		content, err := afero.ReadFile(f.FileSystem, f.Location)
		if err != nil {
			logging.L().Error(zap.Error(err))
			return nil, err
		}
		execCtx.Vars.StepVars[f.OutputVar] = string(content)
	}

	return &ActResult{}, nil
}

// fetchURI executes the FetchURIStep with the specified Location, Uri, and additional arguments,
// and an error if any errors occur.
func (f *FetchURIStep) fetchURI(execCtx TTPExecutionContext) error {
	appFs := f.FileSystem
	absLocal := f.Location

	if appFs == nil {
		var err error
		appFs = afero.NewOsFs()
		absLocal, err = FetchAbs(f.Location, execCtx.Vars.WorkDir)
		if err != nil {
			return err
		}
	}

	if ok, _ := afero.Exists(appFs, absLocal); ok && !f.Overwrite {
		logging.L().Errorw("location exists, remove and retry", "location", absLocal)
		return fmt.Errorf("location [%s] exists and overwrite is set to false. remove and retry", f.Location)
	}

	client := http.DefaultClient
	if f.Proxy != "" && !execCtx.Cfg.NoProxy {
		proxyURI, err := url.Parse(f.Proxy)
		if err != nil {
			return err
		} else if proxyURI.Host == "" || proxyURI.Scheme == "" {
			return fmt.Errorf("invalid URI given for Proxy: %s", f.Proxy)
		}
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxyURI),
		}
		client = &http.Client{Transport: tr}
	}

	resp, err := client.Get(f.FetchURI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fHandle, err := appFs.Create(absLocal)
	if err != nil {
		return err
	}
	defer fHandle.Close()

	_, err = io.Copy(fHandle, resp.Body)
	if err != nil {
		return err
	}

	logging.L().Debugw("wrote contents of URI to specified location", "location", absLocal, "uri", f.FetchURI)

	return nil
}

// GetDefaultCleanupAction will instruct the calling code
// to remove the file fetched by this action.
func (f *FetchURIStep) GetDefaultCleanupAction() Action {
	return &RemovePathAction{
		Path: f.Location,
	}
}

// validateProxy validates if the proxy is a valid uri and returns an error if validate fails, otherwise returns nil.
func (f *FetchURIStep) validateProxy() error {
	uri, err := url.Parse(f.Proxy)
	if err != nil {
		return err
	} else if uri.Host == "" || uri.Scheme == "" {
		return fmt.Errorf("invalid URI given for Proxy: %s", f.Proxy)
	}
	return nil
}

// validateLocation validates that the location is a valid path, and isn't overriding an existing
// file unless explicitly stated.  Returns an error if validation fails, otherwise returns nil.
func (f *FetchURIStep) validateLocation(execCtx TTPExecutionContext) error {
	fsys := f.FileSystem
	if fsys == nil {
		fsys = afero.NewOsFs()
	}

	absLocal, err := FetchAbs(f.Location, execCtx.Vars.WorkDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for location: %w", err)
	}
	exists, err := afero.Exists(fsys, absLocal)
	if err != nil {
		return fmt.Errorf("error checking if location exists (location: %q): %w", absLocal, err)
	}
	if exists && !f.Overwrite {
		return fmt.Errorf("file exists at location %q and overwrite is not enabled", absLocal)
	}
	return nil
}

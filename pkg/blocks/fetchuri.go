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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// FetchURIStep represents a step in a process that consists of a main action,
// a cleanup action, and additional metadata.
type FetchURIStep struct {
	*Act         `yaml:",inline"`
	FetchURI     string     `yaml:"fetch_uri,omitempty"`
	Retries      string     `yaml:"retries,omitempty"`
	Location     string     `yaml:"location,omitempty"`
	Proxy        string     `yaml:"proxy,omitempty"`
	Overwrite    bool       `yaml:"overwrite,omitempty"`
	IgnoreErrors bool       `yaml:"ignore_errors,omitempty,flow"`
	CleanupStep  CleanupAct `yaml:"cleanup,omitempty,flow"`
	FileSystem   afero.Fs   `yaml:"-,omitempty"`
}

// NewFetchURIStep creates a new FetchURIStep instance and returns a pointer to it.
func NewFetchURIStep() *FetchURIStep {
	return &FetchURIStep{
		Act: &Act{
			Type: StepFetchURI,
		},
	}
}

// UnmarshalYAML decodes a YAML node into a FetchURIStep instance. It uses
// the provided struct as a template for the YAML data, and initializes the
// FetchURIStep instance with the decoded values.
//
// **Parameters:**
//
// node: A pointer to a yaml.Node representing the YAML data to decode.
//
// **Returns:**
//
// error: An error if there is a problem decoding the YAML data.
func (u *FetchURIStep) UnmarshalYAML(node *yaml.Node) error {

	type fetchURITmpl struct {
		Act          `yaml:",inline"`
		FetchURI     string    `yaml:"fetch_uri,omitempty"`
		Retries      string    `yaml:"retries,omitempty"`
		Location     string    `yaml:"location,omitempty"`
		Proxy        string    `yaml:"proxy,omitempty"`
		Overwrite    bool      `yaml:"overwrite,omitempty"`
		IgnoreErrors bool      `yaml:"ignore_errors,omitempty,flow"`
		CleanupStep  yaml.Node `yaml:"cleanup,omitempty,flow"`
	}

	// Decode the YAML node into the provided template.
	var tmpl fetchURITmpl
	if err := node.Decode(&tmpl); err != nil {
		return err
	}

	// Initialize the FetchURIStep instance with the decoded values.
	u.Act = &tmpl.Act
	u.FetchURI = tmpl.FetchURI
	u.Location = tmpl.Location
	u.Retries = tmpl.Retries
	u.Proxy = tmpl.Proxy
	u.Overwrite = tmpl.Overwrite
	u.IgnoreErrors = tmpl.IgnoreErrors

	// Check for invalid steps.
	if u.IsNil() {
		return u.ExplainInvalid()
	}

	// If there is no cleanup step or if this step is the cleanup step, exit.
	if tmpl.CleanupStep.IsZero() || u.Type == StepCleanup {
		return nil
	}

	// Create a CleanupStep instance and add it to the FetchURIStep instance.
	logging.L().Debugw("step", "name", tmpl.Name)
	cleanup, err := u.MakeCleanupStep(&tmpl.CleanupStep)
	logging.L().Debugw("step", zap.Error(err))
	if err != nil {
		logging.L().Errorw("error creating cleanup step", zap.Error(err))
		return err
	}

	u.CleanupStep = cleanup

	return nil
}

// GetType returns the type of the step as StepType.
func (u *FetchURIStep) GetType() StepType {
	return StepFetchURI
}

// Cleanup is a method to establish a link with the Cleanup interface.
// Assumes that the type is the cleanup step and is invoked by
// u.CleanupStep.Cleanup.
func (u *FetchURIStep) Cleanup(execCtx TTPExecutionContext) (*ActResult, error) {
	result, err := u.Execute(execCtx)
	if err != nil {
		return nil, err
	}
	return &result.ActResult, err
}

// GetCleanup returns a slice of CleanupAct if the CleanupStep is not nil.
func (u *FetchURIStep) GetCleanup() []CleanupAct {
	if u.CleanupStep != nil {
		return []CleanupAct{u.CleanupStep}
	}
	return []CleanupAct{}
}

// ExplainInvalid returns an error message explaining why the FetchURIStep
// is invalid.
//
// **Returns:**
//
// error: An error message explaining why the FetchURIStep is invalid.
func (u *FetchURIStep) ExplainInvalid() error {
	var err error
	if u.FetchURI == "" {
		err = errors.New("empty FetchURI provided")
	}

	if u.Location == "" && err != nil {
		err = errors.New("empty Location provided")
	}

	if u.Name != "" && err != nil {
		err = fmt.Errorf("[!] invalid FetchURIStep: [%s] %v", u.Name, zap.Error(err))
	}

	return err
}

// IsNil checks if the FetchURIStep is nil or empty and returns a boolean value.
func (u *FetchURIStep) IsNil() bool {
	switch {
	case u.Act.IsNil():
		return true
	case u.FetchURI == "":
		return true
	case u.Location == "":
		return true
	default:
		return false
	}
}

// Execute runs the FetchURIStep and returns an error if any occur.
func (u *FetchURIStep) Execute(execCtx TTPExecutionContext) (*ExecutionResult, error) {
	logging.L().Info("========= Executing ==========")

	if err := u.fetchURI(execCtx); err != nil {
		logging.L().Error(zap.Error(err))
		if u.IgnoreErrors {
			logging.L().Warn("Error ignored due to 'ignore_errors' parameter")
			return &ExecutionResult{}, nil
		}
		return nil, err
	}

	logging.L().Info("========= Result ==========")

	return &ExecutionResult{}, nil
}

// fetchURI executes the FetchURIStep with the specified Location, Uri, and additional arguments,
// and an error if any errors occur.
func (u *FetchURIStep) fetchURI(execCtx TTPExecutionContext) error {
	appFs := u.FileSystem
	absLocal := u.Location

	if appFs == nil {
		var err error
		appFs = afero.NewOsFs()
		absLocal, err = FetchAbs(u.Location, u.WorkDir)
		if err != nil {
			return err
		}
	}

	if ok, _ := afero.Exists(appFs, absLocal); ok && !u.Overwrite {
		logging.L().Errorw("location exists, remove and retry", "location", absLocal)
		return fmt.Errorf("location [%s] exists and overwrite is set to false. remove and retry", u.Location)
	}

	client := http.DefaultClient
	if u.Proxy != "" {
		proxyURI, err := url.Parse(u.Proxy)
		if err != nil {
			return err
		} else if proxyURI.Host == "" || proxyURI.Scheme == "" {
			return fmt.Errorf("invalid URI given for Proxy: %s", u.Proxy)
		}
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxyURI),
		}
		client = &http.Client{Transport: tr}
	}

	resp, err := client.Get(u.FetchURI)
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

	logging.L().Debugw("wrote contents of URI to specified location", "location", absLocal, "uri", u.FetchURI)

	return nil
}

// Validate validates the FetchURIStep. It checks that the
// Act field is valid, Location is set with
// a valid file path, and Uri is set.
//
// If Location is set, it ensures that the path exists and retrieves
// its absolute path.
//
// **Returns:**
//
// error: An error if any validation checks fail.
func (u *FetchURIStep) Validate(execCtx TTPExecutionContext) error {
	if err := u.Act.Validate(); err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	if u.FetchURI == "" {
		err := errors.New("require FetchURI to be set with fetchURI")
		logging.L().Error(zap.Error(err))
		return err
	}

	if u.Location == "" {
		err := errors.New("require Location to be set with fetchURI")
		logging.L().Error(zap.Error(err))
		return err
	}

	if u.Proxy != "" {
		uri, err := url.Parse(u.Proxy)
		if err != nil {
			return err
		} else if uri.Host == "" || uri.Scheme == "" {
			return fmt.Errorf("invalid URI given for Proxy: %s", u.Proxy)
		}
	}

	// Retrieve the absolute path to the file.
	absLocal, err := FetchAbs(u.Location, u.WorkDir)
	if err != nil {
		logging.L().Error(zap.Error(err))
		return err
	}

	_, err = os.Stat(absLocal)
	if !errors.Is(err, fs.ErrNotExist) && !u.Overwrite {
		logging.L().Errorw("FetchURI location exists, remove and retry", "location", absLocal)
		return errors.New("file exists at location specified, remove and retry")
	}

	if u.CleanupStep != nil {
		if err := u.CleanupStep.Validate(execCtx); err != nil {
			logging.L().Errorw("error validating cleanup step", zap.Error(err))
			return err
		}
	}

	return nil
}

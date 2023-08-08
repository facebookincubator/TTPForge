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

package repos

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RepoConfigFileName is the expected name of the configuration
// file for a TTP repository such as ForgeArmory.
// We export it for tests
const RepoConfigFileName = "ttpforge-repo-config.yaml"

// Spec defines the fields that are expected
// to be set in the program-wide configuration file
// in order to add a given repository folder to
// the TTPForge search path for TTPs, templates, etc
type Spec struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Config contains all the fields
// used by higher-level code to search this repository for
// any items of interest.
// The []Spec entry in the program-wide configuration tells
// TTPForge which Config entries to create.
type Config struct {
	TTPSearchPaths []string `yaml:"ttp_search_paths"`
}

// LoadConfigs searches the pat file the provided `specs`
// for repository configuration files
func LoadConfigs(fsys fs.FS, specs []Spec) ([]Config, error) {

	var repoConfigs []Config
	for _, spec := range specs {
		configPath := filepath.Join(spec.Path, RepoConfigFileName)
		contents, err := fs.ReadFile(fsys, configPath)
		if err != nil {
			return nil, fmt.Errorf("could not read repo config at path %v: %v", configPath, err)
		}
		var curRepoConfig Config
		err = yaml.Unmarshal(contents, &curRepoConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid config file found at %v: %v", configPath, err)
		}
		repoConfigs = append(repoConfigs, curRepoConfig)
	}
	return repoConfigs, nil
}

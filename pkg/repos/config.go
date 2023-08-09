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
	"os"
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

type Repo interface {
	FindTTP(ttpPath string) (string, error)
}

// Config contains all the fields
// used by higher-level code to search this repository for
// any items of interest.
// The []Spec entry in the program-wide configuration tells
// TTPForge which Config entries to create.
type repo struct {
	fsys           fs.StatFS
	basePath       string
	TTPSearchPaths []string `yaml:"ttp_search_paths"`
}

func (r *repo) FindTTP(ttpPath string) (string, error) {
	return r.search(r.TTPSearchPaths, ttpPath)
}

func (r *repo) search(dirsToSearch []string, relPath string) (string, error) {
	for _, dirToSearch := range dirsToSearch {
		fullPath := filepath.Join(r.basePath, dirToSearch, relPath)
		if _, err := r.fsys.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return "", err
			}
		} else {
			return fullPath, nil
		}
	}
	return "", nil
}

// LoadConfigs searches the pat file the provided `specs`
// for repository configuration files
func (spec *Spec) Load(fsys fs.StatFS) (Repo, error) {

	configPath := filepath.Join(spec.Path, RepoConfigFileName)
	contents, err := fs.ReadFile(fsys, configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read repo config at path %v: %v", configPath, err)
	}
	var r repo
	err = yaml.Unmarshal(contents, &r)
	if err != nil {
		return nil, fmt.Errorf("invalid config file found at %v: %v", configPath, err)
	}
	r.fsys = fsys
	r.basePath = spec.Path
	return &r, nil
}

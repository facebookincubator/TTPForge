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
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/fileutils"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

const (
	// RepoConfigFileName is the expected name of the configuration
	// file for a TTP repository such as ForgeArmory.
	// We export it for tests
	RepoConfigFileName = "ttpforge-repo-config.yaml"

	// RepoPrefixSep divides the repo name reference from the TTP/template/etc path
	RepoPrefixSep = "//"
)

// Spec defines the fields that are expected
// to be set in the program-wide configuration file
// in order to add a given repository folder to
// the TTPForge search path for TTPs, templates, etc
type Spec struct {
	Name string    `yaml:"name"`
	Path string    `yaml:"path"`
	Git  GitConfig `yaml:"git"`
}

// GitConfig provides instructions for cloning a repo
type GitConfig struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch,omitempty"`
}

// Repo provides an interface for finding TTPs and templates
// from a repo such as ForgeArmory
type Repo interface {
	ListTTPs() ([]string, error)
	FindTTP(ttpRef string) (string, error)
	FindTemplate(templatePath string) (string, error)
	GetFs() afero.Fs
	GetName() string
	GetFullPath() string
	GetTTPSearchPaths() []string
	GetTemplateSearchPaths() []string
}

// Config contains all the fields
// used by higher-level code to search this repository for
// any items of interest.
// The []Spec entry in the program-wide configuration tells
// TTPForge which Config entries to create.
type repo struct {
	fsys                afero.Fs
	fullPath            string
	spec                Spec
	TTPSearchPaths      []string `yaml:"ttp_search_paths"`
	TemplateSearchPaths []string `yaml:"template_search_paths"`
}

// GetTTPSearchPaths returns the absolute paths of all configured TTP search directories
func (r *repo) GetTTPSearchPaths() []string {
	absPaths := make([]string, len(r.TTPSearchPaths))
	for i, path := range r.TTPSearchPaths {
		absPaths[i] = filepath.Join(r.fullPath, path)
	}
	return absPaths
}

// GetTemplateSearchPaths returns the absolute paths of all configured template search directories
func (r *repo) GetTemplateSearchPaths() []string {
	absPaths := make([]string, len(r.TemplateSearchPaths))
	for i, path := range r.TemplateSearchPaths {
		absPaths[i] = filepath.Join(r.fullPath, path)
	}
	return absPaths
}

// ListsTTPs lists the TTPs in this repo
func (r *repo) ListTTPs() ([]string, error) {
	return r.list(r.TTPSearchPaths)
}

// FindTTP locates a TTP if it exists in this repo
func (r *repo) FindTTP(ttpRef string) (string, error) {
	var repoName, ttpPath string
	if strings.Contains(ttpRef, RepoPrefixSep) {
		// a TTP reference ([repo]//path/to/ttp)
		// SubTTPs use this path
		tokens := strings.SplitN(ttpRef, RepoPrefixSep, 2)
		repoName, ttpPath = tokens[0], tokens[1]
		if repoName != "" && repoName != r.spec.Name {
			return "", fmt.Errorf("invalid TTP reference %q; repo name %q does not match %q", ttpRef, repoName, r.spec.Name)
		}
	} else {
		// a regular YAML file path
		ttpPath = ttpRef
	}
	return r.search(r.TTPSearchPaths, ttpPath)
}

// FindTemplate locates a template if it exists in this repo
func (r *repo) FindTemplate(templatePath string) (string, error) {
	return r.search(r.TemplateSearchPaths, templatePath)
}

// GetFs is a convenience function principally used by SubTTPs
func (r *repo) GetFs() afero.Fs {
	return r.fsys
}

// GetName returns the repos name
func (r *repo) GetName() string {
	return r.spec.Name
}

// GetFullPath returns the repos full path
// including the basePath that was passed
// when it was constructed
func (r *repo) GetFullPath() string {
	return r.fullPath
}

func (r *repo) search(dirsToSearch []string, relPath string) (string, error) {
	for _, dirToSearch := range dirsToSearch {
		candidateFullPath := filepath.Join(r.fullPath, dirToSearch, relPath)
		if _, err := r.fsys.Stat(candidateFullPath); err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return "", err
			}
		} else {
			return candidateFullPath, nil
		}
	}
	return "", fmt.Errorf("path %v not found in repo %v", relPath, r.spec.Name)
}

func (r *repo) list(dirsToList []string) ([]string, error) {
	var allResults []string
	splitOnSep := string(os.PathSeparator)
	for _, dirToList := range dirsToList {
		prefix := filepath.Join(r.fullPath, dirToList)
		err := afero.Walk(r.fsys, prefix, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".yaml") {
				trimmed := strings.TrimPrefix(path, prefix)
				trimmedAgain := strings.TrimPrefix(trimmed, splitOnSep)
				tokens := strings.Split(trimmedAgain, splitOnSep)
				result := r.spec.Name + RepoPrefixSep + strings.Join(tokens, "/")
				allResults = append(allResults, result)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

	}
	return allResults, nil
}

// Load will clone a repository if necessary and valdiate
// its configuration, making it usable to lookup TTPs
func (spec *Spec) Load(fsys afero.Fs, basePath string) (Repo, error) {

	// validate spec fields
	if spec.Name == "" {
		return nil, errors.New("repository field `name:` cannot be empty")
	}
	if spec.Path == "" {
		return nil, errors.New("repository field `path:` cannot be empty")
	}

	var repoPath string
	isAbs, err := fileutils.IsAbs(spec.Path)
	if err != nil {
		return nil, fmt.Errorf("error checking whether %q is an absolute path", spec.Path)
	}
	if isAbs {
		repoPath, err = fileutils.ExpandPath(spec.Path)
		if err != nil {
			return nil, fmt.Errorf("could not expand path %q", spec.Path)
		}
	} else {
		repoPath = filepath.Join(basePath, spec.Path)
	}

	err = spec.ensurePresentAtPath(fsys, repoPath)
	if err != nil {
		return nil, err
	}

	repoConfigPath := filepath.Join(repoPath, RepoConfigFileName)
	contents, err := afero.ReadFile(fsys, repoConfigPath)
	if err != nil {
		return nil, fmt.Errorf("could not read repo config at path %v: %w", repoConfigPath, err)
	}
	var r repo
	err = yaml.Unmarshal(contents, &r)
	if err != nil {
		return nil, fmt.Errorf("invalid config file found at %v: %w", repoConfigPath, err)
	}
	r.fsys = fsys
	r.fullPath = repoPath
	r.spec = *spec
	return &r, nil
}

func (spec *Spec) ensurePresentAtPath(fsys afero.Fs, repoPath string) error {
	// if repo is present we can return early
	exists, err := afero.Exists(fsys, repoPath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if spec.Git.URL == "" {
		return fmt.Errorf(
			"repo at %q not found - clone manually or see docs for how to add git clone instructions",
			repoPath,
		)
	}

	branchName := spec.Git.Branch
	if branchName == "" {
		branchName = "main"
	}

	gitCmd := exec.Command("git", "clone", "--single-branch", "--branch", branchName, spec.Git.URL, repoPath)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	err = gitCmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone repo to %q: %w", repoPath, err)
	}
	return nil
}

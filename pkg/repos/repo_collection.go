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
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// RepoCollection provides useful methods for resolving and
// navigating TTPs stored in various different repositories
type RepoCollection interface {
	GetRepo(repoName string) (Repo, error)
	ResolveTTPRef(ttpRef string) (Repo, string, error)
	ListTTPs() ([]string, error)
}

type repoCollection struct {
	repos       []Repo
	reposByName map[string]Repo
	fsys        afero.Fs
}

// NewRepoCollection validates the provided repo specs
// and assembles them into a RepoCollection
//
// **Parameters:**
//
// fsys: base file system (used for unit testing)
// specs: a list of repo.Spec entries (usually from the config file)
//
// **Returns:**
//
// RepoCollection: assembled RepoCollection, or nil if there was an error
// error: an error if there is a problem
func NewRepoCollection(fsys afero.Fs, specs []Spec, basePath string) (RepoCollection, error) {
	// load all repos
	rc := repoCollection{
		fsys:        fsys,
		reposByName: make(map[string]Repo),
	}
	for _, spec := range specs {
		r, err := spec.Load(fsys, basePath)
		if err != nil {
			return nil, err
		}
		repoName := r.GetName()
		if _, found := rc.reposByName[repoName]; !found {
			rc.reposByName[repoName] = r
		} else {
			return nil, fmt.Errorf("duplicate repo name: %v", repoName)
		}
		rc.repos = append(rc.repos, r)
	}
	return &rc, nil
}

// GetRepo retrieves a Repo reference
// for a repo of the specified name,
// or returns an error
// if the repo is not in this collection
//
// **Parameters:**
//
// repoName: the repoistory name
//
// **Returns:**
//
// Repo: the located repo
// error: an error if there is a problem
func (rc *repoCollection) GetRepo(repoName string) (Repo, error) {
	r, ok := rc.reposByName[repoName]
	if !ok {
		return nil, fmt.Errorf("repository not found: %v", repoName)
	}
	return r, nil
}

// ResolveTTPRef turns a provided TTP reference into
// a Repo and absolute TTP file path
//
// **Parameters:**
//
// ttpRef: one of two things:
//
// 1. a reference of the form repo//path/to/ttp
// 2. an absolute or relative file path
//
// **Returns:**
//
// Repo: the located repo
// string: the absolute path to the specified TTP
// error: an error if there is a problem
func (rc *repoCollection) ResolveTTPRef(ttpRef string) (Repo, string, error) {

	tokens := strings.Split(ttpRef, RepoPrefixSep)
	sepCount := len(tokens) - 1
	if sepCount > 1 {
		return nil, "", fmt.Errorf("too many occurrences of '%v'", RepoPrefixSep)
	}
	if sepCount == 0 {
		// if an existing absolute or relative TTP path was specified
		// (meaning repository search is not necessary) then we just use that
		// This would happen for example if you ran `ttpforge run ~/src/myrepo/ttp.yaml`
		absPath, err := rc.resolveAbsPath(ttpRef)
		if err != nil {
			return nil, "", err
		}

		// we still need that TTP to be part of a valid repo
		// otherwise stuff like subttps won't work
		repo, err := rc.findParentRepo(absPath)
		if err != nil {
			return nil, "", err
		}
		return repo, absPath, nil
	}

	// sepCount == 1
	repoName, scopedRef := tokens[0], tokens[1]
	if _, found := rc.reposByName[repoName]; !found {
		return nil, "", fmt.Errorf("repository '%v' not found - add it with 'ttpforge install'?", repoName)
	}
	r := rc.reposByName[repoName]
	absPath, err := r.FindTTP(scopedRef)
	if err != nil {
		return nil, "", err
	}
	return r, absPath, nil
}

// ListTTPs lists all TTPs in the RepoCollection
//
// **Returns:**
//
// []string: the list of TTPs
// error: an error if there is a problem
func (rc *repoCollection) ListTTPs() ([]string, error) {
	var refsForAllTTPs []string
	for _, repo := range rc.repos {
		ttpRefs, err := repo.ListTTPs()
		if err != nil {
			return nil, err
		}
		refsForAllTTPs = append(refsForAllTTPs, ttpRefs...)
	}
	return refsForAllTTPs, nil
}

func (rc *repoCollection) findParentRepo(absPath string) (Repo, error) {
	var repoName, repoPath string
	remainingPath := absPath
	sep := string(os.PathSeparator)
	lastSepIdx := strings.LastIndex(remainingPath, sep)
	// walk up the directory tree to look for our repo
	for lastSepIdx > 0 {
		dirToSearch := remainingPath[:lastSepIdx]
		repoConfigPath := filepath.Join(dirToSearch, RepoConfigFileName)
		exists, err := afero.Exists(rc.fsys, repoConfigPath)
		if err != nil {
			return nil, err
		}
		if exists {
			repoPath = dirToSearch
			_, repoName = filepath.Split(dirToSearch)
			break
		}
		remainingPath := dirToSearch
		lastSepIdx = strings.LastIndex(remainingPath, sep)
	}
	if repoName == "" {
		return nil, fmt.Errorf("no parent repository found for path %v - you must create a %v file in the repo root", absPath, RepoConfigFileName)
	}

	// actually load the repo
	spec := Spec{
		Name: repoName,
		Path: repoPath,
	}
	r, err := spec.Load(rc.fsys, "")
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (rc *repoCollection) resolveAbsPath(path string) (string, error) {
	exists, err := afero.Exists(rc.fsys, path)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("path '%v' does not exist", path)
	}

	// afero in-memory filesystems that we use for tests
	// don't have a concept of a working directory.
	// filepath uses the OS working directory of our actual process -
	// this creates incompability that we resolve manually like this
	switch rc.fsys.(type) {
	case *afero.OsFs:
		fullPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return fullPath, nil
	}
	return path, nil
}

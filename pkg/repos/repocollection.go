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

	"github.com/facebookincubator/ttpforge/pkg/fileutils"
	"github.com/spf13/afero"
)

// RepoCollection provides useful methods for resolving and
// navigating TTPs stored in various different repositories
type RepoCollection interface {
	AddRepo(r Repo) error
	GetRepo(repoName string) (Repo, error)
	ResolveTTPRef(ttpRef string) (Repo, string, error)
	ListTTPs() ([]string, error)
	FindParentRepo(absPath string) (Repo, error)
	ParseTTPRef(ttpRef string) (Repo, string, error)
	ConvertAbsPathToAbsRef(repo Repo, absPath string) (string, error)
}

type repoCollection struct {
	repos       []Repo
	reposByName map[string]Repo
	fsys        afero.Fs
}

func (rc *repoCollection) AddRepo(r Repo) error {
	repoName := r.GetName()
	if _, found := rc.reposByName[repoName]; !found {
		rc.reposByName[repoName] = r
	} else {
		return fmt.Errorf("duplicate repo name: %v", repoName)
	}

	rc.repos = append(rc.repos, r)

	return nil
}

// NewRepoCollection validates the provided repo specs
// and assembles them into a RepoCollection
//
// **Parameters:**
//
// fsys: base file system (used for unit testing)
// specs: a list of repo.Spec entries (usually from the config file)
// basePath: the directory path that relative repo paths are resolved against
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
		if err := rc.AddRepo(r); err != nil {
			return nil, err
		}
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

// ConvertAbsPathToAbsRef converts an absolute file path to a TTP reference
// by finding the relative path from the repository's search paths
//
// **Parameters:**
//
// repo: the repository that contains the file
// absPath: the absolute path to the TTP file
//
// **Returns:**
//
// string: the TTP reference in the format "repo_name//path/to/ttp"
// error: an error if the path is not under any valid search path
func (rc *repoCollection) ConvertAbsPathToAbsRef(repo Repo, absPath string) (string, error) {
	searchPaths := repo.GetTTPSearchPaths()

	for _, searchPath := range searchPaths {
		absSearchPath, err := rc.resolveAbsPath(searchPath)
		if err != nil {
			return "", err
		}

		// absPath could be a relative path repo/b/even/more/ttps during testing
		rel, err := filepath.Rel(absSearchPath, absPath)
		if err == nil {
			return (repo.GetName() + RepoPrefixSep + rel), nil
		}
	}

	return "", fmt.Errorf("filepath %s is not under a valid search path", absPath)
}

// ParseTTPRef parses a TTP reference and returns the repository and normalized reference
// Supports both repository references (repo//path) and absolute/relative file paths
//
// **Parameters:**
//
// ttpRef: the TTP reference to parse, which can be:
//   - "repo//path/to/ttp" (repository reference)
//   - "/absolute/path/to/ttp" (absolute path)
//   - "relative/path/to/ttp" (relative path)
//
// **Returns:**
//
// Repo: the repository containing the TTP, nil if no repository is found (found by name or by searching parent directories)
// string: an **unverified** TTP reference
// error: an error if the reference is invalid or repository is not found
func (rc *repoCollection) ParseTTPRef(ttpRef string) (Repo, string, error) {
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
		repo, err := rc.FindParentRepo(absPath)
		if err != nil {
			return nil, "", err
		}

		ref, err := rc.ConvertAbsPathToAbsRef(repo, absPath)
		if err != nil {
			return nil, "", err
		}

		return repo, ref, nil
	}

	// sepCount == 1
	repoName := tokens[0]

	if repoName == "" {
		return nil, ttpRef, nil
	}

	repo, found := rc.reposByName[repoName]

	if !found {
		return nil, "", fmt.Errorf("repository '%v' not found - add it with 'ttpforge install'?", repoName)
	}

	return repo, ttpRef, nil
}

// ResolveTTPRef turns a provided TTP reference into
// a Repo and a verified absolute TTP file path
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
	repo, ref, err := rc.ParseTTPRef(ttpRef)

	if err != nil {
		return nil, "", err
	}

	if repo == nil {
		return nil, "", fmt.Errorf("no repository found for TTP reference '%v'", ttpRef)
	}

	_, scopedRef, _ := strings.Cut(ref, RepoPrefixSep)

	absPath, err := repo.FindTTP(scopedRef)
	if err != nil {
		return nil, "", err
	}

	absPath, err = rc.resolveAbsPath(absPath)
	if err != nil {
		return nil, "", err
	}

	return repo, absPath, nil
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

func isSubpath(parent string, child string) bool {
	// Clean both paths to handle . and .. elements
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	// Use filepath.Rel to determine relationship
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", child is not under parent
	// If rel is ".", they are the same path (not a subpath)
	return rel != "." && !strings.HasPrefix(rel, "..")
}

// FindParentRepo searches for a repository that contains the specified path
// First checks loaded repositories, then walks up the directory tree looking for repo config files
//
// **Parameters:**
//
// absPath: the absolute path to search from
//
// **Returns:**
//
// Repo: the repository that contains the given path
// error: an error if no parent repository is found or if there are loading issues
func (rc *repoCollection) FindParentRepo(absPath string) (Repo, error) {
	// first check if the path is already in a loaded repo
	for _, r := range rc.reposByName {
		repoPath, err := fileutils.AbsPath(r.GetFullPath())
		if err != nil {
			return nil, err
		}

		// no need to check if the config file exists because it was loaded successfully
		if isSubpath(repoPath, absPath) {
			return r, nil
		}
	}

	// if not, walk up the directory tree until we find a repo
	var repoName, repoPath string
	remainingPath := absPath
	sep := string(os.PathSeparator)
	lastSepIdx := strings.LastIndex(remainingPath, sep)

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
		remainingPath = dirToSearch
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

	// afero in-memory filesystems that we use for tests
	// don't have a concept of a working directory.
	// filepath uses the OS working directory of our actual process -
	// this creates incompability that we resolve manually like this
	switch rc.fsys.(type) {
	case *afero.OsFs:
		fullPath, err := fileutils.AbsPath(path)
		if err != nil {
			return "", err
		}
		return fullPath, nil
	default:
		return path, nil
	}
}

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

package repos_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchType int

const (
	stTTP searchType = iota
	stTemplate
)

func makeRepoTestFs() fs.StatFS {
	return fstest.MapFS{
		"repos/a/" + repos.RepoConfigFileName: &fstest.MapFile{
			Data: []byte(`ttp_search_paths: ["ttps", "more/ttps"]`),
		},
		"repos/a/ttps/foo/bar/baz/wut.yaml": &fstest.MapFile{
			Data: []byte("placeholder"),
		},
		"repos/a/more/ttps/absolute/victory.yaml": &fstest.MapFile{
			Data: []byte("placeholder"),
		},
	}
}

func TestFindTTP(t *testing.T) {

	tests := []struct {
		name                 string
		spec                 repos.Spec
		fsys                 fs.StatFS
		expectLoadError      bool
		searchType           searchType
		searchQuery          string
		expectSearchError    bool
		expectedSearchResult string
	}{
		{
			name: "Valid Repo (TTP Found in First Dir)",
			spec: repos.Spec{
				Name: "default",
				Path: "repos/a",
			},
			fsys:                 makeRepoTestFs(),
			searchType:           stTTP,
			searchQuery:          "foo/bar/baz/wut.yaml",
			expectedSearchResult: "repos/a/ttps/foo/bar/baz/wut.yaml",
		},
		{
			name: "Valid Repo (TTP Not Found)",
			spec: repos.Spec{
				Name: "default",
				Path: "repos/a",
			},
			fsys:                 makeRepoTestFs(),
			searchType:           stTTP,
			searchQuery:          "not gonna find this",
			expectedSearchResult: "",
		},
		{
			name: "Valid Repo (TTP Found in Second Dir)",
			spec: repos.Spec{
				Name: "default",
				Path: "repos/a",
			},
			fsys:                 makeRepoTestFs(),
			searchType:           stTTP,
			searchQuery:          "absolute/victory.yaml",
			expectedSearchResult: "repos/a/more/ttps/absolute/victory.yaml",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := tc.spec.Load(tc.fsys)
			if tc.expectLoadError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			var result string
			switch tc.searchType {
			case stTTP:
				result, err = repo.FindTTP(tc.searchQuery)
			}
			if tc.expectSearchError {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedSearchResult, result)
		})
	}
}

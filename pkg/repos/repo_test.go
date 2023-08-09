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
	"path/filepath"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchType int

const (
	stTTP searchType = iota
	stTemplate
)

func makeAferoTestFs(filesMap map[string][]byte) (afero.Fs, error) {
	fsys := afero.NewMemMapFs()
	for path, contents := range filesMap {
		dirPath := filepath.Dir(path)
		err := fsys.MkdirAll(dirPath, 0700)
		if err != nil {
			return nil, err
		}
		err = afero.WriteFile(fsys, path, contents, 0644)
		if err != nil {
			return nil, err
		}
	}
	return fsys, nil
}

func makeRepoTestFs(t *testing.T) afero.Fs {
	fsys, err := makeAferoTestFs(map[string][]byte{
		"repos/a/" + repos.RepoConfigFileName:                                                    []byte(`ttp_search_paths: ["ttps", "more/ttps"]`),
		"repos/a/ttps/foo/bar/baz/wut.yaml":                                                      []byte("placeholder"),
		"repos/a/more/ttps/absolute/victory.yaml":                                                []byte("placeholder"),
		"repos/invalid/" + repos.RepoConfigFileName:                                              []byte("this: is: invalid: yaml"),
		"repos/template-only/" + repos.RepoConfigFileName:                                        []byte(`template_search_paths: ["some_templates", "more/templates"]`),
		"repos/template-only/some_templates/my_template/ttp.yaml.tpl" + repos.RepoConfigFileName: []byte("placeholder"),
	},
	)
	require.NoError(t, err)
	return fsys
}

func TestFindTTP(t *testing.T) {

	tests := []struct {
		name                 string
		spec                 repos.Spec
		fsys                 afero.Fs
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
			fsys:                 makeRepoTestFs(t),
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
			fsys:                 makeRepoTestFs(t),
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
			fsys:                 makeRepoTestFs(t),
			searchType:           stTTP,
			searchQuery:          "absolute/victory.yaml",
			expectedSearchResult: "repos/a/more/ttps/absolute/victory.yaml",
		},
		{
			name: "Invalid Repo (Corrupt Config)",
			spec: repos.Spec{
				Name: "bad",
				Path: "repos/invalid",
			},
			fsys:            makeRepoTestFs(t),
			expectLoadError: true,
		},
		{
			name: "Valid Repo (Template Found in First Dir)",
			spec: repos.Spec{
				Name: "templates",
				Path: "repos/template-only",
			},
			fsys:                 makeRepoTestFs(t),
			searchType:           stTemplate,
			searchQuery:          "my_template",
			expectedSearchResult: "repos/template-only/some_templates/my_template",
		},
		{
			name: "Valid Repo (Template Not Found)",
			spec: repos.Spec{
				Name: "templates",
				Path: "repos/template-only",
			},
			fsys:                 makeRepoTestFs(t),
			searchType:           stTemplate,
			searchQuery:          "foo",
			expectedSearchResult: "",
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
			case stTemplate:
				result, err = repo.FindTemplate(tc.searchQuery)
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

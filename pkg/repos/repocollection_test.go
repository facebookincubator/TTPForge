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
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRepoCollectionTestFs(t *testing.T) afero.Fs {
	fsys, err := testutils.MakeAferoTestFs(map[string][]byte{
		"repos/a/" + RepoConfigFileName:             []byte(`ttp_search_paths: ["ttps", "more/ttps"]`),
		"repos/a/ttps/foo/bar/baz/wut.yaml":         []byte("placeholder"),
		"repos/a/more/ttps/absolute/victory.yaml":   []byte("placeholder"),
		"repos/b/" + RepoConfigFileName:             []byte(`ttp_search_paths: ["even/more/ttps"]`),
		"repos/b/even/more/ttps/attempt/again.yaml": []byte(`ttp_search_paths: ["even/more/ttps"]`),
		"not-a-repo/my-ttp.yaml":                    []byte("placeholder"),
	},
	)
	require.NoError(t, err)
	return fsys
}

func TestResolveTTPRef(t *testing.T) {

	tests := []struct {
		name               string
		description        string
		specs              []Spec
		fsys               afero.Fs
		expectLoadError    bool
		ttpRef             string
		expectResolveError bool
		expectedRepoName   string
		expectedPath       string
	}{
		{
			name: "Invalid Specs - Empty Name",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Path: "repos/b",
				},
			},
			fsys:            makeRepoCollectionTestFs(t),
			expectLoadError: true,
		},
		{
			name: "Invalid Specs - Empty Path",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "additional",
				},
			},
			fsys:            makeRepoCollectionTestFs(t),
			expectLoadError: true,
		},
		{
			name: "Invalid Specs - Same Name",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "default",
					Path: "repos/b",
				},
			},
			fsys:            makeRepoCollectionTestFs(t),
			expectLoadError: true,
		},
		{
			name: "Invalid Ref - Multiple Separators",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
			},
			fsys:               makeRepoCollectionTestFs(t),
			ttpRef:             "default//foo/bar//baz/wut.yaml",
			expectResolveError: true,
		},
		{
			name: "Invalid Ref - Non-Existent Repo",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
			},
			fsys:               makeRepoCollectionTestFs(t),
			ttpRef:             "notreal//foo/bar/baz/wut.yaml",
			expectResolveError: true,
		},
		{
			name: "Invalid Ref - Non-Existent TTP",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
			},
			fsys:               makeRepoCollectionTestFs(t),
			ttpRef:             "default//foo/bar/baz/notreal.yaml",
			expectResolveError: true,
		},
		{
			name: "Valid TTP Ref (First Repo)",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
			},
			fsys:             makeRepoCollectionTestFs(t),
			ttpRef:           "default//foo/bar/baz/wut.yaml",
			expectedRepoName: "default",
			expectedPath:     "repos/a/ttps/foo/bar/baz/wut.yaml",
		},
		{
			name: "Valid TTP Ref (Second Repo)",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "additional",
					Path: "repos/b",
				},
			},
			fsys:             makeRepoCollectionTestFs(t),
			ttpRef:           "additional//attempt/again.yaml",
			expectedRepoName: "additional",
			expectedPath:     "repos/b/even/more/ttps/attempt/again.yaml",
		},
		{
			name: "Valid TTP Path (Not Ref)",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "additional",
					Path: "repos/b",
				},
			},
			fsys:             makeRepoCollectionTestFs(t),
			ttpRef:           "repos/b/even/more/ttps/attempt/again.yaml",
			expectedRepoName: "b",
			expectedPath:     "repos/b/even/more/ttps/attempt/again.yaml",
		},
		{
			name:             "Valid TTP Path (No Repos in Collection)",
			description:      "This case verifies that regular paths still resolve even if the collection is empty",
			fsys:             makeRepoCollectionTestFs(t),
			ttpRef:           "repos/b/even/more/ttps/attempt/again.yaml",
			expectedRepoName: "b",
			expectedPath:     "repos/b/even/more/ttps/attempt/again.yaml",
		},
		{
			name: "Invalid TTP path (parent directory does not contain config)",
			specs: []Spec{
				{
					Name: "default",
					Path: "repos/a",
				},
				{
					Name: "additional",
					Path: "repos/b",
				},
			},
			fsys:               makeRepoCollectionTestFs(t),
			ttpRef:             "not-a-repo/my-ttp.yaml",
			expectResolveError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rc, err := NewRepoCollection(tc.fsys, tc.specs, "")
			if tc.expectLoadError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			r, absPath, err := rc.ResolveTTPRef(tc.ttpRef)
			if tc.expectResolveError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expectedRepoName, r.GetName())
			assert.Equal(t, tc.expectedPath, absPath)
		})
	}
}

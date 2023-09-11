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

package blocks_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalFetchURI(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple fetch",
			content: `
name: test
description: this is a test
steps:
  - name: fetch file
    fetch_uri: http://someuri.com
    location: ./location
`,
			wantError: false,
		},
		{
			name: "Overwrite fetch",
			content: `
name: test
description: this is a test
steps:
  - name: fetch file
    fetch_uri: http://someuri.com
    location: ./location
    overwrite: true
`,
			wantError: false,
		},
		{
			name: "proxy fetch",
			content: `
name: test
description: this is a test
steps:
  - name: fetch file
    fetch_uri: http://someuri.com
    location: ./location
    proxy: http://localhost:8080
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFetchURI(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "simple fetch",
			content: `
name: test
description: this is a test
steps:
  - name: simple_fetch
    fetch_uri: http://someuri.com
    location: ./location
`,
			wantError: false,
		},
		{
			name: "overwrite fetch",
			content: `
name: test
description: this is a test
steps:
  - name: overwrite_fetch
    fetch_uri: http://someuri.com
    location: ./location
    overwrite: true
`,
			wantError: false,
		},
		{
			name: "proxy fetch",
			content: `
name: test
description: this is a test
steps:
  - name: proxy_fetch
    fetch_uri: http://someuri.com
    location: ./location
    proxy: http://localhost:8080
`,
			wantError: false,
		},
		{
			name: "bad proxy",
			content: `
name: test
description: this is a test
steps:
  - name: bad_proxy
    fetch_uri: http://someuri.com
    location: ./location
    proxy: bad param
`,
			wantError: true,
		},
		{
			name: "proxy without scheme",
			content: `
name: test
description: this is a test
steps:
  - name: bad_proxy
    fetch_uri: http://someuri.com
    location: ./location
    proxy: localhost:8888
`,
			wantError: true,
		},
		{
			name: "non http/https proxy scheme",
			content: `
name: test
description: this is a test
steps:
  - name: bad_proxy
    fetch_uri: http://someuri.com
    location: ./location
    proxy: ssh://localhost:8888
`,
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps blocks.TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			assert.NoError(t, err)

			execCtx := blocks.TTPExecutionContext{
				Cfg: blocks.TTPExecutionConfig{
					Args: map[string]any{},
				},
			}

			err = ttps.ValidateSteps(execCtx)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetchURIStepOutput(t *testing.T) {

	// prepare step
	content := `
name: test_fetch_file
fetch_uri: OVERWRITTEN
location: /tmp/test.html
overwrite: true
`
	var s blocks.FetchURIStep
	execCtx := blocks.TTPExecutionContext{
		Cfg: blocks.TTPExecutionConfig{
			Args: map[string]any{},
		},
	}
	err := yaml.Unmarshal([]byte(content), &s)
	require.NoError(t, err)

	tmpfile, err := os.CreateTemp("", "fetchtest")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up

	s.Location = tmpfile.Name()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	s.FetchURI = ts.URL

	err = s.Validate(execCtx)
	require.NoError(t, err)

	// execute and check result
	_, err = s.Execute(execCtx)
	require.NoError(t, err)

	f, err := os.Stat(s.Location)
	require.NoError(t, err)
	assert.Greaterf(t, f.Size(), int64(0), "file exists but contains no data")

	dat, err := os.ReadFile(s.Location)
	require.NoError(t, err)

	assert.Equal(t, string(dat), "Hello, client\n")

}

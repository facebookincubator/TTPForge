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
	"fmt"
	"github.com/spf13/afero"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
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
			var ttps TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetchURI(t *testing.T) {
	testCases := []struct {
		name                string
		content             string
		expectValidateError bool
		expectTemplateError bool
		expectExecuteError  bool
		fsysContents        map[string][]byte
		stepVars            map[string]string
		noProxy             bool
	}{
		{
			name: "simple fetch",
			content: `
name: simple_fetch
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
`,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "overwrite fetch",
			content: `
name: overwrite_fetch
fetch_uri: http://someuri.com
location: /tmp/test.txt
overwrite: true
`,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "proxy fetch",
			content: `
name: proxy_fetch
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
proxy: http://localhost:8080
`,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "bad proxy",
			content: `
name: bad_proxy
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
proxy: bad param
`,
			expectValidateError: true,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "proxy without scheme",
			content: `
name: bad_proxy
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
proxy: localhost:8888
`,
			expectValidateError: true,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "non http/https proxy scheme",
			content: `
name: bad_proxy
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
proxy: ssh://localhost:8888
`,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "proxy ignored via NoProxy flag",
			content: `
name: noproxy_fetch
fetch_uri: http://someuri.com
location: /tmp/new_output.txt
proxy: http://nonexistent-proxy.invalid:9999
`,
			noProxy: true,
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "templates fields",
			content: `
name: proxy_fetch
fetch_uri: http://{[{.StepVars.site}]}.com
location: /tmp/{[{.StepVars.filename}]}.txt
proxy: http://{[{.StepVars.proxy}]}:8080
retries: "{[{.StepVars.retries}]}"
`,
			stepVars: map[string]string{
				"site":     "someuri",
				"filename": "output",
				"proxy":    "localhost",
				"retries":  "3",
			},
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
		},
		{
			name: "errors on missing fields",
			content: `
name: proxy_fetch
fetch_uri: http://{[{.StepVars.site}]}.com
location: ./{[{.StepVars.location}]}
proxy: http://{[{.SiteVars.proxy}]}:8080
retries: "{[{.StepVars.retries}]}"
`,
			expectTemplateError: true,
		},
		{
			name: "fails validation after templating proxy",
			content: `
name: fetch
fetch_uri: http://someuri.com
proxy: ://{[{.StepVars.proxy}]}:8080
location: /tmp/output.txt
`,
			stepVars: map[string]string{
				"proxy": "someuri",
			},
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
			expectTemplateError: true,
		},
		{
			name: "fails validation after templating location",
			content: `
name: fetch
fetch_uri: http://someuri.com
location: /tmp/{[{.StepVars.location}]}.txt
`,
			stepVars: map[string]string{
				"location": "test",
			},
			fsysContents: map[string][]byte{
				"/tmp/test.txt": []byte("Test file"),
			},
			expectTemplateError: true,
		},
	}

	// prepare test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Here's some data!"))
	}))
	defer testServer.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// parse step
			var step FetchURIStep
			err := yaml.Unmarshal([]byte(tc.content), &step)
			assert.NoError(t, err)

			// prepare execution context
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

			// prepare filesystem
			if tc.fsysContents != nil {
				fsys, err := testutils.MakeAferoTestFs(tc.fsysContents)
				require.NoError(t, err)
				step.FileSystem = fsys
			} else {
				step.FileSystem = afero.NewMemMapFs()
			}

			// validate
			err = step.Validate(execCtx)
			if tc.expectValidateError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// template
			err = step.Template(execCtx)
			if tc.expectTemplateError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Point all requests to the test server
			step.FetchURI = testServer.URL

			// Set NoProxy flag if specified, otherwise clear proxy since httptest doesn't support it
			if tc.noProxy {
				execCtx.Cfg.NoProxy = true
			} else {
				step.Proxy = ""
			}

			// execute
			_, err = step.Execute(execCtx)
			if tc.expectExecuteError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
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
	var s FetchURIStep
	execCtx := NewTTPExecutionContext()
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

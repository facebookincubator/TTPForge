/*
Copyright © 2025-present, Meta Platforms, Inc. and affiliates
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/yaml.v3"
)

func TestUnmarshalHTTPRequest(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple http request",
			content: `
name: test basic
description: this is a test basic test
steps:
  - name: get url
    http_request: http://someuri.com
`,
			wantError: false,
		},
		{
			name: "Request with specified type GET",
			content: `
name: test get
description: This is a test with specified type
steps:
  - name: specific request type
    http_request: http://someuri.com
    type: GET
`,
			wantError: false,
		},
		{
			name: "Request with proxy",
			content: `
name: test proxy
description: This is a test with proxy
steps:
  - name: request through proxy
    http_request: http://someuri.com
    proxy: http://localhost:8080
`,
			wantError: false,
		},
		{
			name: "Request with headers",
			content: `
name: test headers
description: this is a test setting headers
steps:
  - name: request with headers
    http_request: http://someuri.com
    headers:
      - field: "User-Agent"
        value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
      - field: "Content-Type"
        value: "application/x-www-form-urlencoded; charset=UTF-8"
`,
			wantError: false,
		},
		{
			name: "Request with parameters",
			content: `
name: test parameters
description: this is a test with http parameters
steps:
  - name: get with parameters
    http_request: http://someuri.com
    parameters:
      - name: "foo"
        value: "bar"
      - name: "moo"
        value: "cow"
`,
			wantError: false,
		},
		{
			name: "POST Request with headers and body",
			content: `
name: test headers and body of POST request
description: this is a test typical POST test
steps:
  - name: request with headers
    http_request: http://someuri.com
    type: POST
    headers:
      - field: "User-Agent"
        value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
      - field: "Content-Type"
        value: "application/x-www-form-urlencoded; charset=UTF-8"
    body: "{'this': 'is', 'a': 'test', 'body': 'of', 'post': 'request'}"
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

func TestValidateHTTPequest(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "Simple http request",
			content: `
name: test basic
description: this is a test basic test
steps:
  - name: get url
    http_request: http://someuri.com
`,
			wantError: false,
		},
		{
			name: "Request with specified type GET",
			content: `
name: test get
description: This is a test with specified type
steps:
  - name: specific request type
    http_request: http://someuri.com
    type: GET
`,
			wantError: false,
		},
		{
			name: "Request with proxy",
			content: `
name: test proxy
description: This is a test with proxy
steps:
  - name: request through proxy
    http_request: http://someuri.com
    proxy: http://localhost:8080
`,
			wantError: false,
		},
		{
			name: "Request with headers",
			content: `
name: test headers
description: this is a test
steps:
  - name: request with headers
    http_request: http://someuri.com
    headers:
      - field: "User-Agent"
        value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
      - field: "Content-Type"
        value: "application/x-www-form-urlencoded; charset=UTF-8"
`,
			wantError: false,
		},
		{
			name: "Request with parameters",
			content: `
name: test parameters
description: this is a test with http parameters
steps:
  - name: fetch file
    http_request: http://someuri.com
    parameters:
      - name: "foo"
        value: "bar"
      - name: "moo"
        value: "cow"
`,
			wantError: false,
		},
		{
			name: "POST Request with headers and body",
			content: `
name: test headers and body of POST request
description: this is a test typical POST test
steps:
  - name: request with headers
    http_request: http://someuri.com
    type: POST
    headers:
      - field: "User-Agent"
        value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
      - field: "Content-Type"
        value: "application/x-www-form-urlencoded; charset=UTF-8"
    body: "{'this': 'is', 'a': 'test', 'body': 'of', 'post': 'request'}"
`,
			wantError: false,
		},
		{
			name: "Invalid URL http request",
			content: `
name: test invalid url
description: this is a test with an invalid url
steps:
  - name: get screwy url
    http_request: somebrokenurl
`,
			wantError: true,
		},
		{
			name: "Invalid Proxy URL request",
			content: `
name: test invalid proxy
description: this is a test with an invalid proxy url
steps:
    - name: get url through bad proxy
      http_request: https://someuri.com
      proxy: thisnotaurlok
`,
			wantError: true,
		},
		{
			name: "Malformed headers",
			content: `
name: malformed headers
description: this is a test with headers missing value
steps:
  - name: get request with malformed headers
    http_request: https://someuri.com
    headers:
      - field: "User-Agent"
        value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
      - field: "Content-Type"
`,
			wantError: true,
		},
		{
			name: "Malformed parameters",
			content: `
name: malformed parameters
description: this is a test with parameter missing value
steps:
  - name: get url through bad proxy
    http_request: https://someuri.com
    parameters:
      - name: "search"
        value: "everything"
      - name: "something_else"
`,
			wantError: true,
		},
		{
			name: "Bad Regex",
			content: `
name: malformed regular expression
description: this is a test with a weird regex pattern
steps:
  - name: get url through bad proxy
    http_request: https://someuri.com
    parameters:
      - name: "search"
        value: "everything"
    regex: "(dfaefawefaew"
`,
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ttps TTP
			err := yaml.Unmarshal([]byte(tc.content), &ttps)
			assert.NoError(t, err)

			execCtx := NewTTPExecutionContext()
			err = ttps.Validate(execCtx)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPequest(t *testing.T) {
	testCases := []struct {
		name                string
		content             string
		expectValidateError bool
		expectTemplateError bool
		expectExecuteError  bool
		stepVars            map[string]string
		overwriteProxy      bool
		expectedOutput      string
	}{
		{
			name: "Simple http request",
			content: `
name: get url
http_request: http://someuri.com
`,
		},
		{
			name: "Request with specified type GET",
			content: `
name: specific request type
http_request: http://someuri.com
type: GET
`,
		},
		{
			name: "Request with proxy",
			content: `
name: request through proxy
http_request: http://someuri.com
proxy: http://localhost:8080
`,
			overwriteProxy: true,
		},
		{
			name: "Request with headers",
			content: `
name: request with headers
http_request: http://someuri.com
headers:
- field: "User-Agent"
  value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
- field: "Content-Type"
  value: "application/x-www-form-urlencoded; charset=UTF-8"
`,
		},
		{
			name: "Request with parameters",
			content: `
name: fetch file
http_request: http://someuri.com
parameters:
- name: "foo"
  value: "bar"
- name: "moo"
  value: "cow"
`,
		},
		{
			name: "POST Request with headers and body",
			content: `
name: request with headers
http_request: http://someuri.com
type: POST
headers:
- field: "User-Agent"
  value: "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail appname/appversion Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version appname/appversion;"
- field: "Content-Type"
  value: "application/x-www-form-urlencoded; charset=UTF-8"
body: "{'this': 'is', 'a': 'test', 'body': 'of', 'post': 'request'}"
`,
		},
		{
			name: "Invalid URL http request",
			content: `
name: get screwy url
http_request: somebrokenurl
`,
			expectValidateError: true,
		},
		{
			name: "Invalid Proxy URL request",
			content: `
name: get url through bad proxy
http_request: https://someuri.com
proxy: thisnotaurlok
`,
			expectValidateError: true,
		},
		{
			name: "Bad Regex",
			content: `
name: get url through bad proxy
http_request: https://someuri.com
parameters:
- name: "search"
  value: "everything"
regex: "(dfaefawefaew"
`,
			expectValidateError: true,
		},
		{
			name: "Successful Template",
			content: `
name: template test
http_request: http://someuri.com/{[{.StepVars.path}]}
proxy: http://{[{.StepVars.proxy}]}:8080
type: "{[{.StepVars.type}]}"
headers:
- field: "{[{.StepVars.name}]}"
  value: "{[{.StepVars.value}]}"
parameters:
- name: "{[{.StepVars.name}]}"
  value: "{[{.StepVars.value}]}"
body: "{'body': '{[{.StepVars.body}]}'}"
regex: "{[{.StepVars.regex}]}"
`,
			stepVars: map[string]string{
				"path":  "some/path",
				"proxy": "localhost",
				"type":  "POST",
				"name":  "key",
				"value": "value",
				"body":  "this is some data",
				"regex": ".*",
			},
			overwriteProxy: true,
		},
		{
			name: "Error on Template",
			content: `
name: template test
http_request: http://someuri.com/{[{.StepVars.path}]}
proxy: http://{[{.StepVars.proxy}]}:8080
type: "{[{.StepVars.type}]}"
headers:
- field: "{[{.StepVars.name}]}"
  value: "{[{.StepVars.value}]}"
parameters:
- name: "{[{.StepVars.name}]}"
  value: "{[{.StepVars.value}]}"
body: "{'body': '{[{.StepVars.body}]}'}"
regex: "{[{.StepVars.regex}]}"
`,
			expectTemplateError: true,
		},
		{
			name: "Fails validation after templating uri",
			content: `
name: template test
http_request: ://someuri.com/{[{.StepVars.path}]}
`,
			stepVars: map[string]string{
				"path": "some/path",
			},
			expectTemplateError: true,
		},
		{
			name: "Fails validation after templating proxy",
			content: `
name: template test
http_request: http://someuri.com/
proxy: ://{[{.StepVars.proxy}]}:8080
`,
			stepVars: map[string]string{
				"proxy": "localhost",
			},
			expectTemplateError: true,
		},
		{
			name: "Fails validation after templating type",
			content: `
name: template test
http_request: http://someuri.com/
type: "{[{.StepVars.type}]}"
`,
			stepVars: map[string]string{
				"type": "WTF",
			},
			expectTemplateError: true,
		},
		{
			name: "Fails validation after templating regex",
			content: `
name: template test
http_request: http://someuri.com/
regex: "{[{.StepVars.regex}]}"
`,
			stepVars: map[string]string{
				"regex": "[",
			},
			expectTemplateError: true,
		},
		{
			name: "Outputs to var properly",
			content: `
name: template test
http_request: http://someuri.com/
outputvar: testvar
`,
			stepVars:       map[string]string{},
			expectedOutput: "Here's some data!",
		},
		{
			name: "Overwrites output to var properly",
			content: `
name: template test
http_request: http://someuri.com/
outputvar: testvar
`,
			stepVars: map[string]string{
				"testvar": "some other data",
			},
			expectedOutput: "Here's some data!",
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
			var step HTTPRequestStep
			err := yaml.Unmarshal([]byte(tc.content), &step)
			assert.NoError(t, err)

			// prep execution context
			execCtx := NewTTPExecutionContext()
			execCtx.Vars.StepVars = tc.stepVars

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

			// overwrite proxy if needed, since its hard to actually test with the test server
			if tc.overwriteProxy {
				step.Proxy = ""
			}
			// Overwrite url with test server url for execution
			step.HTTPRequest = testServer.URL

			// execute
			_, err = step.Execute(execCtx)
			if tc.expectExecuteError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// check output
			if tc.expectedOutput != "" {
				assert.Equal(t, tc.expectedOutput, execCtx.Vars.StepVars[step.OutputVar])
			}
			assert.NoError(t, err)
		})
	}
}

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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"go.uber.org/zap"
)

// HTTPHeader represents a key-value pair for HTTP header.
type HTTPHeader struct {
	Field string `yaml:"field,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// HTTPParameter represents a single HTTP parameter.
type HTTPParameter struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// HTTPRequestStep represents a step in a process that consists of a main action,
// a cleanup action, and additional metadata.
type HTTPRequestStep struct {
	actionDefaults `yaml:",inline"`
	HTTPRequest    string           `yaml:"http_request,omitempty"`
	Type           string           `yaml:"type,omitempty"`
	Headers        []*HTTPHeader    `yaml:"headers,omitempty"`
	Parameters     []*HTTPParameter `yaml:"parameters,omitempty"`
	Body           string           `yaml:"body,omitempty"`
	Regex          string           `yaml:"regex,omitempty"`
	Proxy          string           `yaml:"proxy,omitempty"`
	Response       string           `yaml:"response,omitempty"`
}

// NewHTTPRequestStep creates a new HTTPRequestStep instance and returns a pointer to it.
func NewHTTPRequestStep() *HTTPRequestStep {
	return &HTTPRequestStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (r *HTTPRequestStep) IsNil() bool {
	return r.HTTPRequest == ""
}

// Validate validates the HTTPRequestStep.
func (r *HTTPRequestStep) Validate(execCtx TTPExecutionContext) error {

	// Validate the target URL
	if r.HTTPRequest != "" {
		uri, err := url.Parse(r.HTTPRequest)
		if err != nil {
			return err
		} else if uri.Host == "" || uri.Scheme == "" {
			return fmt.Errorf("invalid URL given for request URL: %s", r.HTTPRequest)
		}
	}

	// Validate the proxy URL
	if r.Proxy != "" {
		uri, err := url.Parse(r.Proxy)
		if err != nil {
			return err
		} else if uri.Host == "" || uri.Scheme == "" {
			return fmt.Errorf("invalid URL given for Proxy: %s", r.Proxy)
		}
	}

	// Validate the http request type is valid
	if r.Type != "" {
		isHTTPMethod := false
		for _, method := range []string{"GET", "POST", "PUT", "DELETE", "HEAD", "PATCH"} {
			if strings.EqualFold(r.Type, method) {
				isHTTPMethod = true
				break
			}
		}
		if !isHTTPMethod {
			return fmt.Errorf("unsupported HTTP request type: %s", r.Type)
		}
	}

	// Validate headers
	for _, header := range r.Headers {
		if header.Field == "" || header.Value == "" {
			return fmt.Errorf("broken HTTP header %s: %s", header.Value, header.Field)
		}
	}

	// Validate parameters
	for _, parameter := range r.Parameters {
		if parameter.Name == "" || parameter.Value == "" {
			return fmt.Errorf("broken HTTP parameters %s: %s", parameter.Name, parameter.Value)
		}
	}

	// Validate regex
	if r.Regex != "" {
		regexTrim := strings.TrimSuffix(r.Regex, "\n")
		_, err := regexp.Compile(regexTrim)
		if err != nil {
			return fmt.Errorf("invalid regular expression: %v", err)
		}
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (r *HTTPRequestStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	logging.L().Info("========= Executing ==========")
	logging.L().Infof("HTTPRequest to: %s", r.HTTPRequest)
	if err := r.SendRequest(execCtx); err != nil {
		logging.L().Error(zap.Error(err))
		return nil, err
	}
	logging.L().Info("========= Complete ==========")
	return &ActResult{}, nil
}

// HTTPRequest executes the HTTPRequestStep.
func (r *HTTPRequestStep) SendRequest(execCtx TTPExecutionContext) error {

	// Gather the parameters
	params := url.Values{}
	for _, parameter := range r.Parameters {
		params.Add(parameter.Name, parameter.Value)
	}
	// Construct the full URL with parameters
	fullURL := fmt.Sprintf("%s?%s", r.HTTPRequest, params.Encode())

	// Trim the body of any trailing new lines
	trimBody := strings.TrimSuffix(r.Body, "\n")

	// Create a new request with the specified method, URL, and body.
	req, err := http.NewRequest(r.Type, fullURL, strings.NewReader(trimBody))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	// Loop through and set each header
	for _, header := range r.Headers {
		if header.Field != "" && header.Value != "" {
			req.Header.Set(header.Field, header.Value)
		}

	}

	// Send the request using the default HTTP client
	client := &http.Client{}

	// Set proxy if specified.
	if r.Proxy != "" {
		proxyURI, err := url.Parse(r.Proxy)
		if err != nil {
			return err
		}
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxyURI),
		}
		client = &http.Client{Transport: tr}
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}

	// Build final response
	finalResponse := string(body)

	// If using regex to extract, part of the response.
	if r.Regex != "" {
		// Remove pesky new line from end of r.Regex input.
		regexTrim := strings.TrimSuffix(r.Regex, "\n")
		// Compile the regular expression.
		re := regexp.MustCompile(regexTrim)

		// Find all matches in the response body
		matches := re.FindAllString(string(body), -1)
		if matches != nil {
			// If there are matches, use the first one as the final response.
			finalResponse = matches[0]
		} else {
			finalResponse = "No matches for pattern found..."
		}
	}

	// Store the response as an environment variable.
	if r.Response != "" {
		err = os.Setenv(r.Response, finalResponse)
		if err != nil {
			return fmt.Errorf("Error setting environment variable: %v", err)
		}
	}

	logging.L().Infof("Response: %s", finalResponse)

	return nil
}

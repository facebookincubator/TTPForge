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
	"encoding/json"
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

// ResponseData represents the response data from an HTTP request.
type ResponseData struct {
	Head map[string][]string `json:"head"`
	Body string              `json:"body"`
}

// HTTPRequestStep represents a step in a process that consists of a main action,
// a cleanup action, and additional metadata.
type HTTPRequestStep struct {
	actionDefaults   `yaml:",inline"`
	HTTPRequest      string           `yaml:"http_request,omitempty"`
	Type             string           `yaml:"type,omitempty"`
	Headers          []*HTTPHeader    `yaml:"headers,omitempty"`
	Parameters       []*HTTPParameter `yaml:"parameters,omitempty"`
	Body             string           `yaml:"body,omitempty"`
	Regex            string           `yaml:"regex,omitempty"`
	Proxy            string           `yaml:"proxy,omitempty"`
	DisableRedirects bool             `yaml:"disable_redirects,omitempty"`
	ResponseHeaders  bool             `yaml:"response_headers,omitempty"`
	Response         string           `yaml:"response,omitempty"`
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
	// Validate the target URL, skip if contains template
	if r.HTTPRequest != "" && !execCtx.containsStepTemplating(r.HTTPRequest) {
		err := r.validateURL()
		if err != nil {
			return err
		}
	}

	// Validate the proxy URL, skip if contains template
	if r.Proxy != "" && !execCtx.containsStepTemplating(r.Proxy) {
		err := r.validateProxy()
		if err != nil {
			return err
		}
	}

	// Validate the http request type is valid, skip if contains template
	if r.Type != "" && !execCtx.containsStepTemplating(r.Type) {
		err := r.validateType()
		if err != nil {
			return err
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
			return fmt.Errorf("invalid regular expression: %w", err)
		}
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (r *HTTPRequestStep) Template(execCtx TTPExecutionContext) error {
	var err error

	// Template and revalidate httprequest
	if execCtx.containsStepTemplating(r.HTTPRequest) {
		r.HTTPRequest, err = execCtx.templateStep(r.HTTPRequest)
		if err != nil {
			return err
		}
		err = r.validateURL()
		if err != nil {
			return err
		}
	}

	// Template and revalidate proxy
	if execCtx.containsStepTemplating(r.Proxy) {
		r.Proxy, err = execCtx.templateStep(r.Proxy)
		if err != nil {
			return err
		}
		err = r.validateProxy()
		if err != nil {
			return err
		}
	}

	// Template and revalidate type
	if execCtx.containsStepTemplating(r.Type) {
		r.Type, err = execCtx.templateStep(r.Type)
		if err != nil {
			return err
		}
		err = r.validateType()
		if err != nil {
			return err
		}
	}

	// Template headers
	for i := range r.Headers {
		r.Headers[i].Field, err = execCtx.templateStep(r.Headers[i].Field)
		if err != nil {
			return err
		}
		r.Headers[i].Value, err = execCtx.templateStep(r.Headers[i].Value)
		if err != nil {
			return err
		}
	}

	// Template parameters
	for i := range r.Parameters {
		r.Parameters[i].Name, err = execCtx.templateStep(r.Parameters[i].Name)
		if err != nil {
			return err
		}
		r.Parameters[i].Value, err = execCtx.templateStep(r.Parameters[i].Value)
		if err != nil {
			return err
		}
	}

	// Template body
	r.Body, err = execCtx.templateStep(r.Body)
	if err != nil {
		return err
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
	fullURL := r.HTTPRequest
	if len(params) > 0 {
		fullURL = fmt.Sprintf("%s?%s", fullURL, params.Encode())
	}

	// Trim the body of any trailing new lines
	trimBody := strings.TrimSuffix(r.Body, "\n")

	// Create a new request with the specified method, URL, and body.
	req, err := http.NewRequest(r.Type, fullURL, strings.NewReader(trimBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Loop through and set each header
	for _, header := range r.Headers {
		if header.Field != "" && header.Value != "" {
			req.Header.Set(header.Field, header.Value)
		}

	}

	// Create base transport
	tr := &http.Transport{}

	// Configure proxy if specified
	if r.Proxy != "" {
		proxyURI, err := url.Parse(r.Proxy)
		if err != nil {
			return err
		}
		tr.Proxy = http.ProxyURL(proxyURI)
	}

	// Create HTTP client with configured transport
	client := &http.Client{
		Transport: tr,
	}

	// Configure redirect behavior if specified
	if r.DisableRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Sanity Check Expectations
	if r.Type == "HEAD" && !r.ResponseHeaders {
		logging.L().Warn("You're making a HEAD request with response_headers set to false! YOU WILL NOT SEE ANYTHING!")
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Default response to just include the body
	finalResponse := string(body)

	// If using response_headers, include the full response.
	if r.ResponseHeaders {
		// Map headers
		headersMap := make(map[string][]string)
		for k, v := range resp.Header {
			headersMap[k] = v
		}

		// ResponseData struct
		data := ResponseData{
			Head: headersMap,
			Body: string(body),
		}

		fullResponse, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("error marshalling response: %w", err)
		}
		// Convert the byte slice to a string and send back as the response.
		finalResponse = string(fullResponse)
	}

	// If using regex to extract, part of the response.
	if r.Regex != "" {
		// Remove pesky new line from end of r.Regex input.
		regexTrim := strings.TrimSuffix(r.Regex, "\n")
		// Compile the regular expression.
		re := regexp.MustCompile(regexTrim)

		// Find all matches in the response body
		matches := re.FindAllString(finalResponse, -1)
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
			return fmt.Errorf("error setting environment variable: %w", err)
		}
	}
	logging.L().Infof("Response: %s", finalResponse)

	if r.OutputVar != "" {
		execCtx.Vars.StepVars[r.OutputVar] = finalResponse
	}

	return nil
}

// validateURL validates that the URL is valid URI.  Returns an error if validation fails, otherwise returns nil
func (r *HTTPRequestStep) validateURL() error {
	uri, err := url.Parse(r.HTTPRequest)
	if err != nil {
		return err
	} else if uri.Host == "" || uri.Scheme == "" {
		return fmt.Errorf("invalid URL given for request URL: %s", r.HTTPRequest)
	}

	return nil
}

// validateProxy validates that the proxy is a valid URI.  Returns an error if validation fails, otherwise returns nil
func (r *HTTPRequestStep) validateProxy() error {
	uri, err := url.Parse(r.Proxy)
	if err != nil {
		return err
	} else if uri.Host == "" || uri.Scheme == "" {
		return fmt.Errorf("invalid URL given for Proxy: %s", r.Proxy)
	}

	return nil
}

// ValidHTTPMethods contains all supported HTTP methods
var ValidHTTPMethods = []string{"GET", "POST", "PUT", "DELETE", "HEAD", "PATCH", "OPTIONS"}

// validateType validates that the request type is a valid HTTP request type.  Returns an error if validation fails, otherwise returns nil
func (r *HTTPRequestStep) validateType() error {
	isHTTPMethod := false
	for _, method := range ValidHTTPMethods {
		if strings.EqualFold(r.Type, method) {
			isHTTPMethod = true
			break
		}
	}
	if !isHTTPMethod {
		return fmt.Errorf("unsupported HTTP request type: %s", r.Type)
	}

	return nil
}

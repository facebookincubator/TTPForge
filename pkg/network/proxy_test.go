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

package network_test

import (
	"os"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/network"
)

func TestEnableHTTPSProxy(t *testing.T) {
	tests := []struct {
		name  string
		proxy string
	}{
		{
			name:  "Enable proxy",
			proxy: "http://proxy.example.com:8080",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			network.EnableHTTPSProxy(tc.proxy)

			httpsProxy, _ := os.LookupEnv("https_proxy")

			if httpsProxy != tc.proxy {
				t.Errorf("expected https_proxy to be set to %s, got https_proxy: %s", tc.proxy, httpsProxy)
			}
		})
	}
}

func TestDisableHTTPSProxy(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Disable proxy",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			network.DisableHTTPSProxy()

			httpsProxy, httpsExists := os.LookupEnv("https_proxy")

			if httpsExists {
				t.Errorf("expected https_proxy to be unset, https_proxy: %s", httpsProxy)
			}
		})
	}
}

func TestEnableNoProxy(t *testing.T) {
	testCases := []struct {
		name    string
		domains string
	}{
		{
			name:    "Single domain",
			domains: "example.com",
		},
		{
			name:    "Multiple domains",
			domains: "example.com,example.org",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			network.EnableNoProxy(tc.domains)
			result := os.Getenv("no_proxy")
			if result != tc.domains {
				t.Errorf("Expected no_proxy to be set to '%s', but got '%s'", tc.domains, result)
			}
		})
	}
}

func TestDisableNoProxy(t *testing.T) {
	os.Setenv("no_proxy", "example.com")
	network.DisableNoProxy()
	result := os.Getenv("no_proxy")
	if result != "" {
		t.Errorf("Expected no_proxy to be unset, but got '%s'", result)
	}
}

func TestEnableHTTPProxy(t *testing.T) {
	proxy := "http://proxy.example.com:8080"
	network.EnableHTTPProxy(proxy)
	result := os.Getenv("http_proxy")
	if result != proxy {
		t.Errorf("Expected http_proxy to be set to '%s', but got '%s'", proxy, result)
	}
}

func TestDisableHTTPProxy(t *testing.T) {
	os.Setenv("http_proxy", "http://proxy.example.com:8080")
	network.DisableHTTPProxy()
	result := os.Getenv("http_proxy")
	if result != "" {
		t.Errorf("Expected http_proxy to be unset, but got '%s'", result)
	}
}

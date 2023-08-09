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

package network

import "os"

// EnableHTTPSProxy sets the environment variables "https_proxy"
// and "http_proxy" to the provided proxy string.
// This enables an HTTPS proxy for network operations in the current process.
//
// **Parameters:**
//
// proxy: A string representing the URL of the proxy server to be used for HTTPS connections.
func EnableHTTPSProxy(proxy string) {
	os.Setenv("https_proxy", proxy)
	os.Setenv("http_proxy", proxy)
}

// DisableHTTPSProxy unsets the "https_proxy" and "http_proxy"
// environment variables. This disables the HTTPS proxy for network
// operations in the current process.
//
// **Parameters:**
//
// proxy: A string representing the URL of the proxy server to be used for HTTPS connections.
func DisableHTTPSProxy() {
	os.Unsetenv("https_proxy")
}

// EnableNoProxy sets the "no_proxy" environment variable to
// the provided domains string. This excludes the specified domains
// from being proxied for network operations in the current process.
//
// **Parameters:**
//
// domains: A string representing a comma-separated list of domain names
// to be excluded from proxying.
func EnableNoProxy(domains string) {
	os.Setenv("no_proxy", domains)
}

// DisableNoProxy unsets the "no_proxy" environment variable.
// This clears the list of domains excluded from being
// proxied for network operations in the current process.
func DisableNoProxy() {
	os.Unsetenv("no_proxy")
}

// EnableHTTPProxy sets the "http_proxy" environment variable
// to the provided proxy string. This enables an HTTP proxy
// for network operations in the current process.
//
// **Parameters:**
//
// proxy: A string representing the URL of the proxy server to be used for HTTP connections.
func EnableHTTPProxy(proxy string) {
	os.Setenv("http_proxy", proxy)
}

// DisableHTTPProxy unsets the "http_proxy" environment variable.
// This disables the HTTP proxy for network operations in the current process.
func DisableHTTPProxy() {
	os.Unsetenv("http_proxy")
}

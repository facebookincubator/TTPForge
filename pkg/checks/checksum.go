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

package checks

import (
	"crypto/sha256"
	"fmt"
)

// Checksum is a struct that contains different types
// of checksums against which a file can be verified.
// Supports SHA256 checksums.
type Checksum struct {
	SHA256 string `yaml:"sha256,omitempty"`
}

// Verify computes the checksum of the contents
// and compares it to the expected value.
// Supports SHA256 checksums.
func (c *Checksum) Verify(contents []byte) error {
	if c.SHA256 == "" {
		return fmt.Errorf("checksum is empty - must provide sha256")
	}

	// Verify SHA256
	rawResult := sha256.Sum256(contents)
	actualSHA256 := fmt.Sprintf("%x", rawResult)
	if actualSHA256 != c.SHA256 {
		return fmt.Errorf("sha256 checksum mismatch: expected %s, got %s",
			c.SHA256, actualSHA256)
	}

	return nil
}

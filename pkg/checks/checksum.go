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
// Right now we just support SHA256 checksums, but in the future
// others such as MD5 can be added if needed.
type Checksum struct {
	SHA256 string `yaml:"sha256"`
}

// Verify computes the checksum of the contents
// and compares it to the expected value
func (c *Checksum) Verify(contents []byte) error {
	if c.SHA256 == "" {
		return fmt.Errorf("Checksum is empty")
	}
	rawResult := sha256.Sum256(contents)
	if fmt.Sprintf("%x", rawResult) != c.SHA256 {
		return fmt.Errorf("contents do not match checksum")
	}
	return nil
}

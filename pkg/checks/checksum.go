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

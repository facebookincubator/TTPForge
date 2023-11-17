package checks

import (
	"fmt"

	"github.com/spf13/afero"
)

// PathExists is a condition that verifies that a file exists at a given path
// It can also verify the contents of the file against a checksum
type PathExists struct {
	Path     string    `yaml:"path_exists"`
	Checksum *Checksum `yaml:"checksum"`
}

// Verify checks the condition and returns an error if it fails
func (c *PathExists) Verify(ctx VerificationContext) error {
	fsys := ctx.FileSystem

	// basic existence check
	exists, err := afero.Exists(fsys, c.Path)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("file %q does not exist", c.Path)
	}

	// verify the checksum if provided
	if c.Checksum != nil {
		contentBytes, err := afero.ReadFile(fsys, c.Path)
		if err != nil {
			return err
		}
		return c.Checksum.Verify(contentBytes)
	}
	return nil
}

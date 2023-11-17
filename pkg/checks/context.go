package checks

import (
	"github.com/spf13/afero"
)

// VerificationContext contains contextual
// information required to verify conditions
// of various types
type VerificationContext struct {
	FileSystem afero.Fs
}

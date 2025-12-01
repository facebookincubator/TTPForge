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

package platforms

import (
	"errors"
	"fmt"

	"github.com/facebookincubator/ttpforge/pkg/logging"
)

// Spec defines a platform as an
// os/arch pair.
type Spec struct {
	OS   string
	Arch string
}

// GetValidOS returns all valid operating system values supported by TTPForge
func GetValidOS() []string {
	return []string{
		"android",
		"darwin",
		"dragonfly",
		"freebsd",
		"linux",
		"netbsd",
		"openbsd",
		"plan9",
		"solaris",
		"windows",
	}
}

// IsCompatibleWith returns true if the current spec is compatible with the
// spec specified as its argument.
// TTPs will often not care about the architecture and will
// therefore specify requirements like:
//
//	platforms:
//	  - os: windows
//
// In such cases, we can assume that the TTP is compatible with all
// architectures.
func (s *Spec) IsCompatibleWith(otherSpec Spec) bool {
	// would be a bit weird to have a TTP
	// that cares about architecture but not OS,
	// but no harm in it
	if s.OS != "" && s.OS != otherSpec.OS {
		return false
	}
	if s.Arch != "" {
		return s.Arch == otherSpec.Arch
	}
	return true
}

// String returns a human readable representation of the spec;
// it is mainly used for error messages.
func (s *Spec) String() string {
	anyOS := "[any OS]"
	anyArch := "[any architecture]"
	fmtStr := "%v/%v"
	if s.OS != "" {
		if s.Arch != "" {
			return fmt.Sprintf(fmtStr, s.OS, s.Arch)
		}
		return fmt.Sprintf(fmtStr, s.OS, anyArch)
	} else if s.Arch != "" {
		return fmt.Sprintf(fmtStr, anyOS, s.Arch)
	}
	if s.OS != "" && s.Arch != "" {
		return fmt.Sprintf(fmtStr, anyOS, s.Arch)
	}
	return fmt.Sprintf(fmtStr, anyOS, anyArch)
}

// Validate checks whether the platform spec is valid.
// To be valid, the spec must be enforceable, so
// at least one of the fields must be non-empty.
func (s *Spec) Validate() error {
	if s.OS == "" && s.Arch == "" {
		return fmt.Errorf("os and arch cannot both be empty")
	}

	validOSList := GetValidOS()
	validOSMap := make(map[string]bool)
	for _, os := range validOSList {
		validOSMap[os] = true
	}

	if s.OS != "" && !validOSMap[s.OS] {
		errorMsg := fmt.Sprintf("invalid `os` value %q specified", s.OS)
		logging.L().Errorf(errorMsg)
		logging.L().Errorf("valid values are:")
		for _, os := range validOSList {
			logging.L().Errorf("\t%s", os)
		}
		return errors.New(errorMsg)
	}

	// https://stackoverflow.com/a/20728862
	validArch := map[string]bool{
		"arm":      true,
		"386":      true,
		"amd64":    true,
		"arm64":    true,
		"ppc64":    true,
		"ppc64le":  true,
		"mips":     true,
		"mipsle":   true,
		"mips64":   true,
		"mips64le": true,
	}
	if s.Arch != "" && !validArch[s.Arch] {
		errorMsg := fmt.Sprintf("invalid `arch` value %q specified", s.Arch)
		logging.L().Errorf(errorMsg)
		logging.L().Errorf("valid values are:")
		for k := range validArch {
			logging.L().Errorf("\t%s", k)
		}
		return errors.New(errorMsg)
	}
	return nil
}

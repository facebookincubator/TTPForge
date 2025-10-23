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

package blocks

import (
	"fmt"
	"runtime"

	"github.com/facebookincubator/ttpforge/pkg/checks"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/platforms"
)

// RequirementsConfig specifies the prerequisites that must be
// satisfied before executing a particular TTP.
//
// **Attributes:**
//
// ExpectSuperuser: Whether the TTP assumes superuser privileges
type RequirementsConfig struct {
	ExpectSuperuser bool             `yaml:"superuser,omitempty"`
	Platforms       []platforms.Spec `yaml:"platforms,omitempty"`
}

// Validate checks that the requirements section
// is well-formed - it does not actually
// check that the requirements are met.
func (rc *RequirementsConfig) Validate() error {
	// nil is valid - it just means don't enforce any
	// requirements
	if rc == nil {
		return nil
	}
	for _, platform := range rc.Platforms {
		if err := platform.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Verify checks that the requirements specified
// in the requirements section are actually satisfied by the environment in
// which the TTP is currently running.
func (rc *RequirementsConfig) Verify(ctx checks.VerificationContext) error {
	// simplifies things a bit for callers
	if rc == nil {
		return nil
	}

	// check platform compatibility:
	// if there are no platforms specified, then we assume
	// that the TTP is compatible with all platforms
	// (even though it probably isn't, but there
	//  are a lot of existing TTPs from before this feature
	// existed that don't explicitly declare supported platforms)
	if len(rc.Platforms) > 0 {
		var ttpIsCompatibleWithCurrentPlatform bool
		for _, platform := range rc.Platforms {
			if platform.IsCompatibleWith(ctx.Platform) {
				ttpIsCompatibleWithCurrentPlatform = true
				break
			}
		}
		if !ttpIsCompatibleWithCurrentPlatform {
			logging.L().Errorf("The current platform %q is not compatible with this TTP", ctx.Platform.String())
			logging.L().Errorf("Supported platforms are:")
			for _, p := range rc.Platforms {
				logging.L().Errorf("\t%v", p.String())
			}
			return fmt.Errorf("the current platform is not compatible with this TTP")
		}
	}

	// check superuser requirement
	if rc.ExpectSuperuser {
		if !isSuperuser() {
			return fmt.Errorf("must be running with elevated privileges to run this TTP")
		}
		if runtime.GOOS == "windows" {
			logging.L().Debug("[+] Running as administrator")
		} else {
			logging.L().Debug("[+] Running as root")
		}
	}
	return nil
}

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
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
)

// CopyPathStep creates a new file and populates it
// with the specified contents from an existing path.
// Its intended use is simulating malicious file copies
// via a C2, where there is no corresponding shell history
// telemetry.
type CopyPathStep struct {
	actionDefaults `yaml:",inline"`
	Source         string   `yaml:"copy_path,omitempty"`
	Destination    string   `yaml:"to,omitempty"`
	Recursive      bool     `yaml:"recursive,omitempty"`
	Overwrite      bool     `yaml:"overwrite,omitempty"`
	Mode           int      `yaml:"mode,omitempty"`
	Direction      string   `yaml:"direction,omitempty"`
	FileSystem     afero.Fs `yaml:"-,omitempty"`
}

// NewCopyPathStep creates a new CopyPathStep instance and returns a pointer to it.
func NewCopyPathStep() *CopyPathStep {
	return &CopyPathStep{}
}

// IsNil checks if the step is nil or empty and returns a boolean value.
func (s *CopyPathStep) IsNil() bool {
	switch {
	case s.Source == "":
		return true
	case s.Destination == "":
		return true
	default:
		return false
	}
}

// Validate validates the step, checking for the necessary attributes and dependencies
func (s *CopyPathStep) Validate(_ TTPExecutionContext) error {
	if s.Source == "" {
		return fmt.Errorf("src field cannot be empty")
	}
	if s.Destination == "" {
		return fmt.Errorf("dest field cannot be empty")
	}
	if s.Direction != "" && s.Direction != "upload" && s.Direction != "download" {
		return fmt.Errorf("invalid direction %q: must be \"upload\" or \"download\"", s.Direction)
	}
	return nil
}

// Template takes each applicable field in the step and replaces any template strings with their resolved values.
//
// **Returns:**
//
// error: error if template resolution fails, nil otherwise
func (s *CopyPathStep) Template(execCtx TTPExecutionContext) error {
	var err error
	s.Source, err = execCtx.templateStep(s.Source)
	if err != nil {
		return err
	}
	s.Destination, err = execCtx.templateStep(s.Destination)
	if err != nil {
		return err
	}
	s.Direction, err = execCtx.templateStep(s.Direction)
	if err != nil {
		return err
	}
	return nil
}

// Execute runs the step and returns an error if one occurs.
func (s *CopyPathStep) Execute(execCtx TTPExecutionContext) (*ActResult, error) {
	logging.L().Infof("Copying file(s) from %v to %v", s.Source, s.Destination)

	var srcFs, dstFs afero.Fs
	if s.FileSystem != nil {
		// Testing path: use injected FS for both sides.
		srcFs = s.FileSystem
		dstFs = s.FileSystem
	} else {
		switch s.Direction {
		case "upload":
			if execCtx.Backend == nil {
				return nil, fmt.Errorf("direction %q requires a remote: block", s.Direction)
			}
			remoteFs, err := execCtx.Backend.GetFs()
			if err != nil {
				return nil, fmt.Errorf("failed to get remote filesystem: %w", err)
			}
			srcFs = afero.NewOsFs()
			dstFs = remoteFs
		case "download":
			if execCtx.Backend == nil {
				return nil, fmt.Errorf("direction %q requires a remote: block", s.Direction)
			}
			remoteFs, err := execCtx.Backend.GetFs()
			if err != nil {
				return nil, fmt.Errorf("failed to get remote filesystem: %w", err)
			}
			srcFs = remoteFs
			dstFs = afero.NewOsFs()
		default:
			// Existing behavior: single FS (local or remote).
			if execCtx.Backend != nil {
				var err error
				srcFs, err = execCtx.Backend.GetFs()
				if err != nil {
					return nil, fmt.Errorf("failed to get filesystem: %w", err)
				}
			} else {
				srcFs = afero.NewOsFs()
			}
			dstFs = srcFs
		}
	}

	// check if source exists.
	sourceExists, err := afero.Exists(srcFs, s.Source)
	if err != nil {
		return nil, err
	}

	// if source does not exist.
	if !sourceExists {
		return nil, fmt.Errorf("source %v does not exist", s.Source)
	}

	// if source is a directory but recursive is false
	srcInfo, err := srcFs.Stat(s.Source)
	if err != nil {
		return nil, err
	}
	if srcInfo.IsDir() && !s.Recursive {
		return nil, fmt.Errorf("source %v is a directory, but the recursive flag is set to false", s.Source)
	}

	// check if destination exists.
	destExists, err := afero.Exists(dstFs, s.Destination)
	if err != nil {
		return nil, err
	}
	// if destination exits, return error if overwrite flag is not true.
	if destExists && !s.Overwrite {
		return nil, fmt.Errorf("dest %v already exists and overwrite was not set", s.Destination)
	}

	// use the default umask
	// https://stackoverflow.com/questions/23842247/reading-default-filemode-when-using-os-o-create
	mode := s.Mode
	if mode == 0 {
		mode = 0666
	}

	// Copy the file or directory
	if srcInfo.IsDir() {
		err = aferoCopyDir(srcFs, dstFs, s.Source, s.Destination)
	} else {
		err = aferoCopyFile(srcFs, dstFs, s.Source, s.Destination, os.FileMode(mode))
	}
	if err != nil {
		return nil, err
	}

	return &ActResult{}, nil
}

// aferoCopyFile copies a single file, reading from srcFs and writing to dstFs.
func aferoCopyFile(srcFs, dstFs afero.Fs, src, dst string, mode os.FileMode) error {
	contents, err := afero.ReadFile(srcFs, src)
	if err != nil {
		return fmt.Errorf("failed to read source %s: %w", src, err)
	}
	return afero.WriteFile(dstFs, dst, contents, mode)
}

// aferoCopyDir recursively copies a directory, reading from srcFs and writing to dstFs.
func aferoCopyDir(srcFs, dstFs afero.Fs, src, dst string) error {
	return afero.Walk(srcFs, src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from source and construct destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return dstFs.MkdirAll(dstPath, info.Mode())
		}

		return aferoCopyFile(srcFs, dstFs, path, dstPath, info.Mode())
	})
}

// GetDefaultCleanupAction will instruct the calling code
// to remove the path created by this action
func (s *CopyPathStep) GetDefaultCleanupAction() Action {
	cleanup := &RemovePathAction{
		Path:      s.Destination,
		Recursive: s.Recursive,
	}
	if s.Direction == "download" {
		// Destination is local — pin to local FS so cleanup doesn't
		// accidentally remove from the remote host.
		cleanup.FileSystem = afero.NewOsFs()
	}
	return cleanup
}

// CanBeUsedInCompositeAction enables this action to be used in a composite action
func (s *CopyPathStep) CanBeUsedInCompositeAction() bool {
	return true
}

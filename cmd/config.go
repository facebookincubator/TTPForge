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

package cmd

import (
	// 'go lint': need blank import for embedding default config
	_ "embed"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var (
	autoInitConfig bool
	//go:embed default-config.yaml
	defaultConfigContents string
	defaultConfigFileName = "config.yaml"
	defaultResourceDir    = ".ttpforge"
)

func ensureDefaultConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	f := afero.NewOsFs()
	defaultConfigDir := filepath.Join(homeDir, defaultResourceDir)
	defaultConfigPath := filepath.Join(defaultConfigDir, defaultConfigFileName)
	exists, err := afero.Exists(f, defaultConfigPath)
	if err != nil {
		return "", err
	}
	if exists {
		return defaultConfigPath, nil
	}

	if err = os.MkdirAll(defaultConfigDir, 0700); err != nil {
		return "", err
	}

	if err := afero.WriteFile(f, defaultConfigPath, []byte(defaultConfigContents), 0600); err != nil {
		return "", err
	}
	return defaultConfigPath, nil
}

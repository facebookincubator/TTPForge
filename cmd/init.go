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
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func copyEmbeddedConfigToPath(configFilePath string) error {
	cfgDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return err
	}
	err := afero.WriteFile(afero.NewOsFs(), configFilePath, []byte(defaultConfigContents), 0600)
	return err
}

func buildInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize TTPForge and Pull Down ForgeArmory",
		Long: `
TTPForge is a Purple Team engagement tool to execute Tactics, Techniques, and Procedures.
    `,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			defaultConfigFilePath, err := getDefaultConfigFilePath()
			if err != nil {
				return fmt.Errorf("could not lookup default config file path: %v", err)
			}
			exists, err := afero.Exists(afero.NewOsFs(), defaultConfigFilePath)
			if err != nil {
				return fmt.Errorf("could not check existence of file %v: %v", defaultConfigFilePath, err)
			}
			if exists {
				logging.L().Warnf("Configuration file %v already exists - TTPForge is probably already initialized", defaultConfigFilePath)
				logging.L().Warn("If you really want to re-initialize it, delete all existing TTPForge configuration files/directories")
				return nil
			}

			logging.L().Infof("Copying embedded configuration file to path: %v", defaultConfigFilePath)
			err = copyEmbeddedConfigToPath(defaultConfigFilePath)
			if err != nil {
				return fmt.Errorf("could not copy default configuration to path %v: %v", defaultConfigFilePath, err)
			}

			logging.L().Info("Loading configuration from %v", defaultConfigFilePath)
			logging.L().Info("Initializing all TTP repositories...")
			cfg := &Config{}
			err = cfg.init()
			if err != nil {
				return fmt.Errorf("failed to initialize TTPForge configuration: %w", err)
			}

			logging.L().Infof("TTPForge Initialized. Now try `ttpforge list ttps` :)")
			return nil
		},
	}
}

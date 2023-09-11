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
	"bytes"
	// needed for embedded filesystem
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// Config stores the variables from the TTPForge global config file
type Config struct {
	RepoSpecs []repos.Spec `yaml:"repos"`

	repoCollection repos.RepoCollection
	cfgFile        string
}

var (
	//go:embed default-config.yaml
	defaultConfigContents string
	defaultConfigFileName = "config.yaml"
	defaultResourceDir    = ".ttpforge"

	// Conf refers to the configuration used throughout TTPForge.
	Conf = &Config{}

	logConfig logging.Config
)

func getDefaultConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	defaultConfigPath := filepath.Join(homeDir, defaultResourceDir, defaultConfigFileName)
	return defaultConfigPath, nil
}

// loadRepoCollection verifies that all repositories specified
// in the configuration file are present on the filesystem
// and clones missing ones if needed
func (cfg *Config) loadRepoCollection() (repos.RepoCollection, error) {
	// locate our config file directory to expend config-relative paths
	var basePath string
	if cfg.cfgFile != "" {
		cfgFileAbsPath, err := filepath.Abs(cfg.cfgFile)
		if err != nil {
			return nil, err
		}
		basePath = filepath.Dir(cfgFileAbsPath)
	}
	fsys := afero.NewOsFs()
	return repos.NewRepoCollection(fsys, cfg.RepoSpecs, basePath)
}

// save() writes the current config back to its file - used by `install“ command
func (cfg *Config) save() error {
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	err := yamlEncoder.Encode(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling config failed: %v", err)
	}
	// YAML won't add this stylistic choice so we do it ourselves
	cfgStr := "---\n" + b.String()
	err = os.WriteFile(cfg.cfgFile, []byte(cfgStr), 0)
	return err
}

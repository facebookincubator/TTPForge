package repos

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RepoConfigFileName is the expected name of the configuration
// file for a TTP repository such as ForgeArmory.
// We export it for tests
const RepoConfigFileName = "ttpforge-repo-config.yaml"

type Spec struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Config struct {
	TTPSearchPaths []string `yaml:"ttp_search_paths"`
}

func LoadConfigs(fsys fs.FS, specs []Spec) ([]Config, error) {

	var repoConfigs []Config
	for _, spec := range specs {
		configPath := filepath.Join(spec.Path, RepoConfigFileName)
		contents, err := fs.ReadFile(fsys, configPath)
		if err != nil {
			return nil, fmt.Errorf("could not read repo config at path %v: %v", configPath, err)
		}
		var curRepoConfig Config
		err = yaml.Unmarshal(contents, &curRepoConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid config file found at %v: %v", configPath, err)
		}
		repoConfigs = append(repoConfigs, curRepoConfig)
	}
	return repoConfigs, nil
}

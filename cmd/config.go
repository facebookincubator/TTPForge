package cmd

import (
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

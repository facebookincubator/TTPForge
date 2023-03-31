package blocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var InventoryPath []string

func FetchAbs(path string, workdir string) (fullpath string, err error) {
	if strings.HasPrefix(path, "~/") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return path, err
		}
		Logger.Sugar().Debugw("homedir", "dir", dirname)
		fullpath, err = filepath.Abs(filepath.Join(dirname, path[2:]))
		if err != nil {
			return path, err
		}
	} else if filepath.IsAbs(path) {
		fullpath = path
	} else {
		fullpath, err = filepath.Abs(filepath.Join(workdir, path))
		if err != nil {
			return path, err
		}
	}

	return fullpath, nil

}

func CheckExist(path string, workdir string, system fs.StatFS) (foundPath string, err error) {
	// Case where we have home expansion
	// Case where filepath starts with frontslash
	// Case where no start slash, so we interpret as relative
	// TODO: break up into windows, linux compiled files to handle path issues
	if system != nil {
		if _, err := system.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		foundPath = path
	} else if filepath.IsAbs(path) {
		Logger.Sugar().Debugw("absolute path found", "path", path)
		// don't allow relative paths when searching inventory,
		// only allow in the working dir
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		foundPath = path

	} else if strings.HasPrefix(path, "../") || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "..\\") {
		Logger.Sugar().Debug("path contains relative ../, using workdir explicitly")
		// don't allow relative paths when searching inventory,
		// only allow in the working dir
		tmppath, err := FetchAbs(path, workdir)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(tmppath); errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		foundPath = path

	} else {
		Logger.Sugar().Debug("path is relative without prefix, searching inventory")
		// check workdir first, takes precedence
		tmppath, err := FetchAbs(path, workdir)
		if err != nil {
			return "", err
		}
		if file, err := os.Stat(tmppath); !errors.Is(err, fs.ErrNotExist) && file.Size() > 0 {
			Logger.Sugar().Debugw("found", "path", tmppath)
			return tmppath, nil
		}
		// then check in the list of paths
		for _, dir := range InventoryPath {
			tmppath, err = FetchAbs(path, dir)
			if err != nil {
				return "", err
			}
			Logger.Sugar().Debugw("searching", "path", tmppath)
			if file, err := os.Stat(tmppath); !errors.Is(err, fs.ErrNotExist) && file.Size() > 0 {
				Logger.Sugar().Debugw("found", "path", tmppath)
				return tmppath, nil
			}
		}
		return "", errors.New(fmt.Sprintf("invalid path provided %s", path))
	}

	return
}

func FetchEnv(environ map[string]string) []string {
	var envSlice []string
	for k, v := range environ {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

func JsonString(in any) (string, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("'%s'", string(out)), nil
}

func contains(key string, search map[string]any) bool {
	if _, ok := search[key]; ok {
		return true
	}
	return false
}

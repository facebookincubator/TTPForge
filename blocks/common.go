package blocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// InventoryPath is a list of paths to search when checking for file existence.
var InventoryPath []string

// FetchAbs returns the absolute path of a file given its path and the working directory. It handles cases where the path starts with "~/",
// is an absolute path, or is a relative path from the working directory. It logs any errors and returns them.
//
// Parameters:
//
// path: A string representing the path to the file.
// workdir: A string representing the working directory.
//
// Returns:
//
// fullpath: A string representing the absolute path to the file.
// error: An error if the path cannot be resolved to an absolute path.
func FetchAbs(path string, workdir string) (fullpath string, err error) {
	if strings.HasPrefix(path, "~/") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			Logger.Sugar().Errorw("failed to get home dir", zap.Error(err))
			return path, err
		}
		Logger.Sugar().Debugw("homedir", "dir", dirname)
		fullpath, err = filepath.Abs(filepath.Join(dirname, path[2:]))
		if err != nil {
			Logger.Sugar().Errorw("failed to get fullpath", zap.Error(err))
			return path, err
		}
	} else if filepath.IsAbs(path) {
		fullpath = path
	} else {
		fullpath, err = filepath.Abs(filepath.Join(workdir, path))
		if err != nil {
			Logger.Sugar().Errorw("failed to get fullpath", zap.Error(err))
			return path, err
		}
	}

	return fullpath, nil

}

// CheckExist checks if a file exists given its path, the working directory, and an optional fs.StatFS. It handles cases where the path starts with "../",
// "~/", or is a relative path. It also checks a list of paths in InventoryPath for the file. It logs any errors and returns them.
//
// Parameters:
//
// path: A string representing the path to the file.
// workdir: A string representing the working directory.
// system: An optional fs.StatFS that can be used to check if the file exists.
//
// Returns:
//
// foundPath: A string representing the path to the file if it exists.
// error: An error if the file cannot be found.
func CheckExist(path string, workdir string, system fs.StatFS) (foundPath string, err error) {
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
			Logger.Sugar().Error("failed to execute FetchAbs() on %s in %s: %v", path, workdir, zap.Error(err))
			return "", err
		}
		if _, err := os.Stat(tmppath); errors.Is(err, fs.ErrNotExist) {
			Logger.Sugar().Error(zap.Error(err))
			return "", err
		}
		foundPath = path

	} else {
		Logger.Sugar().Debug("path is relative without prefix, searching inventory")
		// check workdir first, takes precedence
		tmppath, err := FetchAbs(path, workdir)
		if err != nil {
			Logger.Sugar().Error("failed to execute FetchAbs() on %s in %s: %v", path, workdir, zap.Error(err))
			return "", err
		}
		if file, err := os.Stat(tmppath); !errors.Is(err, fs.ErrNotExist) && file.Size() > 0 {
			Logger.Sugar().Debugw("found", "path", tmppath)
			return tmppath, nil
		} else {
			Logger.Sugar().Error(zap.Error(err))
			return "", err
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

// FetchEnv converts an environment variable map into a slice of strings that can be used as an argument when running a command.
//
// Parameters:
//
// environ: A map of environment variable names to values.
//
// Returns:
//
// []string: A slice of strings representing the environment variables and their values.
func FetchEnv(environ map[string]string) []string {
	var envSlice []string

	for k, v := range environ {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	return envSlice
}

// JSONString returns a string representation of an object in JSON format.
//
// Parameters:
//
// in: An object of any type.
//
// Returns:
//
// string: A string representing the object in JSON format.
// error: An error if the object cannot be encoded as JSON.
func JSONString(in any) (string, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("'%s'", string(out)), nil
}

// Contains checks if a key exists in a map.
//
// Parameters:
//
// key: A string representing the key to search for.
// search: A map of keys and values.
//
// Returns:
//
// bool: A boolean value indicating if the key was found in the map.
func Contains(key string, search map[string]any) bool {
	if _, ok := search[key]; ok {
		return true
	}

	return false
}

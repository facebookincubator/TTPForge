package blocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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
	if path == "" {
		err = errors.New("empty path provided")
		Logger.Sugar().Errorw("failed to get fullpath", zap.Error(err))
		return path, err
	}

	var basePath string
	switch {
	case strings.HasPrefix(path, "~/"):
		basePath, err = os.UserHomeDir()
		if err != nil {
			Logger.Sugar().Errorw("failed to get home dir", zap.Error(err))
			return path, err
		}
		path = path[2:]
	case filepath.IsAbs(path):
		basePath = ""
	default:
		basePath = workdir

		// Remove the common prefix between path and workdir, if any
		relPath, err := filepath.Rel(workdir, path)
		if err == nil {
			path = relPath
		}
	}

	fullpath, err = filepath.Abs(filepath.Join(basePath, path))
	if err != nil {
		Logger.Sugar().Errorw("failed to get fullpath", zap.Error(err))
		return path, err
	}

	Logger.Sugar().Debugw("Full path: ", "fullpath", fullpath)
	return fullpath, nil
}

// FindFilePath checks if a file exists given its path, the working directory, and an optional fs.StatFS. It handles cases where the path starts with "../",
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
func FindFilePath(path string, workdir string, system fs.StatFS) (foundPath string, err error) {
	Logger.Sugar().Debugw("Attempting to find file path", "path", path, "workdir", workdir)

	// Check if file exists using provided fs.StatFS
	if system != nil {

		if _, err := system.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			Logger.Sugar().Debugw("File found using provided fs.StatFS", "path", path)
			return path, nil
		}

		Logger.Sugar().Errorw("file not found using provided fs.StatFS", "path", path, zap.Error(err))
		return "", err
	}

	// Handle home directory representation in Windows.
	if strings.HasPrefix(path, "~/") || (runtime.GOOS == "windows" && strings.HasPrefix(path, "%USERPROFILE%")) {
		if runtime.GOOS == "windows" {
			path = strings.Replace(path, "%USERPROFILE%", "~", 1)
		}
	} else {
		// Convert path to lowercase for Windows, which employs case-insensitive file systems.
		if runtime.GOOS == "windows" {
			path = strings.ToLower(path)
		}
	}

	// Resolve the input path to an absolute path
	absPath, err := FetchAbs(path, workdir)
	if err != nil {
		Logger.Sugar().Errorw("failed to fetch absolute path", "path", path, "workdir", workdir, zap.Error(err))
		return "", err
	}

	// Check if the absolute path exists
	if _, err := os.Stat(absPath); !errors.Is(err, fs.ErrNotExist) {
		Logger.Sugar().Debugw("File found in absolute path", "absPath", absPath)
		return absPath, nil
	}

	// If the path is not found, search for the file in the InventoryPath list
	for _, dir := range InventoryPath {
		inventoryPath, err := FetchAbs(path, dir)
		if err != nil {
			Logger.Sugar().Errorw("failed to fetch absolute path in inventory", "path", path, "dir", dir, zap.Error(err))
			return "", err
		}

		if _, err := os.Stat(inventoryPath); !errors.Is(err, fs.ErrNotExist) {
			Logger.Sugar().Debugw("File found in inventory path", "inventoryPath", inventoryPath)
			return inventoryPath, nil
		}
	}

	// If the file is not found in any of the locations, return an error
	err = fmt.Errorf("invalid path %s provided", path)
	Logger.Sugar().Errorw("file not found in any location", "path", path, zap.Error(err))
	return "", err
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
		Logger.Sugar().Errorw(err.Error(), zap.Error(err))
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
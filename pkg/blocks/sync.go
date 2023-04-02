package blocks

import (
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// FileSystem is a global fs.FS instance used for file system operations.
var FileSystem fs.FS

// LoadTTP reads a TTP file and creates a TTP instance based on its contents.
// If the file is empty or contains invalid data, it returns an error.
//
// Parameters:
//
// filename: A string representing the path to the TTP file.
//
// Returns:
//
// loadedTTPs: The created TTP instance, or an empty TTP if the file is empty or invalid.
// err: An error if the file contains invalid data or cannot be read.
func LoadTTP(filename string) (loadedTTPs TTP, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &loadedTTPs)
	if err != nil {
		return
	}

	return
}

// Actors represents the various actors involved in the execution of TTPs.
// It contains Outputs and Environment, which are maps for storing output
// data and environment variables.
type Actors struct {
	Outputs     map[string]any    // Outputs stores the results generated during the execution of TTPs.
	Environment map[string]string // Environment stores the environment variables required for TTPs execution.
}

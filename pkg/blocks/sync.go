package blocks

import (
	"io"
	"io/fs"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var Logger *zap.Logger

var FileSystem fs.FS

func LoadTTP(filename string) (loadedTTPs TTP, err error) {
	file, err := os.Open(filename) // For read access.
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

type Actors struct {
	Outputs     map[string]any
	Environment map[string]string
}

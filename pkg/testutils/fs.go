package testutils

import (
	"path/filepath"

	"github.com/spf13/afero"
)

func MakeAferoTestFs(filesMap map[string][]byte) (afero.Fs, error) {
	fsys := afero.NewMemMapFs()
	for path, contents := range filesMap {
		dirPath := filepath.Dir(path)
		err := fsys.MkdirAll(dirPath, 0700)
		if err != nil {
			return nil, err
		}
		err = afero.WriteFile(fsys, path, contents, 0644)
		if err != nil {
			return nil, err
		}
	}
	return fsys, nil
}

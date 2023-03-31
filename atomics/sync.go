package atomics

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func init() {
	actionOutputs = make(map[string]string)
}

var actionOutputs map[string]string

var Logger *zap.Logger

type ActionMethod string

func readTTPsFromFile(filename string) (ttps TTP, err error) {
	file, err := os.Open(filename) // For read access.
	if err != nil {
		return
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &ttps)
	return
}

func ExpandNestedSteps(ttp *TTP, TTPSearchPaths []string) ([]*Step, error) {
	expandedSteps := []*Step{}
	hd, _ := os.UserHomeDir()
	for _, step := range ttp.Steps {
		if step.TTP != "" {

			// locate the TTP in our search path
			var ttpPath string
			for _, sp := range TTPSearchPaths {

				// TODO: move this into search path initialization
				// right after config load - shouldn't be here, '~/' should
				// only be replaced once
				candidate := filepath.Join(sp, step.TTP)
				if strings.HasPrefix(candidate, "~/") {
					candidate = strings.Replace(candidate, "~/", hd+"/", 1)
				}

				if _, err := os.Stat(candidate); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					} else {
						return nil, fmt.Errorf("error calling Stat(%v): %v", step.TTP, err)
					}
				} else {
					ttpPath = candidate
					break
				}
			}
			if ttpPath == "" {
				return nil, fmt.Errorf("could not find TTP file: %v", step.TTP)
			}

			tmpTTP, err := readTTPsFromFile(ttpPath)
			if err != nil {
				return nil, err
			}
			for _, newStep := range tmpTTP.Steps {

				// TODO: need to copy all other relevant fields as well
				newStep.Environment = step.Environment
				if newStep.Cleanup != nil {
					newStep.Cleanup.Environment = step.Environment
				}
				expandedSteps = append(expandedSteps, newStep)
			}
		} else {
			expandedSteps = append(expandedSteps, step)
		}
	}
	return expandedSteps, nil
}

func LoadTTP(filename string) (*TTP, error) {

	// initial because we may mangle it with
	// secondary unmarshal steps to load `ttp:` directives
	ttp, err := readTTPsFromFile(filename)
	if err != nil {
		return nil, err
	}

	// substitute WORKDIR in environment variables if needed
	// purpose of this feature is to support environment variables
	// that reference files produced by earlier steps, such as:
	//
	// env:
	//	THRIFT_TLS_CL_KEY_PATH: ${WORKDIR}/fbpkg-stolen-tls-cert.pem
	//	THRIFT_TLS_CL_CERT_PATH: ${WORKDIR}/fbpkg-stolen-tls-cert.pem
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for _, step := range ttp.Steps {
		for k, v := range step.Environment {
			// TODO: make this use specially-created workdir, not just cwd
			step.Environment[k] = strings.Replace(v, "${WORKDIR}", wd, 1)
		}
	}

	ttp.altBaseDir = filepath.Dir(filename)

	return &ttp, nil
}

type Actors struct {
	Outputs     map[string]any
	Environment map[string]string
}

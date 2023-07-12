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

package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/files"
	cp "github.com/otiai10/copy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ScenarioResult struct {
	FileContents map[string]string
}

func runE2ETest(t *testing.T, ttpRelPath string, expectedResult ScenarioResult, opts ...blocks.TTPExecutionConfig) {
	// learned this trick from here
	var ttpConfig blocks.TTPExecutionConfig
	if len(opts) > 0 {
		ttpConfig = opts[0]
	}

	const testTTPDir = "test-ttps"

	// create temporary working directory
	testDir, err := os.MkdirTemp("", "ttpforge-e2e-test")
	require.NoError(t, err, "failed to create temporary directory")
	defer func() {
		err := os.RemoveAll(testDir)
		require.NoError(t, err, "failed to delete test directory")
	}()

	// copy the entire tree so that TTP relative paths work - just as if we
	// were running the real command E2E
	err = cp.Copy(testTTPDir, filepath.Join(testDir, testTTPDir))
	require.NoError(t, err, "failed to copy TTPs")

	// execute the test from the temporary directory to keep things clean
	curDir, err := os.Getwd()
	require.NoError(t, err, "failed to get current directory")
	workDir := filepath.Join(testDir, testTTPDir)
	err = os.Chdir(workDir)
	require.NoError(t, err, "failed to chdir to test directory")
	defer func() {
		err := os.Chdir(curDir)
		require.NoError(t, err, "failed to chdir back to former current directory")
	}()

	_, err = files.ExecuteYAML(ttpRelPath, ttpConfig, nil)
	require.NoError(t, err, "failed to execute TTP")

	// validate that correct files were generated
	ttpDir := filepath.Dir(ttpRelPath)
	for fileName, expectedContents := range expectedResult.FileContents {
		fileRelPath := filepath.Join(ttpDir, fileName)
		r, err := os.Open(fileRelPath)
		require.Nil(t, err)

		contents, err := io.ReadAll(r)
		require.Nil(t, err)
		assert.Equal(t, expectedContents, string(contents))
	}
}

func TestVariableExpansion(t *testing.T) {
	dirname, err := os.UserHomeDir()
	require.Nil(t, err)

	ttpPath := filepath.Join("variable-expansion", "ttp.yaml")
	resultLines := []string{
		fmt.Sprintf("{\"test_key\":\"%v\n\",\"another_key\":\"wut\"}", dirname),
		"you said: foo",
		"cleaning up now",
	}
	runE2ETest(t, ttpPath, ScenarioResult{
		FileContents: map[string]string{
			"result.txt": strings.Join(resultLines, "\n") + "\n",
		},
	})
}

func TestRelativePaths(t *testing.T) {
	ttpPath := filepath.Join("relative-paths", "very", "nested", "ttp.yaml")
	runE2ETest(t, ttpPath, ScenarioResult{
		FileContents: map[string]string{
			"result.txt": "A\nB\nC\nD\nE\n",
		},
	})
}

func TestNoCleanup(t *testing.T) {
	ttpPath := filepath.Join("relative-paths", "very", "nested", "ttp.yaml")
	runE2ETest(t, ttpPath, ScenarioResult{
		FileContents: map[string]string{
			"result.txt": "A\nB\nC\n",
		},
	}, blocks.TTPExecutionConfig{
		NoCleanup: true,
	})
}

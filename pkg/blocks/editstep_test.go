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

package blocks_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalEditValid(t *testing.T) {
	content := `name: test
description: this is a test
steps:
  - name: valid_edit
    edit_file: yolo
    edits:
      - old: foo
        new: yolo
      - old: another
        new: one`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	err = ttp.Validate(blocks.TTPExecutionContext{})
	require.NoError(t, err)
}

func TestUnmarshalEditNoNew(t *testing.T) {
	content := `name: test
description: this is a test
steps:
  - name: missing_new
    edit_file: yolo
    edits:
      - old: foo
        new: yolo
      - old: wutwut
      - old: another
        new: one`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	err = ttp.Validate(blocks.TTPExecutionContext{})
	require.Error(t, err)

	assert.Equal(t, "edit #2 is missing 'new:'", err.Error())
}

func TestUnmarshalEditNoOld(t *testing.T) {
	content := `name: test
description: this is a test
steps:
  - name: missing_old
    edit_file: yolo
    edits:
      - new: yolo
      - old: wutwut
        new: haha
      - old: another
        new: one`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	err = ttp.Validate(blocks.TTPExecutionContext{})
	require.Error(t, err)

	assert.Equal(t, "edit #1 is missing 'old:'", err.Error())
}

func TestUnmarshalNonListEdits(t *testing.T) {
	content := `name: test
description: this is a test
steps:
  - name: non_list_edits
    edit_file: yolo
    edits: haha`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.Error(t, err)
}

func TestUnmarshalEmptyEdits(t *testing.T) {
	content := `name: test
description: this is a test
steps:
  - name: no_edits
    edit_file: yolo`

	var ttp blocks.TTP
	err := yaml.Unmarshal([]byte(content), &ttp)
	require.NoError(t, err)

	err = ttp.Validate(blocks.TTPExecutionContext{})
	assert.Equal(t, "no edits specified", err.Error())
}

func TestExecuteSimple(t *testing.T) {
	content := `
name: valid_edit
edit_file: a.txt
edits:
  - old: foo
    new: yolo
  - old: another
    new: one`

	var step blocks.EditStep
	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	testFs := afero.NewMemMapFs()
	err = afero.WriteFile(testFs, "a.txt", []byte("foo\nanother"), 0644)
	require.NoError(t, err)
	step.FileSystem = testFs

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	_, err = step.Execute(execCtx)
	require.NoError(t, err)

	contents, err := afero.ReadFile(testFs, "a.txt")
	require.NoError(t, err)

	assert.Equal(t, "yolo\none", string(contents))
}

func TestExecuteBackupFile(t *testing.T) {
	content := `
name: valid_edit
edit_file: a.txt
backup_file: backup.txt
edits:
  - old: foo
    new: yolo
  - old: another
    new: one`

	var step blocks.EditStep
	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	testFs := afero.NewMemMapFs()
	origContents := []byte("foo\nanother")
	err = afero.WriteFile(testFs, "a.txt", origContents, 0644)
	require.NoError(t, err)
	step.FileSystem = testFs

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	_, err = step.Execute(execCtx)
	require.NoError(t, err)

	backupContents, err := afero.ReadFile(testFs, "backup.txt")
	require.NoError(t, err)

	assert.Equal(t, string(origContents), string(backupContents))
}

func TestExecuteMultiline(t *testing.T) {
	content := `name: delete_function
edit_file: b.txt
edits:
  - old: (?ms:^await MyAwesomeClass\.myAwesomeAsyncFn\(.*?\)$)
    new: "# function call removed by TTP"
    regexp: true`

	fileContentsToEdit := `otherstuff
await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)
moarawesomestuff`

	correctResult := `otherstuff
# function call removed by TTP
moarawesomestuff`

	var step blocks.EditStep
	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	testFs := afero.NewMemMapFs()
	err = afero.WriteFile(testFs, "b.txt", []byte(fileContentsToEdit), 0644)
	require.NoError(t, err)
	step.FileSystem = testFs
	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	_, err = step.Execute(execCtx)
	require.NoError(t, err)

	contents, err := afero.ReadFile(testFs, "b.txt")
	require.NoError(t, err)

	assert.Equal(t, correctResult, string(contents))
}

func TestExecuteVariableExpansion(t *testing.T) {
	content := `name: delete_function
edit_file: b.txt
edits:
  - old: (?P<fn_call>(?ms:^await MyAwesomeClass\.myAwesomeAsyncFn\(.*?\)$))
    new: "/*${fn_call}*/"
    regexp: true`

	fileContentsToEdit := `otherstuff
await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)
moarawesomestuff`

	correctResult := `otherstuff
/*await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)*/
moarawesomestuff`

	var step blocks.EditStep
	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	testFs := afero.NewMemMapFs()
	err = afero.WriteFile(testFs, "b.txt", []byte(fileContentsToEdit), 0644)
	require.NoError(t, err)
	step.FileSystem = testFs

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	_, err = step.Execute(execCtx)
	require.NoError(t, err)

	contents, err := afero.ReadFile(testFs, "b.txt")
	require.NoError(t, err)

	assert.Equal(t, correctResult, string(contents))
}

func TestExecuteNotFound(t *testing.T) {
	content := `name: delete_function
edit_file: b.txt
edits:
  - old: not_going_to_find_this
    new: will_not_be_used`

	fileContentsToEdit := `hello`

	var step blocks.EditStep
	err := yaml.Unmarshal([]byte(content), &step)
	require.NoError(t, err)

	testFs := afero.NewMemMapFs()
	err = afero.WriteFile(testFs, "b.txt", []byte(fileContentsToEdit), 0644)
	require.NoError(t, err)
	step.FileSystem = testFs

	var execCtx blocks.TTPExecutionContext
	err = step.Validate(execCtx)
	require.NoError(t, err)

	_, err = step.Execute(blocks.TTPExecutionContext{})
	require.Error(t, err, "not finding a search string should result in an error")
	assert.Equal(t, "pattern 'not_going_to_find_this' from edit #1 was not found in file b.txt", err.Error())
}

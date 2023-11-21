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

package blocks

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/testutils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEditStep(t *testing.T) {

	testCases := []struct {
		name                      string
		content                   string
		wantUnmarshalError        bool
		wantValidateError         bool
		wantExecuteError          bool
		expectedContentsAfterEdit string
		expectedErrTxt            string
		fsysContents              map[string][]byte
	}{
		{
			name: "Test Unmarshal Edit Valid",
			content: `
edit_file: /tmp/yolo
edits:
  - old: foo
    new: yolo
  - old: another
    new: one`,
			fsysContents:              map[string][]byte{"/tmp/yolo": []byte("foo\nanother")},
			expectedContentsAfterEdit: "yolo\none",
		},
		{
			name: "Test Unmarshal No New",
			content: `
edit_file: yolo
edits:
  - old: foo
    new: yolo
  - old: wutwut
  - old: another
    new: one`,
			wantValidateError: true,
			expectedErrTxt:    "edit #2 is missing 'new:'",
		},
		{
			name: "Test Unmarshal No Old",
			content: `
edit_file: yolo
edits:
  - new: yolo
  - old: wutwut
    new: haha
  - old: another
    new: one`,
			wantValidateError: true,
			expectedErrTxt:    "edit #1 is missing 'old:'",
		},
		{
			name: "Test Unmarshal Empty Edits",
			content: `
steps:
  - name: no_edits
    edit_file: yolo`,
			wantValidateError: true,
			expectedErrTxt:    "no edits specified",
		},
		{
			name: "Test Execute Simple",
			content: `
edit_file: a.txt
edits:
  - old: foo
    new: yolo
  - old: another
    new: one`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "yolo\none",
		},
		{
			name: "Test Execute Backup File",
			content: `
edit_file: a.txt
backup_file: backup.txt
edits:
  - old: foo
    new: yolo
  - old: another
    new: one`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "yolo\none",
		},
		{
			name: "Test Execute Multiline",
			content: `name: delete_function
edit_file: b.txt
edits:
  - old: (?ms:^await MyAwesomeClass\.myAwesomeAsyncFn\(.*?\)$)
    new: "# function call removed by TTP"
    regexp: true`,
			fsysContents: map[string][]byte{"b.txt": []byte(`others
await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)
moarawesomestuff`)},
			expectedContentsAfterEdit: `others
# function call removed by TTP
moarawesomestuff`,
		},
		{
			name: "Test Variable Expansion",
			content: `name: delete_function
edit_file: b.txt
edits:
  - old: (?P<fn_call>(?ms:^await MyAwesomeClass\.myAwesomeAsyncFn\(.*?\)$))
    new: "/*${fn_call}*/"
    regexp: true`,
			fsysContents: map[string][]byte{"b.txt": []byte(`otherstuff
await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)
moarawesomestuff`)},
			expectedContentsAfterEdit: `otherstuff
/*await MyAwesomeClass.myAwesomeAsyncFn(
	param1,
	param2,
)*/
moarawesomestuff`,
		},
		{
			name: "Test Execute Not Found",
			content: `name: delete_function
edit_file: b.txt
edits:
  - old: not_going_to_find_this
    new: will_not_be_used`,
			fsysContents:     map[string][]byte{"b.txt": []byte("not_goung_to_find_this")},
			wantExecuteError: true,
			expectedErrTxt:   "pattern 'not_going_to_find_this' from edit #1 was not found in file b.txt",
		},
		{
			name: "Test Append Old",
			content: `name: test_append_old
edit_file: yolo
edits:
  - old: help
    append: nooooo`,
			wantValidateError: true,
			expectedErrTxt:    "append is not to be used in conjunction with 'old:'",
		},
		{
			name: "Test Append New",
			content: `name: test_append_new
edit_file: yolo
edits:
  - new: me
    append: nooooo`,
			wantValidateError: true,
			expectedErrTxt:    "append is not to be used in conjunction with 'new:'",
		},
		{
			name: "Test Append Regex",
			content: `name: test_append_regex
edit_file: yolo
edits:
  - old:
    new:
    append: nooooo
    regexp: true`,
			wantValidateError: true,
			expectedErrTxt:    "append is not to be used in conjunction with 'regexp:'",
		},
		{
			name: "Test Delete Old",
			content: `name: delete_old
edit_file: yolo
edits:
  - old: help
    new: me
    delete: you`,
			wantValidateError: true,
			expectedErrTxt:    "delete is not to be used in conjunction with 'old:'",
		},
		{
			name: "Test Delete New",
			content: `name: delete_new
edit_file: yolo
edits:
  - new: me
    delete: you`,
			wantValidateError: true,
			expectedErrTxt:    "delete is not to be used in conjunction with 'new:'",
		},
		{
			name: "Test Append",
			content: `name: test_append
edit_file: a.txt
edits:
  - append: put this at the end`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "foo\nanother\nput this at the end",
		},
		{
			name: "Test Multiple Appends",
			content: `name: test_multiappend
edit_file: a.txt
edits:
  - old: foo
    new: bar
  - append: baz
  - old: baz
    new: wut
  - append: yo`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "bar\nanother\nwut\nyo",
		},
		{
			name: "Test Delete",
			content: `name: test_delete
edit_file: a.txt
edits:
  - delete: foo`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "\nanother",
		},
		{
			name: "Test Delete Line",
			content: `name: test_delete_entire_line
edit_file: a.txt
edits:
  - delete: "foo\n"`,
			fsysContents:              map[string][]byte{"a.txt": []byte("foo\nanother")},
			expectedContentsAfterEdit: "another",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var editStep EditStep
			var execCtx TTPExecutionContext

			// parse the step
			err := yaml.Unmarshal([]byte(tc.content), &editStep)
			if tc.wantUnmarshalError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			// validate the step
			err = editStep.Validate(execCtx)
			if tc.wantValidateError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			// prep filesystem
			if tc.fsysContents != nil {
				fsys, err := testutils.MakeAferoTestFs(tc.fsysContents)
				require.NoError(t, err)
				editStep.FileSystem = fsys
			} else {
				editStep.FileSystem = afero.NewMemMapFs()
			}
			originalContent, err := afero.ReadFile(editStep.FileSystem, editStep.FileToEdit)
			require.NoError(t, err)

			// execute the step and check output
			_, err = editStep.Execute(execCtx)
			if tc.wantExecuteError {
				assert.Equal(t, tc.expectedErrTxt, err.Error())
				return
			}
			require.NoError(t, err)

			// Read the contents of the file after edit step execution
			contents, err := afero.ReadFile(editStep.FileSystem, editStep.FileToEdit)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedContentsAfterEdit, string(contents))
			if editStep.BackupFile != "" {
				backupContents, err := afero.ReadFile(editStep.FileSystem, editStep.BackupFile)
				require.NoError(t, err)
				assert.Equal(t, originalContent, backupContents)
			}
		})
	}
}

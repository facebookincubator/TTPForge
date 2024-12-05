/*
Copyright © 2024-present, Meta Platforms, Inc. and affiliates
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	memory []byte
}

func (mw *mockWriter) Write(p []byte) (n int, err error) {
	mw.memory = append(mw.memory, p...)
	return len(p), nil
}

func (mw *mockWriter) GetMemory() []byte {
	return mw.memory
}

func TestBufferedWriter(t *testing.T) {
	testCases := []struct {
		name       string
		textChunks []string
		wantError  bool
	}{
		{
			name:       "Finished line",
			textChunks: []string{"Hello\n"},
			wantError:  false,
		}, {
			name:       "Unfinished line",
			textChunks: []string{"Hello, world", "!\n"},
			wantError:  false,
		}, {
			name:       "No last newline",
			textChunks: []string{"Hello,\nworld", "!\nfoobar"},
			wantError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := &mockWriter{}
			zw := &bufferedWriter{
				writer: mockWriter,
			}
			var totalBytesWritten int
			for _, text := range tc.textChunks {
				bytesWritten, err := zw.Write([]byte(text))
				totalBytesWritten += bytesWritten
				if tc.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
			zw.Close()

			expectedTotalBytesWritten := 0
			for _, text := range tc.textChunks {
				expectedTotalBytesWritten += len(text)
			}
			assert.Equal(t, expectedTotalBytesWritten, totalBytesWritten)
			expectedBytesWritten := []byte{}
			for _, text := range tc.textChunks {
				text = strings.ReplaceAll(text, "\n", "")
				expectedBytesWritten = append(expectedBytesWritten, []byte(text)...)
			}
			assert.Equal(t, expectedBytesWritten, mockWriter.GetMemory())
		})
	}
}

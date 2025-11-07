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
	"bytes"
	"io"
	"os/exec"

	"github.com/facebookincubator/ttpforge/pkg/logging"
)

type bufferedWriter struct {
	buff   bytes.Buffer
	writer io.Writer
}

func (bw *bufferedWriter) Write(b []byte) (n int, err error) {
	n = len(b)
	for len(b) > 0 {
		b = bw.writeLine(b)
	}
	return n, nil
}

func (bw *bufferedWriter) writeLine(line []byte) (remaining []byte) {
	idx := bytes.IndexByte(line, '\n')
	if idx < 0 {
		// If there are no newlines, buffer the entire string.
		bw.buff.Write(line)
		return nil
	}

	// Split on the newline, buffer and flush the left.
	line, remaining = line[:idx], line[idx+1:]

	// Fast path: if we don't have a partial message from a previous write
	// in the buffer, skip the buffer and log directly.
	if bw.buff.Len() == 0 {
		bw.log(line)
		return remaining
	}

	bw.buff.Write(line)
	bw.log(bw.buff.Bytes())
	bw.buff.Reset()
	return remaining
}

func (bw *bufferedWriter) log(line []byte) {
	_, err := bw.writer.Write(line)
	if err != nil {
		logging.L().Errorw("failed to log", "err", err)
	}
}

func (bw *bufferedWriter) Close() error {
	if bw.buff.Len() != 0 {
		bw.log(bw.buff.Bytes())
		bw.buff.Reset()
	}
	return nil
}

type zapWriter struct {
	prefix string
}

func (zw *zapWriter) Write(p []byte) (n int, err error) {
	logging.L().Info(zw.prefix, string(p))
	return len(p), nil
}

func streamAndCapture(cmd exec.Cmd, stdout, stderr io.Writer) (*ActResult, error) {
	if stdout == nil {
		stdout = &bufferedWriter{
			writer: &zapWriter{
				prefix: "[STDOUT] ",
			},
		}
	}
	if stderr == nil {
		stderr = &bufferedWriter{
			writer: &zapWriter{
				prefix: "[STDERR] ",
			},
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(stderr, &stderrBuf)

	err := cmd.Run()

	// Flush any remaining stdout output that doesn't end with newline
	if bw, ok := stdout.(*bufferedWriter); ok {
		bw.Close()
	}
	// Flush any remaining stderr output that doesn't end with newline
	if bw, ok := stderr.(*bufferedWriter); ok {
		bw.Close()
	}

	if err != nil {
		return nil, err
	}
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	result := ActResult{}
	result.Stdout = outStr
	result.Stderr = errStr
	return &result, nil
}

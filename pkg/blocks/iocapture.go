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

type zapWriter struct {
	prefix string
}

func (z *zapWriter) Write(b []byte) (int, error) {
	n := len(b)
	// extra-defensive programming :P
	if n <= 0 {
		return 0, nil
	}

	// strip trailing newline
	if b[n-1] == '\n' {
		b = b[:n-1]
	}

	// split lines
	lines := bytes.Split(b, []byte{'\n'})
	for _, line := range lines {
		logging.L().Info(z.prefix, string(line))
	}
	return n, nil
}

func streamAndCapture(cmd exec.Cmd, stdout, stderr io.Writer) (*ActResult, error) {
	if stdout == nil {
		stdout = &zapWriter{
			prefix: "[STDOUT] ",
		}
	}
	if stderr == nil {
		stderr = &zapWriter{
			prefix: "[STDERR] ",
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(stderr, &stderrBuf)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	result := ActResult{}
	result.Stdout = outStr
	result.Stderr = errStr
	if err != nil {
		return nil, err
	}
	return &result, nil
}

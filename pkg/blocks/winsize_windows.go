//go:build windows

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

import "github.com/creack/pty"

// pty.Winsize's Rows/Cols are uint16 in the fbcode-vendored creack/pty but uint
// in the upstream module's Windows variant, so a fixed cast fails one build or
// the other. setDim infers the field's own integer type and converts to it,
// letting this compile under both the internal buck opt-win build and the OSS
// GOOS=windows cross-compile.
func newWinsize(rows, cols int) pty.Winsize {
	var ws pty.Winsize
	setDim(&ws.Rows, rows)
	setDim(&ws.Cols, cols)
	return ws
}

func setDim[T ~uint | ~uint16](dst *T, v int) {
	*dst = T(v)
}

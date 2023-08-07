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
	"github.com/stretchr/testify/require"
)

func TestLintTTPCorrect(t *testing.T) {
	ttpStr := `name: valid ttp
description: should pass linting
args:
- name: arg1
steps:
- name: step1
  inline: echo "arg value is {{ .Args.arg1 }}"
- name: step2
  inline: echo "step two"`

	_, err := blocks.LintTTP([]byte(ttpStr))
	require.NoError(t, err)
}

func TestLintTTPDuplicateKey(t *testing.T) {
	ttpStr := `name: duplicate step key
description: should fail linting
steps:
- name: step1
  inline: echo "step one"
steps:
- name: step2
  inline: echo "step two"`

	_, err := blocks.LintTTP([]byte(ttpStr))
	require.Error(t, err)
}

func TestLintTTPScrambled(t *testing.T) {
	ttpStr := `name: scrambled ttp
description: should fail linting due to args after steps
steps:
- name: step1
  inline: echo "arg value is {{ .Args.arg1 }}"
- name: step2
  inline: echo "step two"
args:
- name: arg1`

	_, err := blocks.LintTTP([]byte(ttpStr))
	require.Error(t, err)
}

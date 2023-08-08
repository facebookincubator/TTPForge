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

package preprocess

import (
	"errors"
	"regexp"
)

var (
	stepsTopLevelKeyRegexp *regexp.Regexp
	topLevelKeyRegexp      *regexp.Regexp
)

func init() {
	stepsTopLevelKeyRegexp = regexp.MustCompile("(?m)^steps:")
	topLevelKeyRegexp = regexp.MustCompile(`(?m)^[^\s]+:`)
}

// Result contains the TTP contents divided into sections for
// further template processing and unmarshalling.
//
// **Attributes:**
//
// PreambleBytes: A byte slice representing the preamble section of the TTP.
// StepsBytes: A byte slice representing the individual steps within the TTP.
type Result struct {
	PreambleBytes []byte
	StepsBytes    []byte
}

// Parse handles early-stage processing of a TTP. It does two main tasks:
// 1) It lints the TTP.
// 2) It segments the TTP into "not steps" and "steps" sections, both crucial
// for YAML unmarshalling and templating.
//
// **Parameters:**
//
// ttpBytes: A byte slice with the raw TTP for processing.
//
// **Returns:**
//
// *Result: Pointer to Result with parsed preamble and steps.
// error: Error for parsing issues or top-level key arrangement problems.
func Parse(ttpBytes []byte) (*Result, error) {
	// no duplicate keys
	stepTopLevelKeyLocs := stepsTopLevelKeyRegexp.FindAllIndex(ttpBytes, -1)
	if len(stepTopLevelKeyLocs) != 1 {
		return nil, errors.New("the top-level key `steps:` should occur exactly once")
	}
	stepTopLevelKeyLoc := stepTopLevelKeyLocs[0]

	// `steps:` should always be the last top-level key
	topLevelKeyLocs := topLevelKeyRegexp.FindAllIndex(ttpBytes, -1)
	for _, loc := range topLevelKeyLocs {
		if loc[0] > stepTopLevelKeyLoc[0] {
			return nil, errors.New("the top-level key `steps:` should always be the last top-level key in the file")
		}
	}
	return &Result{
		PreambleBytes: ttpBytes[:stepTopLevelKeyLoc[0]],
		StepsBytes:    ttpBytes[stepTopLevelKeyLoc[0]:stepTopLevelKeyLoc[1]],
	}, nil
}

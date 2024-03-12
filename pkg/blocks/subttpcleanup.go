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

// subTTPCleanupAction ensures that individual
// steps of the subTTP are appropriately cleaned up
type subTTPCleanupAction struct {
	actionDefaults
	step *SubTTPStep
}

// IsNil is not needed here, as this is not a user-accessible step type
func (a *subTTPCleanupAction) IsNil() bool {
	return false
}

// Validate is not needed here, as this is not a user-accessible step type
func (a *subTTPCleanupAction) Validate(execCtx TTPExecutionContext) error {
	return nil
}

// Execute will cleanup the subTTP starting from the last successful step
func (a *subTTPCleanupAction) Execute(_ TTPExecutionContext) (*ActResult, error) {
	cleanupResults, err := a.step.ttp.startCleanupForCompletedSteps(*a.step.subExecCtx)
	if err != nil {
		return nil, err
	}
	return aggregateResults(cleanupResults), nil
}

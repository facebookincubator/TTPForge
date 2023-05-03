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

package cmd

import (
	"go.uber.org/zap"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/spf13/cobra"
)

func init() {

	rootCmd.AddCommand(RunTTPCmd())
}

// RunTTPCmd runs an input TTP.
func RunTTPCmd() *cobra.Command {
	var ttp blocks.TTP
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the TTP contained in the specified YAML file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// root level flag that we need to pass - other flags handled in init()
			ttp.InventoryPath = Conf.InventoryPath

			// TTP not actually loaded till this happens
			relativeTTPPath := args[0]
			err := ttp.InitializeFromYAML(relativeTTPPath)
			if err != nil {
				Logger.Sugar().Errorw("failed to load TTP", zap.Error(err))
				return
			}

			if err := ttp.RunSteps(); err != nil {
				Logger.Sugar().Errorw("failed to execute TTP", zap.Error(err))
				return
			}
		},
	}

	// FLAGS
	runCmd.Flags().StringVar(&ttp.WorkDir, "workdir", "", "Working Directory to use for TTP Execution")
	// leads to some awkward double negatives but given that false is bool default value it is better to have
	// NoCleanup than Cleanup trust me
	runCmd.Flags().BoolVar(&ttp.NoCleanup, "no-cleanup", false, "Disable TTP Cleanup (useful for debugging)")

	return runCmd
}

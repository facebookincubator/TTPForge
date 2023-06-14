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

package main

import (
	"fmt"
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	deltaCmd = &cobra.Command{
		Use:   "delta",
		Short: "Execute login flow for account with delta 2FA triggered",
		Run: func(cmd *cobra.Command, args []string) {
			// Create TTP log
			logPath := fmt.Sprintf("%s.log", cmd.Use)
			if err := logging.InitLog(false, logPath, false, false); err != nil {
				logging.Logger.Sugar().Errorf("failed to initialize logger: %v", err)
				cobra.CheckErr(err)
			}

			logging.Logger.Sugar().Infof(
				"Executing %s - %s, please wait...",
				cmd.Use, cmd.Short)
			fmt.Printf("USER: %s", user)
		},
	}

	ignoreCertErrors bool
	headless         bool
	password         string
	target           string
	user             string
	token            string
)

func init() {
	deltaCmd.Flags().StringVar(&user,
		"user", "", "Email address for the delta user")

	deltaCmd.Flags().StringVar(&password,
		"password", "", "Password for the delta user")

	deltaCmd.Flags().StringVar(&token,
		"token", "", "Delta user's 2FA token")
}

func main() {
	if err := deltaCmd.Execute(); err != nil {
		logging.Logger.Sugar().Errorf("%s failed to run: %v", deltaCmd.Short, err)
		os.Exit(1)
	}
}

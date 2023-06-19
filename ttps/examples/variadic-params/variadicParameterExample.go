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
	variadicParamsExampleCmd = &cobra.Command{
		Use:   "variadicParamsExample",
		Short: "Execute variadic parameters example",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("User input: %s\n", user)
			fmt.Printf("Password input: %s\n", password)
		},
	}

	password string
	user     string
)

func init() {
	variadicParamsExampleCmd.Flags().StringVar(&user,
		"user", "", "Email address for the variadicParamsExample user")

	variadicParamsExampleCmd.Flags().StringVar(&password,
		"password", "", "Password for the variadicParamsExample user")
}

func main() {
	if err := variadicParamsExampleCmd.Execute(); err != nil {
		logging.Logger.Sugar().Errorf("%s failed to run: %v", variadicParamsExampleCmd.Short, err)
		os.Exit(1)
	}
}

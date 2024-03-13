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
	"fmt"
	"text/template"

	"github.com/facebookincubator/ttpforge/pkg/platforms"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const ttpTemplate = `---
api_version: 2.0
uuid: {{.UUID}}
name: {{.Placeholder}}
description: |
  {{.Placeholder}}
requirements:
  platforms:
    - os: {{.Platform.OS}}
mitre:
  tactics:
    - {{.Placeholder}}
  techniques:
    - {{.Placeholder}}
  subtechniques:
    - {{.Placeholder}}
steps:
  - name: hello_world
    inline: |
      echo "hello, world"
    cleanup:
      inline: |
        echo "bye-bye, world"
`

func buildCreateTTPCommand() *cobra.Command {
	createTTPCmd := &cobra.Command{
		Use:   "ttp [path/to/ttp.yaml]",
		Short: "Create a new TTP file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			// verify that newTTPFilePath does not already exist
			newTTPFilePath := args[0]
			fsys := afero.NewOsFs()
			exists, err := afero.Exists(fsys, newTTPFilePath)
			if err != nil {
				return fmt.Errorf("failed to check file existence: %w", err)
			}
			if exists {
				return fmt.Errorf("%v already exists", newTTPFilePath)
			}

			// create the new TTP file with afero
			f, err := fsys.Create(newTTPFilePath)
			if err != nil {
				return fmt.Errorf("failed to create new TTP file: %w", err)
			}
			defer f.Close()

			// render the TTP Template
			tmpl, err := template.New("ttp").Parse(ttpTemplate)
			if err != nil {
				return fmt.Errorf("failed to parse TTP template: %v", ttpTemplate)
			}
			params := struct {
				UUID        string
				Platform    platforms.Spec
				Placeholder string
			}{
				UUID:        uuid.NewString(),
				Platform:    platforms.GetCurrentPlatformSpec(),
				Placeholder: "\"@" + "nocommit\"",
			}
			err = tmpl.Execute(f, params)
			if err != nil {
				return fmt.Errorf("failed to execute TTP template: %w", err)
			}
			return nil
		},
	}
	return createTTPCmd
}

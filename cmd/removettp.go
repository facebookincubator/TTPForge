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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildRemoveTTPCommand(cfg *Config) *cobra.Command {
	var unsafe bool

	removeTTPCommand := &cobra.Command{
		Use:               "ttp [repo_name//path/to/ttp]",
		Short:             "Remove a TTP used by TTPForge",
		TraverseChildren:  true,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTTPRef(cfg, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			sourceTTPRef := args[0]

			// Resolve source TTP
			sourceRepo, sourceAbsPath, err := cfg.repoCollection.ResolveTTPRef(sourceTTPRef)
			if err != nil {
				return fmt.Errorf("failed to resolve source TTP reference %v: %w", sourceTTPRef, err)
			}

			// If the repo is not defined in the config file, add it to the collection to resolve references
			if _, err := cfg.repoCollection.GetRepo(sourceRepo.GetName()); err != nil {
				if err := cfg.repoCollection.AddRepo(sourceRepo); err != nil {
					return err
				}
			}

			sourceRef, err := cfg.repoCollection.ConvertAbsPathToAbsRef(sourceRepo, sourceAbsPath)
			if err != nil {
				return fmt.Errorf("failed to convert source path to reference: %w", err)
			}

			fs := afero.NewOsFs()

			if !unsafe {
				// Gather references of files that use this TTP as a dependency
				matches, err := findTTPReferences(cfg.repoCollection, fs, sourceRef)
				if err != nil {
					return err
				}

				if matches != nil {
					for _, match := range matches {
						fmt.Printf("Dependency at %s\n", match)
					}
					fmt.Println("Ending early - review these dependencies before deletion")
					return nil
				}
			}

			fmt.Printf("Removing TTP %s\n", sourceRef)
			if err := fs.Remove(sourceAbsPath); err != nil {
				return fmt.Errorf("failed to remove file: %w", err)
			}

			fmt.Printf("Successfully removed %s\n", sourceAbsPath)

			return nil
		},
	}

	removeTTPCommand.PersistentFlags().BoolVar(&unsafe, "unsafe", false, "Skip dependency check and perform unsafe deletion")

	return removeTTPCommand
}

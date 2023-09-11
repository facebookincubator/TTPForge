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

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func buildRemoveRepoCommand() *cobra.Command {
	var newRepoSpec repos.Spec
	removeRepoCommand := &cobra.Command{
		Use:              "repo",
		Short:            "remove (uninstall) a repository of TTPs used by TTPForge",
		TraverseChildren: true,
		Args:             cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoToRemove := args[0]
			logging.L().Infof("will attempt to delete repo: %v", repoToRemove)
			r, err := Conf.repoCollection.GetRepo(repoToRemove)
			if err != nil {
				return err
			}

			// delete the repo files
			logging.L().Infof("deleting repo files", repoToRemove)
			fp := r.GetFullPath()
			err = afero.NewOsFs().RemoveAll(fp)
			if err != nil {
				return err
			}

			// write the new config with that repo removed
			var newRepoSpecs []repos.Spec
			for _, spec := range Conf.RepoSpecs {
				if spec.Name == repoToRemove {
					break
				}
				newRepoSpecs = append(newRepoSpecs, spec)
			}
			Conf.RepoSpecs = newRepoSpecs
			logging.L().Infof("writing updated configuration...", repoToRemove)
			err = Conf.save()
			if err != nil {
				return fmt.Errorf("failed to save updated configuration: %v", err)
			}
			logging.L().Infof("repo %v deleted successfully", repoToRemove)

			// remove the repo reference from the configuration file
			return nil
		},
	}
	removeRepoCommand.PersistentFlags().StringVar(&newRepoSpec.Name, "name", "", "The name to use for the new repository")
	return removeRepoCommand
}

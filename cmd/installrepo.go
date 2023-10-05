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
	"net/url"
	"path/filepath"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/cobra"
)

func buildInstallRepoCommand(cfg *Config) *cobra.Command {
	var newRepoSpec repos.Spec
	installRepoCommand := &cobra.Command{
		Use:              "repo --name repo_name [repo_url]",
		Short:            "install a new repository of TTPs for use by TTPForge",
		TraverseChildren: true,
		Args:             cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := url.ParseRequestURI(args[0])
			if err != nil {
				return fmt.Errorf("argument must be a valid URL - '%v' is not", args[0])
			}
			newRepoSpec.Git.URL = u.String()
			newRepoSpec.Path = filepath.Join("repos", newRepoSpec.Name)
			cfg.RepoSpecs = append(cfg.RepoSpecs, newRepoSpec)
			_, err = cfg.loadRepoCollection()
			if err != nil {
				return fmt.Errorf("failed to add new repo: %v", err)
			}

			err = cfg.save()
			if err != nil {
				return fmt.Errorf("failed to save updated configuration: %v", err)
			}
			logging.L().Infof("New repository successfully installed!")
			logging.L().Infof("Name: %v", newRepoSpec.Name)
			logging.L().Infof("Path: %v", newRepoSpec.Path)
			logging.L().Infof("List TTPs from your new repository with the command:")
			logging.L().Infof("\tttpforge list ttps --repo %v", newRepoSpec.Name)
			return nil
		},
	}
	installRepoCommand.PersistentFlags().StringVar(&newRepoSpec.Name, "name", "", "The name to use for the new repository")
	_ = installRepoCommand.MarkPersistentFlagRequired("name")
	return installRepoCommand
}

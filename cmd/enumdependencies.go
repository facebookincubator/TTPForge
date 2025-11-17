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
	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

func findTTPReferences(rc repos.RepoCollection, fs afero.Fs, sourceRef string) ([]string, error) {
	var matchList []string

	// Parse source reference
	index := strings.Index(sourceRef, repos.RepoPrefixSep)
	if index == -1 {
		return nil, fmt.Errorf("invalid source reference format: %s", sourceRef)
	}

	sourceRepo := sourceRef[:index]
	sourceScopedRef := sourceRef[index:]
	sourceBareRef := sourceRef[index+2:]

	// Build patterns for finding references in ttp: YAML fields:
	// 1. Full reference: ttp: examples//actions/inline/basic.yaml
	// 2. Scoped reference: ttp: //actions/inline/basic.yaml
	// 3. Bare reference: ttp: actions/inline/basic.yaml (legacy compatibility)
	// These patterns match the entire ttp: field value including any surrounding whitespace
	fullRefPattern := regexp.MustCompile(`(\s*ttp:\s*)` + regexp.QuoteMeta(sourceRef) + `(\s*)`)
	scopedRefPattern := regexp.MustCompile(`(\s*ttp:\s*)` + regexp.QuoteMeta(sourceScopedRef) + `(\s*)`)
	bareRefPattern := regexp.MustCompile(`(\s*ttp:\s*)` + regexp.QuoteMeta(sourceBareRef) + `(\s*)`)

	ttpRefs, err := rc.ListTTPs()
	if err != nil {
		return nil, fmt.Errorf("failed to list TTPs: %w", err)
	}

	for _, ttpRef := range ttpRefs {
		repo, ttpAbsPath, err := rc.ResolveTTPRef(ttpRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve TTP reference %v: %w", ttpRef, err)
		}

		content, err := afero.ReadFile(fs, ttpAbsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read TTP %v: %w", ttpRef, err)
		}

		// Priority 1: Full reference match (examples//actions/inline/basic.yaml)
		if fullRefPattern.Match(content) {
			// add to storage
			matchList = append(matchList, ttpRef)
		} else if repo.GetName() == sourceRepo {
			// Priority 2: Scoped reference match (//actions/inline/basic.yaml)
			if scopedRefPattern.Match(content) {
				matchList = append(matchList, ttpRef)
			} else if bareRefPattern.Match(content) {
				// Priority 3: Bare reference match (actions/inline/basic.yaml) - legacy compatibility
				matchList = append(matchList, ttpRef)
			}
		}
	}

	return matchList, nil
}

func buildEnumDependenciesCommand(cfg *Config) *cobra.Command {
	enumDependenciesCmd := &cobra.Command{
		Use:               "dependencies [repo_name//path/to/ttp]",
		Short:             "Enumerate TTPs that depend on a given TTP",
		Long:              "Use this command to enumerate TTPs that depend on a given TTP",
		Example:           "ttpforge enum dependencies examples//actions/inline/basic.yaml ",
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

			matches, err := findTTPReferences(cfg.repoCollection, fs, sourceRef)
			if err != nil {
				return err
			}

			if matches != nil {
				fmt.Printf("Total dependencies found: %d\n", len(matches))

				if logConfig.Verbose {
					fmt.Println("Dependencies found: ")
					for _, match := range matches {
						fmt.Printf("\t%s\n", match)
					}
				}
			} else {
				fmt.Printf("No dependencies found\n")
			}

			return nil
		},
	}

	return enumDependenciesCmd
}

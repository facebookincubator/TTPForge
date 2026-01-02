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
	"sort"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/parseutils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type AuthorStats struct {
	Count int
	TTPs  []string
}

func buildEnumAuthorsCommand(cfg *Config) *cobra.Command {
	var repo string

	enumAuthorsCmd := &cobra.Command{
		Use:              "authors",
		Short:            "Enumerate authors and their TTP contributions",
		Long:             "Use this command to enumerate all authors across TTPs and see how many TTPs each author has contributed.",
		Example:          "ttpforge enum authors --repo examples --verbose",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			// Fetch all TTPs from the repo
			ttpRefs, err := gatherTTPsFromRepo(cfg, repo)
			if err != nil {
				return err
			}

			// Map to store author -> AuthorStats
			authorMap := make(map[string]*AuthorStats)
			unattributed := &AuthorStats{Count: 0, TTPs: []string{}}

			fs := afero.NewOsFs()
			for _, ttpRef := range ttpRefs {
				_, path, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
				if err != nil {
					if logConfig.Verbose {
						fmt.Printf("Error resolving TTP ref: %v with error: %v\n", ttpRef, err)
					}
					continue
				}

				content, err := afero.ReadFile(fs, path)
				if err != nil {
					if logConfig.Verbose {
						fmt.Printf("Error reading TTP ref: %v on path %v with error: %v\n", ttpRef, path, err)
					}
					continue
				}

				ttp, err := parseutils.ParseTTP(content, path)
				if err != nil {
					if logConfig.Verbose {
						fmt.Printf("Error parsing TTP ref: %v with error: %v\n", ttpRef, err)
					}
					continue
				}

				// Aggregate authors
				hasValidAuthor := false
				for _, author := range ttp.Authors {
					// Normalize author name (trim spaces)
					author = strings.TrimSpace(author)
					if author == "" {
						continue
					}

					hasValidAuthor = true

					if _, exists := authorMap[author]; !exists {
						authorMap[author] = &AuthorStats{
							Count: 0,
							TTPs:  []string{},
						}
					}
					authorMap[author].Count++
					authorMap[author].TTPs = append(authorMap[author].TTPs, ttpRef)
				}

				if !hasValidAuthor {
					unattributed.Count++
					unattributed.TTPs = append(unattributed.TTPs, ttpRef)
				}
			}

			// Sort authors by count (descending) then alphabetically
			type authorEntry struct {
				name  string
				stats *AuthorStats
			}
			authors := make([]authorEntry, 0, len(authorMap))
			for name, stats := range authorMap {
				authors = append(authors, authorEntry{name, stats})
			}

			sort.Slice(authors, func(i, j int) bool {
				if authors[i].stats.Count == authors[j].stats.Count {
					return authors[i].name < authors[j].name
				}
				return authors[i].stats.Count > authors[j].stats.Count
			})

			// Output results
			fmt.Printf("Unique authors: %d\n\n", len(authors))
			fmt.Println("TTP Contribution by Author:")
			fmt.Println("-----------------------------")

			for _, entry := range authors {
				percentage := 0.0
				if len(ttpRefs) > 0 {
					percentage = (float64(entry.stats.Count) / float64(len(ttpRefs))) * 100.0
				}
				fmt.Printf("%-20s %d TTP(s) (%.2f%%)\n", entry.name, entry.stats.Count, percentage)

				if logConfig.Verbose {
					for _, ttpRef := range entry.stats.TTPs {
						fmt.Printf("  - %s\n", ttpRef)
					}
					fmt.Println()
				}
			}

			// Print unattributed at the end
			if unattributed.Count > 0 {
				percentage := 0.0
				if len(ttpRefs) > 0 {
					percentage = (float64(unattributed.Count) / float64(len(ttpRefs))) * 100.0
				}
				fmt.Printf("%-20s %d TTP(s) (%.2f%%)\n", "Unattributed", unattributed.Count, percentage)

				if logConfig.Verbose {
					for _, ttpRef := range unattributed.TTPs {
						fmt.Printf("  - %s\n", ttpRef)
					}
					fmt.Println()
				}
			}

			fmt.Println("-----------------------------")
			fmt.Printf("%-20s %d TTP(s)\n", "Total", len(ttpRefs))

			return nil
		},
	}

	enumAuthorsCmd.PersistentFlags().StringVar(&repo, "repo", "", "Repo to enumerate authors in")
	enumAuthorsCmd.RegisterFlagCompletionFunc("repo", completeRepoName(cfg, 0))

	return enumAuthorsCmd
}

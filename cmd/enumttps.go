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
	"slices"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/parseutils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var allowedPlatforms = []string{"linux", "windows", "darwin", "any"}

func checkPlatformInputValidity(platforms []string) error {
	for _, p := range platforms {
		// Verify that platform is valid
		found := slices.Contains(allowedPlatforms, p)
		if !found {
			fmt.Printf("Pick a valid platform out of these: %v\n", allowedPlatforms)
			return fmt.Errorf("Invalid platform: %s", p)
		}
	}
	return nil
}

func gatherTTPsFromRepo(cfg *Config, repo string) ([]string, error) {
	// Checking repo and accumulating TTPs
	var ttpRefs []string
	var err error
	if repo == "" {
		fmt.Println("Listing all TTPs from all repositories as repo is not specified")
		ttpRefs, err = cfg.repoCollection.ListTTPs()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Printf("Fetching repo: %s\n", repo)
		r, err := cfg.repoCollection.GetRepo(repo)
		if err != nil {
			return nil, fmt.Errorf("Failed to get repo %s: %w", repo, err)
		}
		ttpRefs, err = r.ListTTPs()
		if err != nil {
			return nil, fmt.Errorf("Failed to list TTPs in repo %s: %w", repo, err)
		}
	}
	return ttpRefs, nil
}

func matchMitreData(ttp parseutils.TTP, tactic string, technique string, subTech string) bool {
	// Searching for the MITRE attack ID patterns in the parsed file content
	dataMatching := false
	if tactic != "" {
		for _, ttpTactic := range ttp.Mitre.Tactics {
			if strings.Contains(ttpTactic, tactic) {
				dataMatching = true
				break
			}
		}
		if !dataMatching {
			return false
		}
		dataMatching = false
	}
	if technique != "" {
		for _, ttpTechnique := range ttp.Mitre.Techniques {
			if strings.Contains(ttpTechnique, technique) {
				dataMatching = true
				break
			}
		}
		if !dataMatching {
			return false
		}
		dataMatching = false
	}
	if subTech != "" {
		for _, ttpSubTech := range ttp.Mitre.Subtechniques {
			if strings.Contains(ttpSubTech, subTech) {
				dataMatching = true
				break
			}
		}
		if !dataMatching {
			return false
		}
	}
	return true
}

func filterTTPs(cfg *Config, platforms []string, tactic string, technique string, subTech string, ttpRefs []string, tally map[string]int, totalCount int, verbose bool) (int, []string) {
	updatedTTPRefs := []string{}
	filterPlatform := !slices.Contains(platforms, "any")
	fmt.Printf("Filtering by platforms: %s\n", platforms)

	if !filterPlatform && tactic == "" && technique == "" && subTech == "" {
		fmt.Println("No filters specified, returning all TTPs")
		return len(ttpRefs), ttpRefs
	}
	platformSet := make(map[string]bool)
	for _, inputPlatform := range platforms {
		platformSet[inputPlatform] = true
	}

	fs := afero.NewOsFs()

	// Iterating over all TTPs and filtering them based on platform and attack ID and updating tally
	for _, ttpRef := range ttpRefs {
		_, path, err := cfg.repoCollection.ResolveTTPRef(ttpRef)
		if err != nil {
			if verbose {
				fmt.Printf("Error resolving TTP ref: %v with error: %v\n", ttpRef, err)
			}
			continue
		}
		content, err := afero.ReadFile(fs, path)
		if err != nil {
			if verbose {
				fmt.Printf("Error reading TTP ref: %v on path %v with error: %v", ttpRef, path, err)
			}
			continue
		}

		ttp, err := parseutils.ParseTTP(content, path)
		if err != nil {
			if verbose {
				fmt.Printf("Error parsing TTP ref: %v with error: %v\n", ttpRef, err)
			}
			continue
		}

		if !matchMitreData(ttp, tactic, technique, subTech) {
			continue
		}

		// Platform filtering and updating tally
		if filterPlatform {
			ttpPlatforms := ttp.Requirements.Platforms
			platformMatch := false
			for _, p := range ttpPlatforms {
				if platformSet[p.OS] {
					platformMatch = true
					tally[p.OS]++
				}
			}
			if platformMatch {
				totalCount++
				updatedTTPRefs = append(updatedTTPRefs, ttpRef)
			}
		} else {
			totalCount++
			updatedTTPRefs = append(updatedTTPRefs, ttpRef)
		}
	}
	return totalCount, updatedTTPRefs
}

func buildEnumTTPsCommand(cfg *Config) *cobra.Command {
	var platform string
	var repo string
	var tactic string
	var technique string
	var subTech string
	var verbose bool
	var tally = map[string]int{
		"linux":   0,
		"windows": 0,
		"darwin":  0,
	}
	var totalCount = 0
	enumTTPsCmd := &cobra.Command{
		Use:              "ttps",
		Short:            "Enumerate TTPs basis optional arguments",
		Long:             "Use this command to enumerate TTPs using optional arguments like platform, repo, category, etc.",
		Example:          "ttpforge enum ttps --platform linux,darwin --repo examples --tactic TA0006 --technique T1555 --sub-tech T1555.005",
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			var ttpRefs []string
			var err error

			// Comma separated list of platforms
			platforms := strings.Split(platform, ",")

			// Validating platforms input
			err = checkPlatformInputValidity(platforms)
			if err != nil {
				return err
			}

			// Fetching all TTPs from the repo input by user
			ttpRefs, err = gatherTTPsFromRepo(cfg, repo)
			if err != nil {
				return err
			}

			fmt.Printf("Total %d TTPs found in repo: %s\n", len(ttpRefs), repo)

			// Filtering by platform and Attack ID
			totalCount, ttpRefs = filterTTPs(cfg, platforms, tactic, technique, subTech, ttpRefs, tally, totalCount, verbose)

			// Printing data as per platform
			if !slices.Contains(platforms, "any") {
				for _, p := range platforms {
					fmt.Printf("%s count: %v\n", p, tally[p])
				}
				fmt.Println("Total matching TTPs found: ", totalCount)
			} else {
				fmt.Println("Total TTPs found: ", totalCount)
			}

			if verbose {
				fmt.Println("Verbose output - TTPs found: ")
				// Printing filtered out TTPs
				for _, ttpRef := range ttpRefs {
					fmt.Println(ttpRef)
				}
			}
			return nil
		},
	}
	enumTTPsCmd.PersistentFlags().StringVar(&platform, "platform", "any", "Platform to enumerate TTPs for")
	enumTTPsCmd.PersistentFlags().StringVar(&repo, "repo", "examples", "Repo to enumerate TTPs in")
	enumTTPsCmd.PersistentFlags().StringVar(&tactic, "tactic", "", "Tactic to search for")
	enumTTPsCmd.PersistentFlags().StringVar(&technique, "technique", "", "Technique to search for")
	enumTTPsCmd.PersistentFlags().StringVar(&subTech, "sub-tech", "", "Sub technique to search for")
	enumTTPsCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output that displays all matching TTPs")
	return enumTTPsCmd
}

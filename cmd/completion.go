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
	"strings"

	"github.com/spf13/cobra"
)

// ensureConfigInitialized initializes the config if needed for completion functions.
// Returns true if config is ready, false if completion should fall back to defaults.
func ensureConfigInitialized(cfg *Config, cmd *cobra.Command) bool {
	// Initialize config if needed (completion runs before PersistentPreRunE)
	if cfg != nil && cfg.repoCollection == nil {
		// During completion, flags haven't been parsed yet, so we need to manually
		// check for the --config flag from the command line environment
		if cfg.cfgFile == "" {
			// Check if --config or -c flag is in the completion line
			if configFlag, _ := cmd.Flags().GetString("config"); configFlag != "" {
				cfg.cfgFile = configFlag
			}
		}

		if err := cfg.init(); err != nil {
			return false
		}
	}

	// If config is not initialized or repoCollection is nil, fall back to default
	return cfg != nil && cfg.repoCollection != nil
}

// completeTTPRef provides completion for TTP references in the format "repo//path/to/ttp.yaml"
func completeTTPRef(cfg *Config, maxArgs int) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Stop completing if we've already provided the maximum number of arguments
		if len(args) >= maxArgs {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Ensure config is initialized
		if !ensureConfigInitialized(cfg, cmd) {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Get all available TTP references
		ttpRefs, err := cfg.repoCollection.ListTTPs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Filter TTP references based on what the user has typed so far
		var completions []string
		for _, ref := range ttpRefs {
			if strings.HasPrefix(ref, toComplete) {
				completions = append(completions, ref)
			}
		}

		// NoFileComp directive tells the shell not to fall back to file completion
		// only if we actually have completions to provide
		if len(completions) > 0 {
			return completions, cobra.ShellCompDirectiveNoFileComp
		}

		// If no TTP matches, allow file completion as fallback
		return nil, cobra.ShellCompDirectiveDefault
	}
}

// completeRepoName provides completion for repository names
func completeRepoName(cfg *Config, maxArgs int) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Stop completing if we've already provided the maximum number of arguments (only for positional args)
		if maxArgs > 0 && len(args) >= maxArgs {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Ensure config is initialized
		if !ensureConfigInitialized(cfg, cmd) {
			return nil, cobra.ShellCompDirectiveDefault
		}

		// Get all available repository names
		repoNames := cfg.repoCollection.ListRepos()

		// Filter repository names based on what the user has typed so far
		var completions []string
		for _, repoName := range repoNames {
			if strings.HasPrefix(repoName, toComplete) {
				completions = append(completions, repoName)
			}
		}

		// Return repo names with NoFileComp if we have matches
		if len(completions) > 0 {
			return completions, cobra.ShellCompDirectiveNoFileComp
		}

		// If no repos match, allow default completion behavior
		return nil, cobra.ShellCompDirectiveDefault
	}
}

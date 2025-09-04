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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/repos"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func findAndReplaceTTPReferences(rc repos.RepoCollection, fs afero.Fs, sourceRepo repos.Repo, sourceRef string, destRef string) error {
	// Parse source reference
	index := strings.Index(sourceRef, repos.RepoPrefixSep)
	if index == -1 {
		return fmt.Errorf("invalid source reference format: %s", sourceRef)
	}

	sourceScopedRef := sourceRef[index:]
	sourceBareRef := sourceRef[index+2:]

	fmt.Println(sourceScopedRef, sourceBareRef)

	// Parse destination reference
	index = strings.Index(destRef, repos.RepoPrefixSep)
	if index == -1 {
		return fmt.Errorf("invalid destination reference format: %s", destRef)
	}

	destRepo := destRef[:index]
	destScopedRef := destRef[index:]

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
		return fmt.Errorf("failed to list TTPs: %w", err)
	}

	for _, ttpRef := range ttpRefs {
		repo, ttpAbsPath, err := rc.ResolveTTPRef(ttpRef)
		if err != nil {
			return fmt.Errorf("failed to resolve TTP reference %v: %w", ttpRef, err)
		}

		content, err := afero.ReadFile(fs, ttpAbsPath)
		if err != nil {
			return fmt.Errorf("failed to read TTP %v: %w", ttpRef, err)
		}

		var updated bool
		var newContent []byte
		var replacement string

		// If the destination repo is in the same repo as the ttp, we can use the scoped ref
		// Otherwise, we need to use the full ref
		if repo.GetName() == destRepo {
			replacement = destScopedRef
		} else {
			replacement = destRef
		}

		// Priority 1: Full reference match (examples//actions/inline/basic.yaml)
		if fullRefPattern.Match(content) {
			newContent = fullRefPattern.ReplaceAll(content, []byte("${1}"+replacement+"${2}"))
			updated = true
		} else if repo.GetName() == sourceRepo.GetName() {
			// Priority 2: Scoped reference match (//actions/inline/basic.yaml)
			if scopedRefPattern.Match(content) {
				newContent = scopedRefPattern.ReplaceAll(content, []byte("${1}"+replacement+"${2}"))
				updated = true
			} else if bareRefPattern.Match(content) {
				// Priority 3: Bare reference match (actions/inline/basic.yaml) - legacy compatibility
				newContent = bareRefPattern.ReplaceAll(content, []byte("${1}"+replacement+"${2}"))
				updated = true
			}
		}

		if updated {
			if err := afero.WriteFile(fs, ttpAbsPath, newContent, 0644); err != nil {
				return fmt.Errorf("failed to write updated content to %v: %w", ttpRef, err)
			}
			fmt.Printf("Updated TTP subdependency in %s\n", ttpRef)
		}
	}

	return nil
}

func moveFile(fs afero.Fs, sourceAbsPath, destAbsPath string) error {
	destDir := filepath.Dir(destAbsPath)
	if err := fs.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := fs.Rename(sourceAbsPath, destAbsPath); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", sourceAbsPath, destAbsPath, err)
	}

	return nil
}

func buildMoveCommand(cfg *Config) *cobra.Command {
	moveCmd := &cobra.Command{
		Use:     "move [repo_name//path/to/ttp] [repo_name//path/to/destination]",
		Short:   "Move or rename a TTPForge TTP",
		Long:    "Use this command to move a TTP to a new location, updating all references.",
		Example: "ttpforge move examples//actions/inline/basic.yaml examples//actions/inline/basic-new.yaml",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			sourceTTPRef := args[0]
			destTTPRef := args[1]

			// Resolve source TTP
			sourceRepo, sourceAbsPath, err := cfg.repoCollection.ResolveTTPRef(sourceTTPRef)
			if err != nil {
				return fmt.Errorf("failed to resolve source TTP reference %v: %w", sourceTTPRef, err)
			}

			sourceRef, err := cfg.repoCollection.ConvertAbsPathToAbsRef(sourceRepo, sourceAbsPath)
			if err != nil {
				return fmt.Errorf("failed to convert source path to reference: %w", err)
			}

			fs := afero.NewOsFs()

			// Parse destination TTP reference
			destRepo, destRef, err := cfg.repoCollection.ParseTTPRef(destTTPRef)
			if err != nil {
				return fmt.Errorf("failed to parse destination TTP reference: %w", err)
			}

			// If no repo specified in destination, use source repo
			if destRepo == nil || destRepo.GetName() == "" {
				destRepo = sourceRepo
				// Reconstruct the destination reference with proper repo prefix
				if !strings.Contains(destTTPRef, repos.RepoPrefixSep) {
					destRef = sourceRepo.GetName() + repos.RepoPrefixSep + destTTPRef
				}
			}

			// Validate destination repo exists
			if _, err := cfg.repoCollection.GetRepo(destRepo.GetName()); err != nil {
				return fmt.Errorf("destination repository %s not found: %w", destRepo.GetName(), err)
			}

			// Check if destination already exists
			searchPaths := destRepo.GetTTPSearchPaths()
			if len(searchPaths) == 0 {
				return fmt.Errorf("no TTP search paths configured for repository %s", destRepo.GetName())
			}

			// Extract just the path part from destRef (after //)
			destPath := destRef

			if idx := strings.Index(destRef, repos.RepoPrefixSep); idx != -1 {
				destPath = destRef[idx+2:]
			}

			for _, searchPath := range searchPaths {
				candidatePath := filepath.Join(searchPath, destPath)
				if exists, err := afero.Exists(fs, candidatePath); err != nil {
					return fmt.Errorf("failed to check if destination exists: %w", err)
				} else if exists {
					return fmt.Errorf("destination %s already exists at %s", destTTPRef, candidatePath)
				}
			}

			// Use first search path as destination (as documented in move.md)
			destAbsPath := filepath.Join(searchPaths[0], destPath)

			// Update all TTP references before moving the file
			if err := findAndReplaceTTPReferences(cfg.repoCollection, fs, sourceRepo, sourceRef, destRef); err != nil {
				return fmt.Errorf("failed to update TTP references: %w", err)
			}

			// Move the actual file
			fmt.Printf("Moving TTP from %s to %s\n", sourceRef, destRef)
			if err := moveFile(fs, sourceAbsPath, destAbsPath); err != nil {
				return fmt.Errorf("failed to move file: %w", err)
			}

			fmt.Printf("Successfully moved TTP to: %s\n", destAbsPath)
			return nil
		},
	}

	return moveCmd
}

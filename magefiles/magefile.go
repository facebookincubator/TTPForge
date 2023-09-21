//go:build mage

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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/l50/goutils/v2/dev/lint"
	mageutils "github.com/l50/goutils/v2/dev/mage"
	"github.com/l50/goutils/v2/docs"
	"github.com/l50/goutils/v2/git"
	"github.com/l50/goutils/v2/sys"
	"github.com/spf13/afero"

	// mage utility functions
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

// InstallDeps Installs go dependencies
func InstallDeps() error {
	fmt.Println("Installing dependencies.")

	if err := mageutils.Tidy(); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	if err := lint.InstallGoPCDeps(); err != nil {
		return fmt.Errorf("failed to install pre-commit dependencies: %v", err)
	}

	if err := mageutils.InstallVSCodeModules(); err != nil {
		return fmt.Errorf("failed to install vscode-go modules: %v", err)
	}

	return nil
}

// FindExportedFuncsWithoutTests finds exported functions without tests
func FindExportedFuncsWithoutTests(pkg string) ([]string, error) {
	funcs, err := mageutils.FindExportedFuncsWithoutTests(os.Args[1])
	if err != nil {
		return funcs, err
	}

	for _, funcName := range funcs {
		fmt.Println(funcName)
	}

	return funcs, nil

}

// GeneratePackageDocs generates package documentation
// for packages in the current directory and its subdirectories.
func GeneratePackageDocs() error {
	fs := afero.NewOsFs()

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repo root: %v", err)
	}
	sys.Cd(repoRoot)

	repo := docs.Repo{
		Owner: "facebookincubator",
		Name:  "ttpforge",
	}

	excludedPkgs := []string{"main"}
	template := filepath.Join(repoRoot, "magefiles", "tmpl", "README.md.tmpl")
	if err := docs.CreatePackageDocs(fs, repo, template, excludedPkgs...); err != nil {
		return fmt.Errorf("failed to create package docs: %v", err)
	}

	fmt.Println("Package docs created.")

	return nil
}

// RunPreCommit runs all pre-commit hooks locally
func RunPreCommit() error {
	fmt.Println("Updating pre-commit hooks.")
	if err := lint.UpdatePCHooks(); err != nil {
		return err
	}

	fmt.Println("Clearing the pre-commit cache to ensure we have a fresh start.")
	if err := lint.ClearPCCache(); err != nil {
		return err
	}

	fmt.Println("Running all pre-commit hooks locally.")
	if err := lint.RunPCHooks(); err != nil {
		return err
	}

	return nil
}

// RunTests runs all of the unit tests
func RunTests() error {
	mg.Deps(InstallDeps)

	fmt.Println("Running unit tests.")
	if err := sh.RunV(filepath.Join(".hooks", "run-go-tests.sh"), "all"); err != nil {
		return fmt.Errorf("failed to run unit tests: %v", err)
	}

	return nil
}

func processLines(r io.Reader, language string) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines, codeBlockLines []string
	var inCodeBlock bool

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		inCodeBlock, codeBlockLines = processLine(trimmedLine, line, inCodeBlock, language, codeBlockLines)
		if !inCodeBlock && len(codeBlockLines) > 0 {
			lines = append(lines, codeBlockLines...)
			codeBlockLines = codeBlockLines[:0]
		} else if !inCodeBlock {
			lines = append(lines, line)
		}
	}

	if len(codeBlockLines) > 0 {
		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, "\t\t\t// ```")
		}
		lines = append(lines, codeBlockLines...)
	}

	return lines, scanner.Err()
}

func processLine(trimmedLine, line string, inCodeBlock bool, language string, codeBlockLines []string) (bool, []string) {
	switch {
	case strings.HasPrefix(trimmedLine, "```"+language):
		inCodeBlock = !inCodeBlock
		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
		}
	case inCodeBlock:
		codeBlockLines = append(codeBlockLines, line)
	case strings.Contains(trimmedLine, "```") && inCodeBlock:
		inCodeBlock = false
	}
	return inCodeBlock, codeBlockLines
}

func extractTTPForgeCommand(r io.Reader) ([]string, error) {
	lines, err := processLines(r, "bash")
	if err != nil {
		return nil, err
	}

	var inCodeBlock bool
	var currentCommand string
	var commands []string
	for _, line := range lines {
		line = strings.TrimSpace(line) // Remove leading and trailing spaces

		// Remove the backslashes at the end
		if strings.HasSuffix(line, "\\") {
			line = line[:len(line)-1]
		}

		if strings.Contains(line, "```bash") {
			inCodeBlock = true
			continue
		}

		if inCodeBlock && strings.Contains(line, "```") {
			inCodeBlock = false
			if currentCommand != "" {
				commands = append(commands, strings.TrimSpace(currentCommand))
				currentCommand = ""
			}
			continue
		}

		if inCodeBlock {
			if currentCommand != "" {
				currentCommand += " " + line
			} else {
				currentCommand = line
			}
		}
	}

	return commands, nil
}

func findReadmeFiles(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %v", path, err)
		}

		if !info.IsDir() && info.Name() == "README.md" && strings.Contains(path, "ttps/examples") {
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error reading %s:%v", path, err)
			}

			commands, err := extractTTPForgeCommand(strings.NewReader(string(contents)))
			if err != nil {
				return fmt.Errorf("failed to parse %v: %v", path, err)
			}

			for _, command := range commands {
				if command != "" {
					parts := strings.Fields(command)

					if len(parts) < 2 {
						return fmt.Errorf("unexpected command format: %s", command)
					}

					mainCommand := parts[0]
					action := parts[1]
					ttp := parts[2]
					args := parts[3:]

					cmdArg := []string{action, ttp}

					// Execute the command
					fmt.Printf("Running command extracted from %s: %s %s %s\n\n", info.Name(), mainCommand, cmdArg, strings.Join(args, " "))
					_, err := sys.RunCommand(mainCommand, append(cmdArg, args...)...)
					if err != nil {
						return fmt.Errorf("failed to run command %s %s %s: %v", mainCommand, cmdArg, strings.Join(args, " "), err)
					}
				}
			}

		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path %s: %v", rootDir, err)
	}

	return nil
}

// RunIntegrationTests runs the example TTPs found in ForgeArmory
func RunIntegrationTests() error {
	home, err := sys.GetHomeDir()
	if err != nil {
		return err
	}
	armoryTTPs := filepath.Join(home, ".ttpforge", "repos", "forgearmory", "ttps")
	findReadmeFiles(armoryTTPs)

	return nil
}

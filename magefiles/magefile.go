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
	"runtime"
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

type compileParams struct {
	GOOS            string
	GOARCH          string
	CompileForLocal bool
}

// Compile is used for the compilation of the Go project using goreleaser.
// If the GOOS and GOARCH environment variables are not set, this function
// will default to the current system's OS and architecture.
//
// **Parameters:**
//
// release: If true, this function will compile all supported releases
// for TTPForge. If false, it will compile only the binary for the specified
// OS and architecture (based on environment variables) or for the current
// OS and architecture (if environment variables aren't set).
//
// **Environment Variables:**
//
// GOOS: The target operating system for which the Go project should be
// compiled. Defaults to the system's current OS if not set.
// GOARCH: The target architecture for which the Go project should be
// compiled. Defaults to the system's current architecture if not set.
//
// **Returns:**
//
// error: An error if any issue occurs during the compilation or moving process.
func Compile(release bool) error {
	if !sys.CmdExists("goreleaser") {
		return fmt.Errorf("goreleaser is not installed, please run mage installdeps")
	}

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repo root: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	if cwd != repoRoot {
		if err := os.Chdir(repoRoot); err != nil {
			return fmt.Errorf("failed to change directory to repo root: %v", err)
		}
		defer os.Chdir(cwd)
	}

	doCompile := func(release bool) error {
		var p compileParams
		p.populateFromEnv() // Populate the GOOS and GOARCH parameters

		var args []string

		if release {
			fmt.Println("Compiling all supported releases for TTPForge with goreleaser")
			args = []string{"release", "--snapshot", "--clean", "--skip", "validate"}
		} else {
			fmt.Printf("Compiling the TTPForge binary for %s/%s, please wait.\n", p.GOOS, p.GOARCH)
			args = []string{"build", "--snapshot", "--clean", "--skip", "validate", "--single-target"}
		}

		if err := sh.RunV("goreleaser", args...); err != nil {
			return fmt.Errorf("goreleaser failed to execute: %v", err)
		}
		return nil
	}

	return doCompile(release)
}

func (p *compileParams) populateFromEnv() {
	if p.GOOS == "" {
		p.GOOS = os.Getenv("GOOS")
		if p.GOOS == "" {
			p.GOOS = runtime.GOOS
		}
	}

	if p.GOARCH == "" {
		p.GOARCH = os.Getenv("GOARCH")
		if p.GOARCH == "" {
			p.GOARCH = runtime.GOARCH
		}
	}
}

// InstallDeps installs the TTPForge's Go dependencies necessary for developing
// on the project.
//
// **Returns:**
//
// error: An error if any issue occurs while trying to
// install the dependencies.
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

// FindExportedFuncsWithoutTests identifies exported functions
// within a package that lack corresponding test functions.
//
// **Parameters:**
//
// pkg: A string representing the package name.
//
// **Returns:**
//
// []string: A list of exported functions without tests.
// error: An error if any issue occurs during the identification
// process.
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

// GeneratePackageDocs creates documentation for the various packages in TTPForge.
//
// **Returns:**
//
// error: An error if any issue occurs during documentation
// generation.
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

// RunPreCommit updates, clears, and executes all pre-commit hooks
// locally. The function follows a three-step process:
//  1. Updates the pre-commit hooks using lint.UpdatePCHooks.
//  2. Clears the pre-commit cache with lint.ClearPCCache to ensure
//     a clean environment.
//  3. Executes all pre-commit hooks locally using lint.RunPCHooks.
//
// **Returns:**
//
// error: An error if any issue occurs at any of the three stages
// of the process.
func RunPreCommit() error {
	if !sys.CmdExists("pre-commit") {
		return fmt.Errorf("pre-commit is not installed, please follow the " +
			"instructions in the dev doc: " +
			"https://github.com/facebookincubator/TTPForge/tree/main/docs/dev")
	}

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

// RunTests executes all unit and integration tests.
//
// **Returns:**
//
// error: An error if any issue occurs while running the tests.
func RunTests() error {
	mg.Deps(
		InstallDeps,
		mg.F(Compile, "", ""),
	)

	fmt.Println("Running unit tests.")
	if _, err := sys.RunCommand(filepath.Join(".hooks", "run-go-tests.sh"), "all"); err != nil {
		return fmt.Errorf("failed to run unit tests: %v", err)
	}

	fmt.Println("Running integration tests.")
	if err := RunIntegrationTests(); err != nil {
		return fmt.Errorf("failed to run integration tests: %v", err)
	}

	return nil
}

// RunIntegrationTests executes all integration tests by extracting the commands
// described in README files of TTP examples and then executing them. This
// dynamic testing approach ensures the reliability of TTP examples.
//
// **Returns:**
//
// error: An error if any issue occurs while running the tests.
func RunIntegrationTests() error {
	home, err := sys.GetHomeDir()
	if err != nil {
		return err
	}

	// Capture existing environment variable values to restore them later
	originalBinPath := os.Getenv("BIN_PATH")
	originalDestPath := os.Getenv("DEST_PATH")
	originalPath := os.Getenv("PATH")
	defer func() {
		// Restore original environment variable values after the tests
		os.Setenv("BIN_PATH", originalBinPath)
		os.Setenv("DEST_PATH", originalDestPath)
		os.Setenv("PATH", originalPath)
	}()

	// Set predefined values for the environment variables required by Compile.
	binDirectory, err := filepath.Abs("./ttpforge_tmp") // Convert to absolute path
	if err != nil {
		return fmt.Errorf("failed to get absolute path for temporary directory: %v", err)
	}
	os.Setenv("BIN_PATH", filepath.Join(binDirectory, "ttpforge"))

	// // Build the ttpforge binary first.
	// if err := Compile(""); err != nil {
	// 	return fmt.Errorf("failed to compile ttpforge binary: %v", err)
	// }

	if err := os.Chmod(filepath.Join(binDirectory, "ttpforge"), 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions on ttpforge binary: %v", err)
	}

	// Adjust the PATH to prioritize the freshly built binary.
	newPath := binDirectory + string(os.PathListSeparator) + originalPath
	os.Setenv("PATH", newPath)

	defer os.RemoveAll(binDirectory) // Remove the temporary directory after tests

	armoryTTPs := filepath.Join(home, ".ttpforge", "repos", "forgearmory", "ttps")

	// Parse README files to extract and run example commands, ensuring the
	// validity of our examples.
	return findReadmeFiles(armoryTTPs)
}

// processLines parses an io.Reader, identifying and marking code blocks
// found in a TTP README.
func processLines(r io.Reader, language string) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines, codeBlockLines []string
	var inCodeBlock bool

	for scanner.Scan() {
		line := scanner.Text()

		inCodeBlock, codeBlockLines = handleLineInCodeBlock(strings.TrimSpace(line), line, inCodeBlock, language, codeBlockLines)

		if !inCodeBlock {
			lines = append(lines, codeBlockLines...)
			codeBlockLines = codeBlockLines[:0]
			if !strings.HasPrefix(line, "```") {
				lines = append(lines, line)
			}
		}
	}

	if inCodeBlock {
		codeBlockLines = append(codeBlockLines, "\t\t\t// ```")
		lines = append(lines, codeBlockLines...)
	}

	return lines, scanner.Err()
}

// handleLineInCodeBlock categorizes and handles each line based on its
// content and relation to code blocks found in a TTP README.
func handleLineInCodeBlock(trimmedLine, line string, inCodeBlock bool, language string, codeBlockLines []string) (bool, []string) {
	switch {
	case strings.HasPrefix(trimmedLine, "```"+language):
		if !inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
		}
		return !inCodeBlock, codeBlockLines
	case inCodeBlock:
		codeBlockLines = append(codeBlockLines, line)
	case strings.Contains(trimmedLine, "```"):
		inCodeBlock = false
	}
	return inCodeBlock, codeBlockLines
}

// extractTTPForgeCommand extracts the TTPForge run commands from the provided
// reader (parsed README content). This approach automates the testing of
// examples by leveraging the commands documented in READMEs.
func extractTTPForgeCommand(r io.Reader) ([]string, error) {
	lines, err := processLines(r, "bash")
	if err != nil {
		return nil, err
	}

	var inCodeBlock bool
	var currentCommand string
	var commands []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Remove the backslashes at the end
		trimmedLine = strings.TrimSuffix(trimmedLine, "\\")

		switch {
		case strings.Contains(trimmedLine, "```bash"):
			inCodeBlock = true
			currentCommand = ""
		case inCodeBlock && strings.Contains(trimmedLine, "```"):
			inCodeBlock = false
			if currentCommand != "" {
				commands = append(commands, strings.TrimSpace(currentCommand))
			}
		case inCodeBlock:
			if currentCommand != "" {
				currentCommand += " " + trimmedLine
			} else {
				currentCommand = trimmedLine
			}
		}
	}

	return commands, nil
}

// findReadmeFiles looks for README.md files in the specified directory.
// The READMEs are expected to contain TTPForge commands that serve as
// user-facing instructions for the examples. By parsing these READMEs, we can
// automatically test and validate these instructions.
func findReadmeFiles(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %v", path, err)
		}

		if !info.IsDir() && info.Name() == "README.md" && strings.Contains(path, "ttps/examples") {
			return processReadme(path, info)
		}
		return nil
	})
}

// processReadme reads the content of a given README file, extracts the
// TTPForge commands, and runs them. This acts as a verification step to
// ensure the examples work as described in the README.
func processReadme(path string, info os.FileInfo) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading %s:%v", path, err)
	}

	commands, err := extractTTPForgeCommand(strings.NewReader(string(contents)))
	if err != nil {
		return fmt.Errorf("failed to parse %v: %v", path, err)
	}

	for _, command := range commands {
		if err := runExtractedCommand(command, info); err != nil {
			return err
		}
	}
	return nil
}

// runExtractedCommand executes the input TTPForge command, acting as a
// dynamic validation step.
func runExtractedCommand(command string, info os.FileInfo) error {
	if command == "" {
		return nil
	}

	parts := strings.Fields(command)
	if len(parts) < 3 {
		return fmt.Errorf("unexpected command format: %s", command)
	}

	mainCommand, action, ttp := parts[0], parts[1], parts[2]
	args := parts[3:]

	fmt.Printf("Running command extracted from %s: %s %s %s\n\n", info.Name(), mainCommand, action, strings.Join(args, " "))

	if _, err := sys.RunCommand(mainCommand, append([]string{action, ttp}, args...)...); err != nil {
		return fmt.Errorf("failed to run command %s %s %s: %v", mainCommand, action, strings.Join(args, " "), err)
	}
	return nil
}

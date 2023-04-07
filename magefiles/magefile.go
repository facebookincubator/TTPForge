//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	goutils "github.com/l50/goutils"

	// mage utility functions
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

// InstallDeps Installs go dependencies
func InstallDeps() error {
	fmt.Println(color.YellowString("Installing dependencies."))

	if err := goutils.Tidy(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install dependencies: %v", err))
	}

	if err := goutils.InstallGoPCDeps(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install pre-commit dependencies: %v", err))
	}

	if err := goutils.InstallVSCodeModules(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install vscode-go modules: %v", err))
	}

	return nil
}

// InstallPreCommitHooks Installs pre-commit hooks locally
func InstallPreCommitHooks() error {
	mg.Deps(InstallDeps)

	fmt.Println(color.YellowString("Installing pre-commit hooks."))
	if err := goutils.InstallPCHooks(); err != nil {
		return err
	}

	return nil
}

// RunPreCommit runs all pre-commit hooks locally
func RunPreCommit() error {
	mg.Deps(InstallDeps)

	fmt.Println(color.YellowString("Updating pre-commit hooks."))
	if err := goutils.UpdatePCHooks(); err != nil {
		return err
	}

	fmt.Println(color.YellowString(
		"Clearing the pre-commit cache to ensure we have a fresh start."))
	if err := goutils.ClearPCCache(); err != nil {
		return err
	}

	fmt.Println(color.YellowString("Running all pre-commit hooks locally."))
	if err := goutils.RunPCHooks(); err != nil {
		return err
	}

	return nil
}

// RunTests runs all of the unit tests
func RunTests() error {
	mg.Deps(InstallDeps)

	fmt.Println(color.YellowString("Running unit tests."))
	if err := sh.RunV(filepath.Join(".hooks", "go-unit-tests.sh"), "all"); err != nil {
		return fmt.Errorf(color.RedString("failed to run unit tests: %v", err))
	}

	return nil
}

// UpdateMirror updates pkg.go.dev with the release associated with the input tag
func UpdateMirror(tag string) error {
	var err error
	fmt.Println(color.YellowString("Updating pkg.go.dev with the new tag %s.", tag))

	err = sh.RunV("curl", "--silent", fmt.Sprintf(
		"https://sum.golang.org/lookup/github.com/facebookincubator/ttpforge@%s",
		tag))
	if err != nil {
		return fmt.Errorf(color.RedString("failed to update proxy.golang.org: %w", err))
	}

	err = sh.RunV("curl", "--silent", fmt.Sprintf(
		"https://proxy.golang.org/github.com/facebookincubator/ttpforge/@v/%s.info",
		tag))
	if err != nil {
		return fmt.Errorf(color.RedString("failed to update pkg.go.dev: %w", err))
	}

	return nil
}

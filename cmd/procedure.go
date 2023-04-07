//go:build embed

package cmd

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/facebookincubator/ttpforge/pkg/blocks"
	"github.com/facebookincubator/ttpforge/pkg/logging"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	generatedDir = ".generated_ttps"
	// EmbeddedTTPs embeds generated TTPs into a compiled TTP Forge.
	EmbeddedTTPs *embed.FS
	runProcCmd   = &cobra.Command{
		Use:              "proc",
		Short:            "Run the embedded procedure.",
		TraverseChildren: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logging.Logger = Logger
		},
	}

	list bool
)

func init() {
	rootCmd.AddCommand(runProcCmd)
	runProcCmd.PersistentFlags().BoolVar(&list, "list", false, "list all subcommands recursively")
}

// addDirCommand adds a command to list the subdirectories and TTP actions of the
// current directory in the YAML files. If the current directory is the root directory,
// the TTP actions are added to the parent command using the addCommands function.
func addDirCommand(path string) *cobra.Command {
	Logger.Sugar().Debugw("Adding directory subcommand", "dir", path)
	newCmd := &cobra.Command{
		Use:              path, // keep it simple, just use the path as the subcommand
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			if list {
				fs.WalkDir(*EmbeddedTTPs, ".", func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						Logger.Error("bad workdir", zap.Error(err))
						return err
					}

					if !d.IsDir() && strings.Contains(p, path) && filepath.Ext(p) == ".yaml" {
						// Remove prefix.
						splitPaths := strings.SplitN(p, path, 2)
						// Remove trailing suffix.
						pathFromBase := strings.Split(splitPaths[1], ".")[0]
						commandPath := strings.Split(pathFromBase, "/")

						Logger.Sugar().Infow("subcommands", "path", strings.Join(commandPath, " "))
					}
					return nil

				})
			}
		},
	}
	return newCmd
}

func addCommands(path string, ttp blocks.TTP) *cobra.Command {
	Logger.Sugar().Debugw("Blocks loaded", "actions", ttp)
	Logger.Sugar().Debugw("Blocks loaded", "path", filepath.Base(path))
	newCmd := &cobra.Command{
		Use:              filepath.Base(path),
		Short:            ttp.Description,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			ttp.SetEmbedHome(generatedDir)
			if err := ttp.RunSteps(); err != nil {
				Logger.Sugar().Errorw("[!] TTP failed with error", "error", err)
			}
			if failed := ttp.Failed(); failed != nil {
				Logger.Sugar().Errorw("[!] TTP failed", "steps", failed)
				return
			}
			Logger.Sugar().Infow("[+] Successfully executed ttps", "name", ttp.Name)
			return
		},
	}

	Logger.Sugar().Debugw("Successfully added command", "ttp", ttp.Name)
	Logger.Sugar().Debugw("Command actions", "steps", ttp.Steps)
	return newCmd
}

// InitYAML initializes YAML files and generates corresponding Cobra commands based on the embedded
// file system. The function walks through the directories and files in the embedded file system,
// creates Cobra commands for each directory and TTP action, and adds them to the parent commands
// accordingly.
//
// Example usage:
//
// embeddings, _ := embed.NewFS()
// InitYAML(&embeddings)
//
// Parameters:
//
// embeddings: A pointer to an embed.FS object, representing the embedded file system containing
// the YAML files.
//
// Side Effects:
//
// The function updates global variables and adds generated Cobra commands to the appropriate
// parent commands based on the structure of the embedded file system.
func InitYAML(embeddings *embed.FS) {
	EmbeddedTTPs = embeddings
	Logger.Debug("Initializing yaml files.")

	tmpCache := make(map[string]*cobra.Command)

	// Skip the true root of "."
	tmpCache[generatedDir] = runProcCmd
	fs.WalkDir(*embeddings, generatedDir, func(path string, d fs.DirEntry, err error) error {
		// Skip the base directory.
		if path == generatedDir {
			return nil
		}
		if err != nil {
			Logger.Error("bad workdir", zap.Error(err))
			return err
		}

		parentCommand := tmpCache[filepath.Dir(path)]
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			Logger.Sugar().Debugw("filewalk", "file found", d.Name())
			actions, err := blocks.LoadTTP(path)
			if err != nil {
				Logger.Sugar().Errorw("failed loading yaml file ttps", "path", path, "err", err)
				return err
			}

			newCmd := addCommands(d.Name(), actions)
			parentCommand.AddCommand(newCmd)

		} else if d.IsDir() {
			Logger.Sugar().Debugw("filewalk", "dir found", d.Name())
			newCmd := addDirCommand(d.Name())
			// Swap in place.
			parentCommand.AddCommand(newCmd)
			if _, ok := tmpCache[path]; !ok {
				tmpCache[path] = newCmd
			}
		}
		return nil

	})
}

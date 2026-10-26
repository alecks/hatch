package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

func addConfigCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(contextDefinition).(*Definition)
	if definition == nil {
		panic("definition missing, this should not happen")
	}
	unlock, err := definition.Lock()
	if err != nil {
		return fmt.Errorf("acquire definition file lock (%s): %w", definition.filePath, err)
	}
	defer unlock()
	if err := definition.Read(); err != nil {
		return fmt.Errorf("read definition file (%s): %w", definition.filePath, err)
	}

	recursive := cmd.Bool(flagRecursive)
	resourceName := cmd.StringArg(argResourceName)
	definitionDir := filepath.Dir(definition.filePath)
	configDir := filepath.Join(definitionDir, hatchConfigDir)
	destPath := cmd.StringArg(argDestPath)
	sourcePath := cmd.StringArg(argSourcePath)
	// storePath is the definitionDir/configDir/the/dest/path, to allow for stow compatibility
	storePath := filepath.Join(configDir, strings.TrimPrefix(destPath, "/"))

	if !filepath.IsAbs(destPath) {
		return fmt.Errorf("dest-path must be an absolute path (%s)", destPath)
	}

	if _, ok := definition.Resources[resourceName]; ok {
		return fmt.Errorf("a resource with the same name already exists in this definition")
	}
	if _, err := os.Stat(storePath); err == nil {
		return fmt.Errorf("a config with the same path already exists in this definition")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", storePath, err)
	}

	sourceIsDir := false
	if recursive {
		if info, err := os.Stat(sourcePath); err == nil {
			if info.IsDir() {
				sourceIsDir = true
				// NOTE: DirFS docs says it follows symlinks, which could cause problems: consider using copyDir if we end up needing more control
				if err := os.CopyFS(storePath, os.DirFS(sourcePath)); err != nil {
					return fmt.Errorf("copy source -> store (%s -> %s): %w", storePath, sourcePath, err)
				}
				slog.Debug("copied source dir to store dir (recursive = true)", "sourcePath", sourcePath, "storePath", storePath)
			} else {
				fmt.Fprintf(os.Stderr, "warning: -r passed, but source-path is not a directory")
			}
		} else {
			return fmt.Errorf("stat source path at %s (recursive = true): %w", sourcePath, err)
		}
	}
	// we may have had recursive flag set, but source-path was a file
	if !sourceIsDir {
		// not recursive: assume we're just dealing with one file
		// don't need to handle dir because os.Open will fail
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("open source file (%s): %w", sourcePath, err)
		}
		defer sourceFile.Close()

		storeDir := filepath.Dir(storePath)
		if err := os.MkdirAll(storeDir, hatchDirPerms); err != nil {
			return fmt.Errorf("create store directory at %s (recursive = false): %w", storeDir, err)
		}
		slog.Debug("created store directory (recursive = false)", "storeDir", storeDir)

		storeFile, err := os.Create(storePath)
		if err != nil {
			return fmt.Errorf("create store file (%s): %w", storePath, err)
		}
		defer storeFile.Close()
		slog.Debug("created store file for added config (recursive = false)", "storePath", storePath)

		_, err = io.Copy(storeFile, sourceFile)
		if err != nil {
			slog.Debug("cleanup: removing store file due to failed copy", "storePath", storePath)
			os.Remove(storePath)
			return fmt.Errorf("copy source file to store file (%s -> %s): %w", sourcePath, storePath, err)
		}
	}

	definition.Resources[resourceName] = Resource{
		Config: &Config{
			DestPath: destPath,
			IsDir:    sourceIsDir,
		},
		DependsOn: cmd.StringSlice(flagDependsOn),
	}
	if err := definition.Write(); err != nil {
		slog.Debug("cleanup: removing store file due to failed definition write", "storePath", storePath)
		os.Remove(storePath)
		return fmt.Errorf("write definition file atomically (%s): %w", definition.filePath, err)
	}
	slog.Debug("wrote definition for added config resource", "definitionPath", definition.filePath)

	return nil
}

func removeConfigCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(contextDefinition).(*Definition)
	if definition == nil {
		panic("definition was nil; this should not happen")
	}
	unlock, err := definition.Lock()
	if err != nil {
		return fmt.Errorf("acquire definition file lock (%s): %w", definition.filePath, err)
	}
	defer unlock()
	if err := definition.Read(); err != nil {
		return fmt.Errorf("read definition file (%s): %w", definition.filePath, err)
	}

	resourceName := cmd.StringArg(argResourceName)
	resource, ok := definition.Resources[resourceName]
	if !ok {
		return fmt.Errorf("resource '%s' doesn't exist", resourceName)
	}
	conf := resource.Config
	if conf == nil {
		return fmt.Errorf("this resource exists, but is not a config")
	}

	// TODO: add DependsOn checks

	if !conf.IsDir {
		configDir := filepath.Join(filepath.Dir(definition.filePath), hatchConfigDir)
		storePath := filepath.Join(configDir, conf.DestPath)

		var confirmed bool
		err := huh.NewConfirm().
			Title("Remove this configuration?").
			Description(fmt.Sprintf("Resource name: %s\nConfig file: %s\n\nOnly the internal store will be affected.", resourceName, storePath)).
			Affirmative("Yes").
			Negative("No").
			Value(&confirmed).
			Run()
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
		if err := os.Remove(storePath); err != nil {
			return fmt.Errorf("delete stored config (%s): %w", storePath, err)
		}

		// TODO: delete parent dirs if they're empty
	} else {
		return errors.New("not implemented")
	}

	delete(definition.Resources, resourceName)
	return definition.Write()
}

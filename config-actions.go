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
		return err
	}
	defer unlock()
	if err := definition.Read(); err != nil {
		return err
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
		return errorResourceExists(resourceName)
	}
	if _, err := os.Stat(storePath); err == nil {
		return fmt.Errorf("a config with the same path already exists in this definition")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
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
				fmt.Fprintf(os.Stderr, "warning: -r passed, but source-path is not a directory\n")
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
			return err
		}
		slog.Debug("created store directory (recursive = false)", "storeDir", storeDir)

		storeFile, err := os.Create(storePath)
		if err != nil {
			return err
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
		fmt.Fprintf(os.Stderr, "cleanup: removing store file (%s) due to failed definition write", storePath)
		os.Remove(storePath)
		return err
	}

	fmt.Fprintf(os.Stderr, "config resource %q added to definition %q\n", resourceName, definition.Name)
	fmt.Println(resourceName)
	return nil
}

func removeConfigCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(contextDefinition).(*Definition)
	if definition == nil {
		panic("definition was nil; this should not happen")
	}
	unlock, err := definition.Lock()
	if err != nil {
		return err
	}
	defer unlock()
	if err := definition.Read(); err != nil {
		return err
	}

	resourceName := cmd.StringArg(argResourceName)
	resource, ok := definition.Resources[resourceName]
	if !ok {
		return errorResourceNotFound(resourceName)
	}
	conf := resource.Config
	if conf == nil {
		return errorResourceWrongType(resourceName, "config")
	}

	// TODO: add DependsOn checks

	configDir := filepath.Join(filepath.Dir(definition.filePath), hatchConfigDir)
	storePath := filepath.Join(configDir, conf.DestPath)
	storePathTmp := storePath + ".tmp"
	forceYes := cmd.Bool(flagForceYes)

	if !forceYes {
		if !conf.IsDir {
			var confirmed bool
			err := huh.NewConfirm().
				Title("Remove this configuration?").
				Description(fmt.Sprintf("Resource name: %s\nConfig file: %s\n\nOnly the internal store will be affected.", resourceName, storePath)).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed).
				WithTheme(huh.ThemeBase16()).
				Run()
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
			// TODO: delete parent dirs if they're empty
		} else {
			var confirmed bool
			err := huh.NewConfirm().
				Title("Remove this configuration? (WARNING: config is a directory)").
				Description(fmt.Sprintf(
					"Resource name: %s\nConfig dir: %s\n\n"+
						"Only the store will be affected.\n"+
						"Any files manually moved to the store after this resource was created will be removed.",
					resourceName, storePath)).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed).
				WithTheme(huh.ThemeBase16()).
				Run()
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}
		}
	}

	if err := os.Rename(storePath, storePathTmp); err != nil {
		return fmt.Errorf("rename stored config (%s): %w", storePath, err)
	}

	delete(definition.Resources, resourceName)
	if err := definition.Write(); err != nil {
		if err := os.Rename(storePathTmp, storePath); err != nil {
			return fmt.Errorf("FAILED to revert config marked for deletion. Please manually rename %s to %s. err: %w", storePathTmp, storePath, err)
		}
		return err
	}

	fmt.Fprintf(os.Stderr, "deleted config at %s; removed resource %q from hatchfile\n", storePath, resourceName)
	fmt.Println(resourceName)

	slog.Debug("deleting temp (renamed config)", "storePathTmp", storePathTmp, "isDir", conf.IsDir)
	if conf.IsDir {
		return os.RemoveAll(storePathTmp)
	} else {
		return os.Remove(storePathTmp)
	}
}

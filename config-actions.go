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

	"github.com/urfave/cli/v3"
)

func addConfigCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(definitionKey).(*Definition)
	if definition == nil {
		panic("definition missing, this should not happen")
	}
	if err := definition.Read(); err != nil {
		return fmt.Errorf("read in definition file (%s): %w", definition.filePath, err)
	}

	resourceName := cmd.String(flagResourceName)
	definitionDir := filepath.Dir(definition.filePath)
	configDir := filepath.Join(definitionDir, "config")
	destPath := cmd.StringArg(argDestPath)
	sourcePath := cmd.StringArg(argSourcePath)
	// TODO: ensure storepath doesn't escape configDir
	storePath := filepath.Join(configDir, strings.TrimPrefix(destPath, "/"))
	storeDir := filepath.Dir(storePath)

	if !filepath.IsAbs(destPath) {
		return fmt.Errorf("dest-path must be an absolute path (%s)", destPath)
	}

	if existing, ok := definition.Configs[resourceName]; ok {
		return fmt.Errorf("a config with the same resource name already exists in this definition (its destPath: %s)", existing.DestPath)
	}
	if _, err := os.Stat(storePath); err == nil {
		return fmt.Errorf("a config with the same path already exists in this project")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", storePath, err)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source file (%s): %w", sourcePath, err)
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(storeDir, hatchDirPerms); err != nil {
		return fmt.Errorf("create store directory (%s): %w", storeDir, err)
	}
	slog.Debug("created store directory", "storeDir", storeDir)

	storeFile, err := os.Create(storePath)
	if err != nil {
		return fmt.Errorf("create store file (%s): %w", storePath, err)
	}
	defer storeFile.Close()
	slog.Debug("created store file for added config", "storePath", storePath)

	_, err = io.Copy(storeFile, sourceFile)
	if err != nil {
		slog.Debug("cleanup: removing store file due to failed copy", "storePath", storePath)
		os.Remove(storePath)
		return fmt.Errorf("copy source file to store file (%s -> %s): %w", sourcePath, storePath, err)
	}

	definition.Configs[resourceName] = Config{
		DestPath:  destPath,
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

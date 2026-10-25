package main

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"

	"github.com/urfave/cli/v3"
)

func addConfigCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(definitionKey)
	if definition == nil {
		return fmt.Errorf("missing definition (this should not happen)")
	}
	definitionDir := cmd.String(flagDir)

	configDir := filepath.Join(definitionDir, "config")
	_ = Config{
		// TODO: use incremental instead
		Name:     filepath.Base(definitionDir) + "-" + strconv.Itoa(rand.IntN(1000)),
		DestPath: Path(cmd.StringArg(argDestPath)),
	}

	sourcePath := cmd.StringArg(argSourcePath)
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file (%s), err: %v", sourcePath, err)
	}
	defer sourceFile.Close()

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config dir at %s (used for storing added configurations), err: %v", configDir, err)
	}

	storePath := filepath.Join(configDir, cmd.StringArg("dest-path"))
	storeFile, err := os.Open(storePath)
	if err != nil {
		return fmt.Errorf("failed to open store file (%s), err: %v", storePath, err)
	}
	defer storeFile.Close()

	_, err = io.Copy(storeFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy source file to store file: %s -> %s, error: %v", sourcePath, storePath, err)
	}

	return nil
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v3"
)

type contextKey string

const (
	definitionKey contextKey = "definition"

	flagDir       string = "dir"
	flagVerbose   string = "verbose"
	argSourcePath string = "source"
	argDestPath   string = "dest"

	// 0755: owner read/write, others read (+execute)
	hatchDirPerms  os.FileMode = 0755
	hatchConfigDir string      = "config"
)

func main() {
	cwdPath, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get cwd: %v", err)
		os.Exit(1)
	}

	cmd := &cli.Command{
		Name:    "hatch",
		Usage:   "a tool for documenting and reusing configurations on small servers and workstations",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        flagDir,
				Aliases:     []string{"d"},
				Usage:       "path to the hatch project dir",
				Value:       path.Join(cwdPath, "hatch"),
				DefaultText: "./hatch/",
			},
			&cli.BoolFlag{
				Name:        flagVerbose,
				Aliases:     []string{"v"},
				Usage:       "set log level to debug",
				Value:       false,
				DefaultText: "false",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Bool(flagVerbose) {
				slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
			}

			dirPath := cmd.String(flagDir)
			filePath := filepath.Join(dirPath, "hatchfile.toml")
			slog.Debug("attempting to open hatchfile", "path", dirPath)

			file, err := os.Open(filePath)
			if errors.Is(err, os.ErrNotExist) {
				slog.Debug("hatchfile does not exist, creating", "path", filePath)

				if err := os.MkdirAll(dirPath, hatchDirPerms); err != nil {
					return ctx, fmt.Errorf("failed to create hatch project directory (%s): %v", filepath.Dir(dirPath), err)
				}
				file, err = os.Create(filePath)
				if err != nil {
					return ctx, fmt.Errorf("failed to create hatchfile (%s): %v", filePath, err)
				}
				defer file.Close()

				toml.NewEncoder(file).Encode(&Definition{
					Name: filepath.Base(dirPath),
				})
			}
			defer file.Close()

			definition := &Definition{}
			if err := toml.NewDecoder(file).Decode(definition); err != nil {
				return ctx, fmt.Errorf("failed to decode hatchfile (%s): %v", dirPath, err)
			}

			slog.Debug("loaded hatchfile", "path", dirPath, "definition", definition)
			return context.WithValue(ctx, definitionKey, definition), nil
		},
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "apply a hatch definition to this machine",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return errors.New("not implemented")
				},
			},
			{
				Name:  "down",
				Usage: "remove all hatch-managed resources from this machine",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return errors.New("not implemented")
				},
			},
			{
				Name:  "status",
				Usage: "shows the current state of hatch-managed resources on this machine",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return errors.New("not implemented")
				},
			},
			{
				Name:  "config",
				Usage: "manage hatch-defined configuration files",
				Commands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add a configuration file to the hatch definition",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      "source-path",
								UsageText: "the local configuration file to add to the hatch definition",
							},
							&cli.StringArg{
								Name:      "dest-path",
								UsageText: "the path where the configuration file should be placed on the target machine",
							},
						},
						Action: addConfigCmd,
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "remove a configuration file from the hatch definition",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return errors.New("not implemented")
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "failed: %v\n", err)
		os.Exit(1)
	}
}

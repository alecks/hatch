package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

type contextKey string

const (
	contextDefinition contextKey = "definition"

	flagDir       string = "dir"
	flagVerbose   string = "verbose"
	flagDependsOn string = "depends"
	flagRecursive string = "recursive"

	argResourceName string = "resource"
	argSourcePath   string = "source"
	argDestPath     string = "dest"

	// 0755: owner read/write/execute, others read/execute)
	hatchDirPerms os.FileMode = 0755
	// 0644: owner read/write, others read-only
	hatchFilePerms os.FileMode = 0644
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
			definition := &Definition{
				Name:     filepath.Base(dirPath),
				filePath: filePath,
			}

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

				definition.Write()
			}
			defer file.Close()
			slog.Debug("opened hatchfile", "path", dirPath)

			return context.WithValue(ctx, contextDefinition, definition), nil
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
						Usage: "add a configuration resource to the hatch definition",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  flagDependsOn,
								Usage: "names of other resources that this resource depends on",
							},
							&cli.BoolFlag{
								Name:        flagRecursive,
								Aliases:     []string{"r"},
								Usage:       "allow adding a directory as a configuration resource",
								Value:       false,
								DefaultText: "false",
							},
						},
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      argResourceName,
								UsageText: "resource-name",
								Required:  true,
							},
							&cli.StringArg{
								Name:      argSourcePath,
								UsageText: "source-path",
								Required:  true,
							},
							&cli.StringArg{
								Name:      argDestPath,
								UsageText: "dest-path",
								Required:  true,
							},
						},
						Action: addConfigCmd,
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "remove a configuration resource from the hatch definition",
						Arguments: []cli.Argument{
							&cli.StringArg{
								Name:      argResourceName,
								UsageText: "resource-name",
								Required:  true,
							},
						},
						Action: removeConfigCmd,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

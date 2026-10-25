package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"
	"github.com/urfave/cli/v3"
)

type contextKey string

const definitionKey contextKey = "definition"

func main() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get user home dir: %v\n", err)
		os.Exit(1)
	}
	hatchfileFlag := &cli.StringFlag{
		Name:        "hatchfile",
		Aliases:     []string{"f"},
		Usage:       "path to the hatchfile",
		Value:       path.Join(dirname, ".hatch", "hatchfile.toml"),
		DefaultText: "~/.hatch/hatchfile.toml",
	}

	cmd := &cli.Command{
		Name:        "hatch",
		Usage:       "a tool for documenting and reusing configurations on small servers and workstations",
		Version:     "0.1.0",
		Description: "hatch is a tool for documenting and reusing configurations on small servers and workstations. It uses the HCL configuration language to define configurations, systemd services, Docker containers and other resources.",
		Flags:       []cli.Flag{hatchfileFlag},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			path := cmd.String("hatchfile")
			slog.Debug("attempting to open hatchfile", "path", path)

			file, err := os.Open(path)
			if errors.Is(err, os.ErrNotExist) {
				slog.Debug("hatchfile does not exist, creating one", "path", path)
				file, err = os.Create(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to create hatchfile (%s): %v\n", path, err)
					os.Exit(1)
				}
			}

			var definition *Definition
			if err := toml.NewDecoder(file).Decode(definition); err != nil {
				fmt.Fprintf(os.Stderr, "failed to decode hatchfile (%s): %v\n", path, err)
				os.Exit(1)
			}

			slog.Debug("loaded hatchfile", "path", path, "definition", definition)
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
						Usage: "add a configuration file to the hatch defnition",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return errors.New("not implemented")
						},
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

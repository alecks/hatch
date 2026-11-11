package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

func addCommandCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(contextDefinition).(*Definition)
	if definition == nil {
		panic("definition was nil (this should not happen)")
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
	runCommand := cmd.StringArg(argRunCommand)

	if _, ok := definition.Resources[resourceName]; ok {
		return errorResourceExists(resourceName)
	}

	definition.Resources[resourceName] = Resource{
		DependsOn: cmd.StringSlice(flagDependsOn),
		Command: &Command{
			Run: runCommand,
		},
	}

	if err := definition.Write(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "command resource %q added to definition %q\n", resourceName, definition.Name)
	fmt.Println(resourceName)
	return nil
}

func removeCommandCmd(ctx context.Context, cmd *cli.Command) error {
	definition := ctx.Value(contextDefinition).(*Definition)
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
	if resource.Command == nil {
		return errorResourceWrongType(resourceName, "command")
	}

	if !cmd.Bool(flagForceYes) {
		var confirmed bool
		err = huh.NewConfirm().
			Title("Remove this command?").
			Description(fmt.Sprintf("Resource name: %s\nCommand: %s\n\nOnly the hatchfile will be affected.", resourceName, resource.Run)).
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

	delete(definition.Resources, resourceName)
	if err := definition.Write(); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "deleted command resource %q from definition %q\n", resourceName, definition.Name)
	fmt.Println(resourceName)
	return nil
}

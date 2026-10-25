package main

import "errors"

type Path string

type Definition struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description,omitempty"`
	Configs     []Config `toml:"config,omitempty"`
}

type Resource struct {
	Name      string   `toml:"id"`
	DependsOn []string `toml:"depends_on,omitempty"`
}

type Config struct {
	Resource
	DestPath Path `toml:"dest"`
}

func (d *Definition) Write() error {
	return errors.New("not implemented")
}

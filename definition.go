package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Definition struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description,omitempty"`
	Configs     map[string]Config `toml:"config,omitempty"`
	filePath    string            `toml:"-"`
}

type Resource struct {
	DependsOn []string `toml:"depends_on,omitempty"`
}

type Config struct {
	Resource
	DestPath string `toml:"dest"`
}

func (d *Definition) Write() error {
	definitionToml, err := toml.Marshal(d)
	if err != nil {
		return err
	}
	return writeFileAtomic(d.filePath, definitionToml, hatchFilePerms)
}

func (d *Definition) Read() error {
	data, err := os.ReadFile(d.filePath)
	if err != nil {
		return err
	}

	// ensure we reset missing toml fields to zero values instead of retaining old vals
	var loaded Definition
	if err := toml.Unmarshal(data, &loaded); err != nil {
		return err
	}

	loaded.filePath = d.filePath
	if loaded.Configs == nil {
		loaded.Configs = make(map[string]Config)
	}

	*d = loaded
	return nil
}

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/sys/unix"
)

type Definition struct {
	Name        string              `toml:"name"`
	Description string              `toml:"description,omitempty"`
	Resources   map[string]Resource `toml:"resources,omitempty"`
	filePath    string              `toml:"-"`
}

type Resource struct {
	DependsOn []string `toml:"depends_on,omitempty"`
	// explicitly has the config struct at resource.config, rather than flattening it out
	*Config `toml:"config,omitempty"`
}

type Config struct {
	DestPath string `toml:"dest"`
	IsDir    bool   `toml:"is_dir"`
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
	if loaded.Resources == nil {
		loaded.Resources = make(map[string]Resource)
	}

	*d = loaded
	return nil
}

// Lock creates the (special) unix lock file, and returns an error if the lock is already owned.
// Returns a function which unlocks when called.
func (d *Definition) Lock() (func() error, error) {
	lockPath := filepath.Join(filepath.Dir(d.filePath), "hatch.lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		file.Close()

		if errors.Is(err, unix.EWOULDBLOCK) {
			// TODO: perhaps block until we get the lock
			return nil, fmt.Errorf("another hatch process has a lock on this definition (%s): %w", d.filePath, err)
		}

		return nil, err
	}

	return func() error {
		return file.Close()
	}, nil
}

package main

type Definition struct {
	Configs []Config `toml:"config,omitempty"`
}

type Id string
type Path string

type Resource struct {
	Id        Id   `toml:"id"`
	DependsOn []Id `toml:"depends_on,omitempty"`
}

type Config struct {
	Resource
	SourcePath Path `toml:"source"`
	DestPath   Path `toml:"dest"`
}

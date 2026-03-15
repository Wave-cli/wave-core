// Package config handles reading and merging wave configuration from
// the global ~/.wave/config and project-level Wavefile.
package config

// GlobalConfig represents ~/.wave/config.
type GlobalConfig struct {
	Core     CoreConfig        `toml:"core"`
	Projects ProjectsConfig    `toml:"projects"`
	Plugins  map[string]string `toml:"plugins"`
}

// CoreConfig holds core wave settings.
type CoreConfig struct {
	LogsDir string `toml:"logs_dir"`
}

// ProjectsConfig holds project folder locations.
type ProjectsConfig struct {
	Folders []string `toml:"folders"`
}

// Wavefile represents a project-level Wavefile.
type Wavefile struct {
	Project  ProjectMeta
	Sections map[string]map[string]any // per-plugin sections
}

// ProjectMeta holds the [project] section of a Wavefile.
type ProjectMeta struct {
	Name     string   `toml:"name"`
	Version  string   `toml:"version"`
	Owner    string   `toml:"owner"`
	Category string   `toml:"category"`
	Tags     []string `toml:"tags"`
}

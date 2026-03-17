package integrationmapping

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// ExtensionOption modifies Extension configuration
type ExtensionOption func(*Config)

// Config holds configuration for integration mapping generation
type Config struct {
	// OutputDir is the directory where the generated mapping file is written
	OutputDir string
	// PackageName is the Go package name for the generated file
	PackageName string
	// EntPackage is the ent generated package import path used for typed ingest contracts
	EntPackage string
	// GalaPackage is the gala package import path used for typed ingest topics
	GalaPackage string
}

// Extension implements entc.Extension for integration mapping generation
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new integration mapping extension with sensible defaults
func New(opts ...ExtensionOption) *Extension {
	ext := &Extension{
		config: &Config{
			PackageName: "integrationgenerated",
		},
	}

	for _, opt := range opts {
		opt(ext.config)
	}

	return ext
}

// WithOutputDir sets the directory where the generated mapping file is written
func WithOutputDir(dir string) ExtensionOption {
	return func(c *Config) {
		c.OutputDir = dir
	}
}

// WithPackageName sets the Go package name for the generated file
func WithPackageName(name string) ExtensionOption {
	return func(c *Config) {
		c.PackageName = name
	}
}

// WithEntPackage sets the ent generated package import path used for typed ingest contracts
func WithEntPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.EntPackage = path
	}
}

// WithGalaPackage sets the gala package import path used for typed ingest topics
func WithGalaPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.GalaPackage = path
	}
}

// Hooks satisfies the entc.Extension interface
func (e Extension) Hooks() []gen.Hook {
	return []gen.Hook{e.Hook()}
}

// Hook returns the gen.Hook that runs integration mapping generation
func (e Extension) Hook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			if e.config.OutputDir == "" {
				return next.Generate(g)
			}

			data := collectMappingData(g, e.config)
			if len(data.Schemas) == 0 {
				return next.Generate(g)
			}

			if err := writeMappingFile(e.config.OutputDir, data); err != nil {
				return err
			}

			return next.Generate(g)
		})
	}
}

// Annotations satisfies the entc.Extension interface
func (Extension) Annotations() []entc.Annotation {
	return nil
}

// Options satisfies the entc.Extension interface
func (Extension) Options() []entc.Option {
	return nil
}

// Templates satisfies the entc.Extension interface
func (Extension) Templates() []*gen.Template {
	return nil
}

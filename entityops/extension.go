package entityops

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// ExtensionOption modifies Extension configuration
type ExtensionOption func(*Config)

// Config holds configuration for entity operations generation
type Config struct {
	// OutputDir is the directory where generated entity operation files are written
	OutputDir string
	// PackageName is the Go package name for the generated files
	PackageName string
	// EntPackage is the ent generated package import path
	EntPackage string
	// GalaPackage is the gala package import path for typed event topics
	GalaPackage string
	// JsonxPackage is the jsonx package import path
	JsonxPackage string
	// LogxPackage is the logx package import path
	LogxPackage string
	// ContextxPackage is the contextx package import path for typed context keys
	ContextxPackage string
	// CelxPackage is the celx package import path for typed entity expression evaluation
	CelxPackage string
}

// Extension implements entc.Extension for entity operations generation
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new entity operations extension with sensible defaults
func New(opts ...ExtensionOption) *Extension {
	ext := &Extension{
		config: &Config{
			PackageName: "entityops",
		},
	}

	for _, opt := range opts {
		opt(ext.config)
	}

	return ext
}

// WithOutputDir sets the directory where generated files are written
func WithOutputDir(dir string) ExtensionOption {
	return func(c *Config) {
		c.OutputDir = dir
	}
}

// WithPackageName sets the Go package name for generated files
func WithPackageName(name string) ExtensionOption {
	return func(c *Config) {
		c.PackageName = name
	}
}

// WithEntPackage sets the ent generated package import path
func WithEntPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.EntPackage = path
	}
}

// WithGalaPackage sets the gala package import path
func WithGalaPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.GalaPackage = path
	}
}

// WithJsonxPackage sets the jsonx package import path
func WithJsonxPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.JsonxPackage = path
	}
}

// WithLogxPackage sets the logx package import path
func WithLogxPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.LogxPackage = path
	}
}

// WithContextxPackage sets the contextx package import path
func WithContextxPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.ContextxPackage = path
	}
}

// WithCelxPackage sets the celx package import path
func WithCelxPackage(path string) ExtensionOption {
	return func(c *Config) {
		c.CelxPackage = path
	}
}

// Hooks satisfies the entc.Extension interface
func (e Extension) Hooks() []gen.Hook {
	return []gen.Hook{e.Hook()}
}

// Hook returns the gen.Hook that runs entity operations generation
func (e Extension) Hook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			if e.config.OutputDir == "" {
				return next.Generate(g)
			}

			data, err := collectEntityData(g, e.config)
			if err != nil {
				return err
			}

			if len(data.Schemas) == 0 {
				return next.Generate(g)
			}

			if err := generateEntityFiles(e.config.OutputDir, data); err != nil {
				return err
			}

			return next.Generate(g)
		})
	}
}

package oscalgen

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// ExtensionOption is a function that modifies the Extension configuration
type ExtensionOption func(*Extension)

// Config is the configuration for OSCAL mapping generation
type Config struct {
	// SchemaPath is the path to ent schema definitions
	SchemaPath string
	// OutputDir is the directory where OSCAL metadata is generated
	OutputDir string
	// PackageName is the Go package name for generated metadata
	PackageName string
	// BuildFlags are forwarded to ent graph loading
	BuildFlags []string
}

// Extension implements entc.Extension for OSCAL mapping generation
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new OSCAL generation extension
func New(opts ...ExtensionOption) *Extension {
	ext := &Extension{
		config: &Config{
			SchemaPath:  "./schema",
			OutputDir:   "./internal/ent/oscalgenerated",
			PackageName: "oscalgenerated",
			BuildFlags:  nil,
		},
	}

	for _, opt := range opts {
		opt(ext)
	}

	return ext
}

// WithSchemaPath sets the schema path
func WithSchemaPath(schemaPath string) ExtensionOption {
	return func(e *Extension) {
		e.config.SchemaPath = schemaPath
	}
}

// WithGeneratedDir sets the generated output directory
func WithGeneratedDir(outputDir string) ExtensionOption {
	return func(e *Extension) {
		e.config.OutputDir = outputDir
	}
}

// WithPackageName sets the generated package name
func WithPackageName(packageName string) ExtensionOption {
	return func(e *Extension) {
		e.config.PackageName = packageName
	}
}

// WithBuildFlags sets build flags used when loading the ent graph
func WithBuildFlags(flags ...string) ExtensionOption {
	return func(e *Extension) {
		e.config.BuildFlags = append([]string(nil), flags...)
	}
}

// Hooks satisfies the entc.Extension interface
func (e Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		e.Hook(),
	}
}

// Hook runs OSCAL registry generation after ent codegen
func (e Extension) Hook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			if err := next.Generate(g); err != nil {
				return err
			}

			generator := NewOSCALGenerator(e.config.SchemaPath, e.config.OutputDir).
				WithPackage(e.config.PackageName)

			return generator.Generate(e.config.BuildFlags...)
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

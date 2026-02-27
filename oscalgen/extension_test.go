package oscalgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewDefaults validates extension default configuration values
func TestNewDefaults(t *testing.T) {
	ext := New()

	assert.NotNil(t, ext)
	assert.NotNil(t, ext.config)
	assert.Equal(t, "./schema", ext.config.SchemaPath)
	assert.Equal(t, "./internal/ent/oscalgenerated", ext.config.OutputDir)
	assert.Equal(t, "oscalgenerated", ext.config.PackageName)
	assert.Nil(t, ext.config.BuildFlags)
}

// TestOptions validates extension option setters
func TestOptions(t *testing.T) {
	ext := New(
		WithSchemaPath("./ent/schema"),
		WithGeneratedDir("./internal/generated"),
		WithPackageName("generated"),
		WithBuildFlags("-tags=test", "-mod=mod"),
	)

	assert.Equal(t, "./ent/schema", ext.config.SchemaPath)
	assert.Equal(t, "./internal/generated", ext.config.OutputDir)
	assert.Equal(t, "generated", ext.config.PackageName)
	assert.Equal(t, []string{"-tags=test", "-mod=mod"}, ext.config.BuildFlags)
}

// TestHookExposure validates exposed extension interfaces
func TestHookExposure(t *testing.T) {
	ext := New()

	assert.NotNil(t, ext.Hook())
	assert.Len(t, ext.Hooks(), 1)
	assert.NotNil(t, ext.Hooks()[0])
	assert.Nil(t, ext.Annotations())
	assert.Nil(t, ext.Options())
	assert.Nil(t, ext.Templates())
}

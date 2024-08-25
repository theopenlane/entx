package entx

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/stretchr/testify/assert"
)

func TestCascadeAnnotation(t *testing.T) {
	f := gofakeit.Name()
	ca := CascadeAnnotationField(f)

	assert.Equal(t, ca.Name(), CascadeAnnotationName)
	assert.Equal(t, ca.Field, f)
}

func TestCascadeThroughAnnotation(t *testing.T) {
	f := gofakeit.Name()
	s := gofakeit.Name()
	schemas := []ThroughCleanup{
		{
			Through: s,
			Field:   f,
		},
	}
	ca := CascadeThroughAnnotationField(schemas)

	assert.Equal(t, ca.Name(), CascadeThroughAnnotationName)
	assert.Equal(t, ca.Schemas[0].Field, f)
	assert.Equal(t, ca.Schemas[0].Through, s)
}

func TestSchemaGenAnnotation(t *testing.T) {
	s := gofakeit.Bool()
	sa := SchemaGenSkip(s)

	assert.Equal(t, sa.Name(), SchemaGenAnnotationName)
	assert.Equal(t, sa.Skip, s)
}

package entx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithJSONScalar(t *testing.T) {
	ex := &Extension{}
	err := WithJSONScalar()(ex)

	assert.NoError(t, err)
	assert.Len(t, ex.gqlSchemaHooks, 1)
}

func TestNewExtension(t *testing.T) {
	ex, err := NewExtension()

	assert.NoError(t, err)
	assert.NotNil(t, ex)
	assert.Empty(t, ex.templates)
	assert.Empty(t, ex.gqlSchemaHooks)
}

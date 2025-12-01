package entx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkipSoftDelete(t *testing.T) {
	parent := context.Background()
	ctx := SkipSoftDelete(parent)

	assert.True(t, CheckSkipSoftDelete(ctx))
}

func TestCheckSkipSoftDelete(t *testing.T) {
	ctx := context.WithValue(context.Background(), SoftDeleteSkipKey{}, true)

	assert.True(t, CheckSkipSoftDelete(ctx))
	assert.False(t, CheckSkipSoftDelete(context.Background()))
}

func TestIsSoftDelete(t *testing.T) {
	parent := context.Background()
	ctx := IsSoftDelete(parent, "TestObject")

	assert.True(t, CheckIsSoftDeleteType(ctx, "TestObject"))
}

func TestCheckIsSoftDelete(t *testing.T) {
	ctx := context.WithValue(context.Background(), SoftDeleteKey{}, "TestObject")
	assert.True(t, CheckIsSoftDelete(ctx))
	assert.False(t, CheckIsSoftDelete(context.Background()))
}

func TestCheckIsSoftDeleteType(t *testing.T) {
	ctx := context.WithValue(context.Background(), SoftDeleteKey{}, "TestObject")
	assert.True(t, CheckIsSoftDeleteType(ctx, "TestObject"))
	assert.False(t, CheckIsSoftDeleteType(context.Background(), "TestObject"))
}

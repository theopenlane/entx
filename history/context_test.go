package history

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithContextRoundTrip(t *testing.T) {
	ctx := WithContext(context.Background())

	_, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.True(t, IsHistoryRequest(ctx))
}

package enthistory

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonUnmarshalAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		expected  Annotations
		expectErr bool
	}{
		{
			name: "happy path",
			data: map[string]interface{}{
				"exclude":   true,
				"isHistory": false,
			},
			expected: Annotations{
				Exclude:   true,
				IsHistory: false,
			},
		},
		{
			name:     "nil data",
			data:     nil,
			expected: Annotations{},
		},
		{
			name:      "string data",
			data:      "meow",
			expected:  Annotations{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := jsonUnmarshalAnnotations(tt.data)
			if tt.expectErr {
				require.Error(t, err)

				_, ok := err.(*json.UnmarshalTypeError)
				assert.True(t, ok)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, a)
		})
	}
}

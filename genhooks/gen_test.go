package genhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToFirstLower(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{

		{
			name:     "upper first",
			input:    "HelloThere",
			expected: "helloThere",
		},
		{
			name:     "all Upper",
			input:    "HELLO",
			expected: "HELLO",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toFirstLower(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}

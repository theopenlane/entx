package customtypes

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixedIdentifierValue(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		input  string
		want   driver.Value
	}{
		{
			name:   "with prefix",
			prefix: "TSK",
			input:  "TSK-000006",
			want:   6,
		},
		{
			name:   "with wrong prefix",
			prefix: "TEST",
			input:  "TSK-000006",
			want:   0,
		},
		{
			name:   "without prefix",
			prefix: "test",
			input:  "123",
			want:   123,
		},
		{
			name:   "empty input",
			prefix: "test",
			input:  "",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PrefixedIdentifier{prefix: tt.prefix}
			got, err := p.Value(tt.input)
			require.NoError(t, err)

			assert.Equal(t, tt.want, got)
		})
	}
}
func TestPrefixedIdentifierFromValue(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		input   driver.Value
		want    string
		wantErr bool
	}{
		{
			name:   "valid input",
			prefix: "test",
			input:  &sql.NullString{String: "1", Valid: true},
			want:   "test-000001",
		},
		{
			name:   "valid input",
			prefix: "test",
			input:  &sql.NullString{String: "999999", Valid: true},
			want:   "test-999999",
		},
		{
			name:    "invalid input type",
			prefix:  "test",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:   "null string input",
			prefix: "test",
			input:  &sql.NullString{String: "", Valid: false},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PrefixedIdentifier{prefix: tt.prefix}

			got, err := p.FromValue(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

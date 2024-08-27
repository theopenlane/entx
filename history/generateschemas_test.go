package enthistory

import (
	"testing"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldGenerate(t *testing.T) {
	graph, err := entc.LoadGraph("./testdata/schema", &gen.Config{})
	require.NoError(t, err)

	tests := []struct {
		name          string
		schemaName    string
		expectedValue bool
	}{
		{
			name:          "No annotations, include history",
			schemaName:    "User",
			expectedValue: true,
		},
		{
			name:          "Exclude annotation, exclude history",
			schemaName:    "Todo",
			expectedValue: false,
		},
		{
			name:          "History schema, exclude history",
			schemaName:    "UserHistory",
			expectedValue: false,
		},
		{
			name:          "Has annotation, but set to exclude false, include history",
			schemaName:    "List",
			expectedValue: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema *load.Schema

			for _, s := range graph.Schemas {
				if s.Name == tt.schemaName {
					schema = s
					break
				}
			}

			if schema == nil {
				t.Fatalf("schema %s not found", tt.schemaName)
			}

			got := shouldGenerate(schema)

			assert.Equal(t, tt.expectedValue, got)
		})
	}
}

func TestGetAuthzPolicyInfo(t *testing.T) {
	graph, err := entc.LoadGraph("./testdata/schema", &gen.Config{})
	require.NoError(t, err)

	tests := []struct {
		name          string
		schemaName    string
		expectedValue authzPolicyInfo
	}{
		{
			name:       "self referencing schema",
			schemaName: "User",
			expectedValue: authzPolicyInfo{
				ObjectType:      "user",
				NillableIDField: false,
				IDField:         "Ref",
			},
		},
		{
			name:       "org owned schema",
			schemaName: "List",
			expectedValue: authzPolicyInfo{
				ObjectType:      "organization",
				NillableIDField: false,
				IDField:         "OwnerID",
			},
		},
		{
			name:       "nothing set but nillable, use schema name",
			schemaName: "Todo",
			expectedValue: authzPolicyInfo{
				ObjectType:      "todo",
				NillableIDField: true,
				IDField:         "Ref",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var schema *load.Schema

			for _, s := range graph.Schemas {
				if s.Name == tt.schemaName {
					schema = s
					break
				}
			}

			if schema == nil {
				t.Fatalf("schema %s not found", tt.schemaName)
			}

			info := &templateInfo{
				AuthzPolicy: authzPolicyInfo{},
			}

			err := info.getAuthzPolicyInfo(schema)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedValue, info.AuthzPolicy)
		})
	}
}

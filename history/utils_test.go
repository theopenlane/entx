package enthistory

import (
	"testing"

	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "happy path",
			args: "CaptainAmerica",
			want: "captain_america",
		},
		{
			name: "already snake case",
			args: "captain_america",
			want: "captain_america",
		},
		{
			name: "single word",
			args: "hulk",
			want: "hulk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSnakeCase(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetUpdatedByField(t *testing.T) {
	tests := []struct {
		name               string
		updatedByValueType string
		want               *load.Field
		wantErr            bool
	}{
		{
			name:               "happy path - string",
			updatedByValueType: "String",
			want: &load.Field{
				Name:      "updated_by",
				Optional:  true,
				Nillable:  true,
				Immutable: true,
				Info: &field.TypeInfo{
					Type: field.TypeString,
				},
			},
			wantErr: false,
		},
		{
			name:               "happy path  int",
			updatedByValueType: "Int",
			want: &load.Field{
				Name:      "updated_by",
				Optional:  true,
				Nillable:  true,
				Immutable: true,
				Info: &field.TypeInfo{
					Type: field.TypeInt,
				},
			},
			wantErr: false,
		},
		{
			name:               "happy path - lowercase string",
			updatedByValueType: "string",
			want: &load.Field{
				Name:      "updated_by",
				Optional:  true,
				Nillable:  true,
				Immutable: true,
				Info: &field.TypeInfo{
					Type: field.TypeString,
				},
			},
			wantErr: false,
		},
		{
			name:               "happy path - lowercase int",
			updatedByValueType: "int",
			want: &load.Field{
				Name:      "updated_by",
				Optional:  true,
				Nillable:  true,
				Immutable: true,
				Info: &field.TypeInfo{
					Type: field.TypeInt,
				},
			},
			wantErr: false,
		},
		{
			name:               "invalid type",
			updatedByValueType: "json",
			want:               nil,
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUpdatedByField(tt.updatedByValueType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Optional, got.Optional)
			assert.Equal(t, tt.want.Nillable, got.Nillable)
			assert.Equal(t, tt.want.Immutable, got.Immutable)
			assert.Equal(t, tt.want.Info.Type, got.Info.Type)
		})
	}
}

func TestGetHistoryAnnotations(t *testing.T) {
	tests := []struct {
		name   string
		schema *load.Schema
		want   Annotations
	}{
		{
			name: "regular schema, no history, do not exclude",
			schema: &load.Schema{
				Annotations: map[string]any{
					"History": map[string]any{
						"exclude":   false,
						"isHistory": false,
					},
				},
			},
			want: Annotations{
				Exclude:   false,
				IsHistory: false,
			},
		},
		{
			name: "exclude schema",
			schema: &load.Schema{
				Annotations: map[string]any{
					"History": map[string]any{
						"exclude":   true,
						"isHistory": false,
					},
				},
			},
			want: Annotations{
				Exclude:   true,
				IsHistory: false,
			},
		},
		{
			name: "history schema",
			schema: &load.Schema{
				Annotations: map[string]any{
					"History": map[string]any{
						"exclude":   false,
						"isHistory": true,
					},
				},
			},
			want: Annotations{
				Exclude:   false,
				IsHistory: true,
			},
		},
		{
			name:   "empty annotation, should return false",
			schema: &load.Schema{},
			want: Annotations{
				Exclude:   false,
				IsHistory: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHistoryAnnotations(tt.schema)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetSchemaTableName(t *testing.T) {
	tests := []struct {
		name   string
		schema *load.Schema
		want   string
	}{
		{
			name: "happy path",
			schema: &load.Schema{
				Name: "MattIsTheBest",
				Annotations: map[string]any{
					"EntSQL": map[string]any{
						"table": "matt_is_the_best",
					},
				},
			},
			want: "matt_is_the_best",
		},
		{
			name: "not set, should use schema name",
			schema: &load.Schema{
				Name: "MattIsTheBest",
			},
			want: "matt_is_the_best",
		},
		{
			name: "empty string, use schema name",
			schema: &load.Schema{
				Name: "MattIsTheBest",
				Annotations: map[string]any{
					"EntSQL": map[string]any{
						"table": "",
					},
				},
			},
			want: "matt_is_the_best",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSchemaTableName(tt.schema)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPkgFromSchemaPath(t *testing.T) {
	tests := []struct {
		name       string
		schemaPath string
		want       string
		wantErr    bool
	}{
		{
			name:       "happy path",
			schemaPath: "github.com/golanglemonade/foobar",
			want:       "foobar",
			wantErr:    false,
		},
		{
			name:       "invalid",
			schemaPath: "github.com/golanglemonade/foobar/",
			want:       "",
			wantErr:    true,
		},
		{
			name:       "empty string",
			schemaPath: "",
			want:       "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPkgFromSchemaPath(tt.schemaPath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetIDType(t *testing.T) {
	tests := []struct {
		name   string
		idType string
		want   string
	}{
		{
			name:   "string lower",
			idType: "string",
			want:   "string",
		},
		{
			name:   "string title",
			idType: "String",
			want:   "string",
		},
		{
			name:   "string crazy",
			idType: "StRiNg",
			want:   "string",
		},
		{
			name:   "int lower",
			idType: "int",
			want:   "int",
		},
		{
			name:   "int title",
			idType: "Int",
			want:   "int",
		},
		{
			name:   "int crazy",
			idType: "InT",
			want:   "int",
		},
		{
			name:   "not cool",
			idType: "BoolFool",
			want:   "string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIDType(tt.idType)

			assert.Equal(t, tt.want, got)
		})
	}
}

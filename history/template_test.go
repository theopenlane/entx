package enthistory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractUpdatedByKey(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{
			name: "happy path",
			val: &UpdatedBy{
				key:       "userID",
				valueType: ValueTypeString,
			},
			want: "userID",
		},
		{
			name: "nil updated by",
			val:  &UpdatedBy{},
			want: "",
		},
		{
			name: "bad type",
			val:  "something else",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUpdatedByKey(tt.val)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExxtractUpdatedByValueType(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want string
	}{
		{
			name: "happy path, string",
			val: &UpdatedBy{
				key:       "userID",
				valueType: 1, // 1 is string
			},
			want: "string",
		},
		{
			name: "happy path, int",
			val: &UpdatedBy{
				key:       "userID",
				valueType: 0, // 0 is int
			},
			want: "int",
		},
		{
			name: "invalid type",
			val: &UpdatedBy{
				key:       "userID",
				valueType: 42,
			},
			want: "",
		},
		{
			name: "empty type, defaults to int",
			val: &UpdatedBy{
				key: "userID",
			},
			want: "int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUpdatedByValueType(tt.val)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFieldPropertiesNillable(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   bool
	}{
		{
			name: "nillable true",
			config: Config{
				FieldProperties: &FieldProperties{
					Nillable: true,
				},
			},
			want: true,
		},
		{
			name: "nillable false",
			config: Config{
				FieldProperties: &FieldProperties{
					Nillable: false,
				},
			},
			want: false,
		},
		{
			name: "not set",
			config: Config{
				FieldProperties: &FieldProperties{},
			},
			want: false,
		},
		{
			name:   "nil config",
			config: Config{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldPropertiesNillable(tt.config)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSlice(t *testing.T) {
	tests := []struct {
		name       string
		typeString string
		want       bool
	}{
		{
			name:       "is slice",
			typeString: "[]string{}",
			want:       true,
		},
		{
			name:       "something else",
			typeString: "bool",
			want:       false,
		},
		{
			name:       "empty",
			typeString: "",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSlice(tt.typeString)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIn(t *testing.T) {
	type args struct {
		str  string
		list []string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "its there! (or should be)",
			args: args{
				str:  "funk",
				list: []string{"anderson", "funk"},
			},
			want: true,
		},
		{
			name: "its not there :( (or shouldn't be)",
			args: args{
				str:  "funkhouser",
				list: []string{"anderson", "funk"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := in(tt.args.str, tt.args.list)
			assert.Equal(t, tt.want, got)
		})
	}
}

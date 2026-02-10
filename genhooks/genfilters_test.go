package genhooks

import (
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/stretchr/testify/assert"
)

func TestEntSkipWhere(t *testing.T) {
	tests := []struct {
		name string
		typ  *load.Schema
		want bool
	}{
		{
			name: "No annotation",
			typ:  &load.Schema{},
			want: false,
		},
		{
			name: "SkipWhereInput annotation set to true",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{
						Skip: entgql.SkipWhereInput,
					},
				},
			},
			want: true,
		},
		{
			name: "SkipAll annotation set to true",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{
						Skip: entgql.SkipAll,
					},
				},
			},
			want: true,
		},
		{
			name: "SkipType annotation set to true",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{
						Skip: entgql.SkipType,
					},
				},
			},
			want: true,
		},
		{
			name: "Skip Create annotation results in false",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{
						Skip: entgql.SkipMutationCreateInput,
					},
				},
			},
			want: false,
		},
		{
			name: "Skip Update Input annotation results in false",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{
						Skip: entgql.SkipMutationUpdateInput,
					},
				},
			},
			want: false,
		},
		{
			name: "No skip results in false",
			typ: &load.Schema{
				Annotations: gen.Annotations{
					entgql.Annotation{}.Name(): entgql.Annotation{},
				},
			},
			want: false,
		},
		{
			name: "Skip when set on field",
			typ: &load.Schema{
				Fields: []*load.Field{
					{
						Name: "TestField",
						Annotations: gen.Annotations{
							entgql.Annotation{}.Name(): entgql.Annotation{
								Skip: entgql.SkipWhereInput,
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Skip all when set on field results in true",
			typ: &load.Schema{
				Fields: []*load.Field{
					{
						Name: "TestField",
						Annotations: gen.Annotations{
							entgql.Annotation{}.Name(): entgql.Annotation{
								Skip: entgql.SkipAll,
							},
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			if tc.typ.Fields != nil {
				got := entSkipWhere(tc.typ.Fields[0])
				assert.Equal(t, tc.want, got)
				return
			}

			got := entSkipWhere(tc.typ)
			assert.Equal(t, tc.want, got)
		})
	}
}

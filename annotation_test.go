package entx

import (
	"encoding/json"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCascadeAnnotation(t *testing.T) {
	f := gofakeit.Name()
	ca := CascadeAnnotationField(f)

	assert.Equal(t, ca.Name(), CascadeAnnotationName)
	assert.Equal(t, ca.Field, f)
}

func TestCascadeThroughAnnotation(t *testing.T) {
	f := gofakeit.Name()
	s := gofakeit.Name()
	schemas := []ThroughCleanup{
		{
			Through: s,
			Field:   f,
		},
	}
	ca := CascadeThroughAnnotationField(schemas)

	assert.Equal(t, ca.Name(), CascadeThroughAnnotationName)
	assert.Equal(t, ca.Schemas[0].Field, f)
	assert.Equal(t, ca.Schemas[0].Through, s)
}

func TestSchemaGenAnnotation(t *testing.T) {
	s := gofakeit.Bool()
	sa := SchemaGenSkip(s)

	assert.Equal(t, sa.Name(), SchemaGenAnnotationName)
	assert.Equal(t, sa.Skip, s)
}

func TestExportableAnnotation(t *testing.T) {
	ea := &Exportable{}

	assert.Equal(t, ea.Name(), "Exportable")

	err := ea.Decode(map[string]any{})
	assert.NoError(t, err)
}

func TestWorkflowEligibleAnnotation(t *testing.T) {
	wea := FieldWorkflowEligible()

	assert.Equal(t, wea.Name(), WorkflowEligibleAnnotationName)
	assert.True(t, wea.Eligible)

	// Test Decode method
	decoded := &WorkflowEligibleAnnotation{}
	err := decoded.Decode(map[string]any{"Eligible": true})
	assert.NoError(t, err)
	assert.True(t, decoded.Eligible)
}

func TestCSVRefBuilder(t *testing.T) {
	tests := []struct {
		name               string
		builder            *CSVRefBuilder
		expectedMatchField string
		expectedColumn     string
		expectedTarget     string
		expectedCreate     bool
	}{
		{
			name: "basic user email lookup",
			builder: CSVRef().
				FromColumn("AssignedToUserEmail").
				MatchOn("email"),
			expectedMatchField: "email",
			expectedColumn:     "AssignedToUserEmail",
		},
		{
			name: "group name lookup",
			builder: CSVRef().
				FromColumn("BlockedGroupNames").
				MatchOn("name"),
			expectedMatchField: "name",
			expectedColumn:     "BlockedGroupNames",
		},
		{
			name: "platform with create if missing",
			builder: CSVRef().
				FromColumn("AccessPlatformNames").
				MatchOn("name").
				CreateIfMissing(),
			expectedMatchField: "name",
			expectedColumn:     "AccessPlatformNames",
			expectedCreate:     true,
		},
		{
			name: "entity name lookup",
			builder: CSVRef().
				FromColumn("EntityName").
				MatchOn("name"),
			expectedMatchField: "name",
			expectedColumn:     "EntityName",
		},
		{
			name: "control ref code with explicit target",
			builder: CSVRef().
				FromColumn("ControlRefCode").
				MatchOn("ref_code").
				TargetEntity("Control"),
			expectedMatchField: "ref_code",
			expectedColumn:     "ControlRefCode",
			expectedTarget:     "Control",
		},
		{
			name: "identity holder email lookup",
			builder: CSVRef().
				FromColumn("IdentityHolderEmail").
				MatchOn("email"),
			expectedMatchField: "email",
			expectedColumn:     "IdentityHolderEmail",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, CSVReferenceAnnotationName, tc.builder.Name())
			assert.Equal(t, tc.expectedMatchField, tc.builder.annotation.MatchField)
			assert.Equal(t, tc.expectedColumn, tc.builder.annotation.CSVColumn)
			assert.Equal(t, tc.expectedTarget, tc.builder.annotation.TargetEntity)
			assert.Equal(t, tc.expectedCreate, tc.builder.annotation.CreateIfMissing)
		})
	}
}

func TestCSVRefBuilderMarshalJSON(t *testing.T) {
	builder := CSVRef().
		FromColumn("BlockedGroupNames").
		MatchOn("name").
		TargetEntity("Group").
		CreateIfMissing()

	data, err := json.Marshal(builder)
	require.NoError(t, err)

	// Verify it marshals as CSVReferenceAnnotation, not as CSVRefBuilder
	var decoded CSVReferenceAnnotation

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "name", decoded.MatchField)
	assert.Equal(t, "BlockedGroupNames", decoded.CSVColumn)
	assert.Equal(t, "Group", decoded.TargetEntity)
	assert.True(t, decoded.CreateIfMissing)
}

func TestIntegrationMappingSchemaLookupAlternativeAndInstanceScoped(t *testing.T) {
	b := IntegrationMappingSchema().
		LookupAlternative("external_id").
		LookupAlternative("email", "tenant_id").
		InstanceScoped()

	assert.Equal(t, IntegrationMappingSchemaAnnotationName, b.Name())
	assert.Equal(t, [][]string{{"external_id"}, {"email", "tenant_id"}}, b.annotation.LookupAlternatives)
	assert.True(t, b.annotation.InstanceScoped)

	raw, err := b.MarshalJSON()
	require.NoError(t, err)

	decoded := &IntegrationMappingSchemaAnnotation{}
	require.NoError(t, json.Unmarshal(raw, decoded))
	assert.Equal(t, [][]string{{"external_id"}, {"email", "tenant_id"}}, decoded.LookupAlternatives)
	assert.True(t, decoded.InstanceScoped)
}

func TestIntegrationMappingSchemaMerge(t *testing.T) {
	base := IntegrationMappingSchema().StockPersist()
	override := IntegrationMappingSchema().Exclude("stakeholder_id").LookupAlternative("email").InstanceScoped()

	merged, ok := base.Merge(override).(*IntegrationMappingSchemaBuilder)
	require.True(t, ok)
	assert.True(t, merged.annotation.StockPersist)
	assert.True(t, merged.annotation.InstanceScoped)
	assert.Equal(t, []string{"stakeholder_id"}, merged.annotation.Exclude)
	assert.Equal(t, [][]string{{"email"}}, merged.annotation.LookupAlternatives)
}

func TestSnapshotRemovalAnnotation(t *testing.T) {
	b := SnapshotRemoval().Episodic()

	assert.Equal(t, SnapshotRemovalAnnotationName, b.Name())
	assert.True(t, b.annotation.Episodic)

	raw, err := b.MarshalJSON()
	require.NoError(t, err)

	decoded := &SnapshotRemovalAnnotation{}
	require.NoError(t, decoded.Decode(json.RawMessage(raw)))
	assert.True(t, decoded.Episodic)
}

func TestFieldSourceManagedAnnotation(t *testing.T) {
	b := FieldSourceManaged()

	assert.Equal(t, FieldSourceManagedAnnotationName, b.Name())

	raw, err := b.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(raw))

	decoded := &FieldSourceManagedAnnotation{}
	require.NoError(t, decoded.Decode(json.RawMessage(raw)))
	assert.Equal(t, FieldSourceManagedAnnotationName, decoded.Name())
}

func TestCatalogEdgeAnnotation(t *testing.T) {
	b := CatalogEdge()

	assert.Equal(t, CatalogEdgeAnnotationName, b.Name())

	raw, err := b.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(raw))

	decoded := &CatalogEdgeAnnotation{}
	require.NoError(t, decoded.Decode(json.RawMessage(raw)))
	assert.Equal(t, CatalogEdgeAnnotationName, decoded.Name())
}

func TestCatalogVisibilityFieldAnnotation(t *testing.T) {
	b := CatalogVisibilityField()

	assert.Equal(t, CatalogVisibilityFieldAnnotationName, b.Name())

	raw, err := b.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(raw))

	decoded := &CatalogVisibilityFieldAnnotation{}
	require.NoError(t, decoded.Decode(json.RawMessage(raw)))
	assert.Equal(t, CatalogVisibilityFieldAnnotationName, decoded.Name())
}

func TestCatalogKeyFieldAnnotation(t *testing.T) {
	b := CatalogKeyField()

	assert.Equal(t, CatalogKeyFieldAnnotationName, b.Name())

	raw, err := b.MarshalJSON()
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(raw))

	decoded := &CatalogKeyFieldAnnotation{}
	require.NoError(t, decoded.Decode(json.RawMessage(raw)))
	assert.Equal(t, CatalogKeyFieldAnnotationName, decoded.Name())
}

func TestCSVReferenceAnnotationDecode(t *testing.T) {
	decoded := &CSVReferenceAnnotation{}
	err := decoded.Decode(map[string]any{
		"MatchField":      "email",
		"CSVColumn":       "UserEmail",
		"TargetEntity":    "User",
		"CreateIfMissing": true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "email", decoded.MatchField)
	assert.Equal(t, "UserEmail", decoded.CSVColumn)
	assert.Equal(t, "User", decoded.TargetEntity)
	assert.True(t, decoded.CreateIfMissing)
}

func TestGetAnnotation(t *testing.T) {
	ant := SchemaGenAnnotation{}

	tests := []struct {
		name        string
		node        *gen.Type
		expectedOk  bool
		expectedVal *SchemaGenAnnotation
	}{
		{
			name: "annotation exists as pointer",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): &SchemaGenAnnotation{Skip: true, SkipSearch: true},
				},
			},
			expectedOk:  true,
			expectedVal: &SchemaGenAnnotation{Skip: true, SkipSearch: true},
		},
		{
			name: "annotation missing",
			node: &gen.Type{
				Annotations: map[string]interface{}{},
			},
			expectedOk:  false,
			expectedVal: nil,
		},
		{
			name: "annotation stored as value, not pointer",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): SchemaGenAnnotation{Skip: true},
				},
			},
			expectedOk:  true,
			expectedVal: &SchemaGenAnnotation{Skip: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := GetAnnotation[*SchemaGenAnnotation](tt.node)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedVal, val)
		})
	}
}

func TestHasAnnotation(t *testing.T) {
	ant := OrgOwnedSchema{}

	tests := []struct {
		name     string
		node     *gen.Type
		expected bool
	}{
		{
			name: "has annotation",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): &OrgOwnedSchema{},
				},
			},
			expected: true,
		},
		{
			name: "missing annotation",
			node: &gen.Type{
				Annotations: map[string]interface{}{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasAnnotation[OrgOwnedSchema](tt.node))
		})
	}
}

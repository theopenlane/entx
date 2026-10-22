package entityops

import (
	"bytes"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema"
	entfield "entgo.io/ent/schema/field"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/entx"
)

func TestCollectAnonymousFields(t *testing.T) {
	anon := gen.Annotations{entx.AnonymousFieldAnnotationName: entx.AnonymousField()}
	node := &gen.Type{
		Name: "Subscriber",
		Fields: []*gen.Field{
			{Name: "email", Annotations: anon},
			{Name: "trust_center_id", Annotations: anon},
			{Name: "verified_email"},
		},
	}

	markers, err := collectFieldMarkers(node)
	require.NoError(t, err)

	schema := &EntitySchema{}
	require.NoError(t, applyFieldMarkers(schema, node.Name, markers))
	require.Equal(t, []string{"email", "trustCenterID"}, schema.AnonymousInputFields)

	markers, err = collectFieldMarkers(&gen.Type{Name: "Widget", Fields: []*gen.Field{{Name: "name"}}})
	require.NoError(t, err)

	schema = &EntitySchema{}
	require.NoError(t, applyFieldMarkers(schema, "Widget", markers))
	require.Empty(t, schema.AnonymousInputFields)
}

// TestStaticTemplatesRenderValidGo verifies the schema-independent templates render to
// syntactically valid Go for a fully configured generator
func TestStaticTemplatesRenderValidGo(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		MapxPackage:  "example.com/app/pkg/mapx",
	}

	for _, name := range []string{
		"entity_schema", "entity_errors", "entity_registry", "entity_workflow", "entity_tasks",
		"entity_links", "entity_integration", "entity_projection", "entity_metadata",
		"entity_changeset", "entity_mutation_events", "entity_listener",
	} {
		t.Run(name, func(t *testing.T) {
			tmpl, err := parseTemplate(name)
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, tmpl.Execute(&buf, data))

			_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
			require.NoError(t, err, buf.String())
		})
	}
}

// TestCapabilityGatedEmission verifies the registry and projection templates emit runtime
// closures and projections only for schemas whose capabilities can reach them: mapped schemas
// get create/update/query, workflow-eligible schemas get load-object, link targets get query
// and a projection, and capability-free schemas get only the universal load and catalogs
func TestCapabilityGatedEmission(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Asset", Snake: "asset", Lower: "asset",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateAssetInput", UpdateInputType: "UpdateAssetInput",
				PredicatePackage: "asset", PredicateImport: "example.com/app/ent/generated/asset",
				IntegrationMapped: true,
				ObjectFields:      []EntityField{{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true}},
				Edges:             []EntityEdge{{Name: "controls", TargetSchema: "Control", TargetInRegistry: true}},
			},
			{
				Name: "Control", Snake: "control", Lower: "control",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateControlInput", UpdateInputType: "UpdateControlInput",
				PredicatePackage: "control", PredicateImport: "example.com/app/ent/generated/control",
				LinkTarget:       true,
				WorkflowEligible: true,
				ObjectFields:     []EntityField{{Name: "RefCode", Snake: "ref_code", Type: "string", MatchKey: true, Projectable: true, WorkflowEligible: true}},
			},
			{
				Name: "Note", Snake: "note", Lower: "note",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateNoteInput", UpdateInputType: "UpdateNoteInput",
				PredicatePackage: "note", PredicateImport: "example.com/app/ent/generated/note",
				ObjectFields: []EntityField{{Name: "Text", Snake: "text", Type: "string", Projectable: true}},
			},
		},
	}

	render := func(name string) string {
		tmpl, err := parseTemplate(name)
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, tmpl.Execute(&buf, data))

		_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
		require.NoError(t, err, buf.String())

		return buf.String()
	}

	registry := render("entity_registry")

	asset := registry[strings.Index(registry, "SchemaAsset = &Schema{"):strings.Index(registry, "SchemaControl = &Schema{")]
	require.Contains(t, asset, "Create: func")
	require.Contains(t, asset, "Update: func")
	require.Contains(t, asset, "Query: func")
	require.Contains(t, asset, "Ingest: &IngestCapability")
	require.NotContains(t, asset, "LoadObject: func")

	control := registry[strings.Index(registry, "SchemaControl = &Schema{"):strings.Index(registry, "SchemaNote = &Schema{")]
	require.NotContains(t, control, "Create: func")
	require.Contains(t, control, "Update: func")
	require.Contains(t, control, "Query: func")
	require.Contains(t, control, "LoadObject: func")

	note := registry[strings.Index(registry, "SchemaNote = &Schema{"):strings.Index(registry, "// init wires")]
	require.Contains(t, note, "Load: func")
	require.NotContains(t, note, "Create: func")
	require.NotContains(t, note, "Update: func")
	require.NotContains(t, note, "Query: func")
	require.NotContains(t, note, "LoadObject: func")
	require.NotContains(t, note, "ProjectionType")

	require.Contains(t, registry, "SchemaAsset.QueryByKey = func")
	require.Contains(t, registry, "SchemaControl.QueryByKey = func")
	require.NotContains(t, registry, "SchemaNote.QueryByKey")
	require.NotContains(t, registry, "LoadMany")
	require.NotContains(t, registry, "func CatalogListeners()")
	require.NotContains(t, registry, "theopenlane/iam/auth")

	projections := render("entity_projection")
	require.Contains(t, projections, "type AssetProjection struct")
	require.Contains(t, projections, "type ControlProjection struct")
	require.NotContains(t, projections, "NoteProjection")
}

// TestValidateProvenanceFields verifies ingest-capable schemas missing provenance fields are rejected
func TestValidateProvenanceFields(t *testing.T) {
	provenance := []EntityField{
		{Name: "SourceDefinitionID", Snake: "source_definition_id", Type: "string"},
		{Name: "SourceDefinitionVersion", Snake: "source_definition_version", Type: "string"},
		{Name: "SourceInstanceID", Snake: "source_instance_id", Type: "string"},
		{Name: "ManagedBy", Snake: "managed_by", Type: "string"},
	}

	complete := EntitySchema{Name: "Asset", IntegrationMapped: true, HasCreate: true, ObjectFields: provenance}
	require.NoError(t, validateProvenanceFields(complete))

	missing := EntitySchema{Name: "Asset", IntegrationMapped: true, HasCreate: true, ObjectFields: provenance[:2]}
	require.ErrorIs(t, validateProvenanceFields(missing), ErrProvenanceFieldsMissing)

	unmapped := EntitySchema{Name: "Note", HasCreate: true}
	require.NoError(t, validateProvenanceFields(unmapped))
}

// TestFieldContractTemplates verifies the field catalog and mapping index template output
func TestFieldContractTemplates(t *testing.T) {
	typ := reflect.TypeFor[EntityField]()

	for _, name := range []string{"CompareKind", "UpdateInput", "Nillable", "FromIntegration", "IntegrationField"} {
		_, ok := typ.FieldByName(name)
		require.Falsef(t, ok, "EntityField still contains %s", name)
	}

	for _, name := range []string{"SystemControlled", "Volatile", "CaseInsensitive"} {
		_, ok := typ.FieldByName(name)
		require.Truef(t, ok, "EntityField missing %s", name)
	}

	require.True(t, isIntegrationSystemField("updated_by_impersonator"), "updated_by_impersonator must be a system-controlled field")

	data := EntityData{
		PackageName: "entityops",
		EntPackage:  "example.com/app/ent/generated",
		GalaPackage: "example.com/app/pkg/gala",
		Schemas: []EntitySchema{
			{
				Name: "Asset", Snake: "asset", Lower: "asset",
				IntegrationMapped: true,
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", IntegrationMapped: true, InputKey: "external_id"},
					{Name: "SourceDefinitionID", Snake: "source_definition_id", Type: "string", IntegrationMapped: true, InputKey: "source_definition_id", SystemControlled: true},
				},
			},
		},
	}

	tmpl, err := parseTemplate("entity_schema")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_schema.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	schema := buf.String()
	require.Contains(t, schema, "SystemControlled bool")
	require.Contains(t, schema, "Volatile bool")
	require.Contains(t, schema, "CaseInsensitive bool")
	require.Contains(t, schema, "func (d FieldDescriptor) Equal(old, proposed any) bool")
	require.Contains(t, schema, "strings.EqualFold(a, b)")
	require.NotContains(t, schema, "\tFromIntegration bool")
	require.NotContains(t, schema, "\tIntegrationField string")

	tmpl, err = parseTemplate("entity_integration")
	require.NoError(t, err)

	buf.Reset()
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_integration.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	index := buf.String()
	require.Contains(t, index, `ExternalID: FieldDescriptor{Name: "external_id", Label: "ExternalID", Type: "string", InputKey: "external_id"}`)
	require.NotContains(t, index, "SourceDefinitionID")
	require.NotContains(t, index, "SchemaAsset")
	require.NotContains(t, index, "func init")

	registryData := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Asset", Snake: "asset", Lower: "asset",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateAssetInput", UpdateInputType: "UpdateAssetInput",
				PredicatePackage: "asset", PredicateImport: "example.com/app/ent/generated/asset",
				IntegrationMapped: true,
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
					{Name: "Status", Snake: "status", Type: "string", IntegrationMapped: true, InputKey: "status", CaseInsensitive: true},
					{Name: "UpdatedByImpersonator", Snake: "updated_by_impersonator", Type: "string", SystemControlled: isIntegrationSystemField("updated_by_impersonator")},
				},
				Edges: []EntityEdge{
					{
						Name: "controls", TargetSchema: "Control",
						ThroughType: "AssetControl", ThroughSourceSetter: "SetAssetID", ThroughTargetSetter: "SetControlID",
						ThroughSourceField: "AssetID", ThroughTargetField: "ControlID",
					},
				},
			},
		},
	}

	tmpl, err = parseTemplate("entity_registry")
	require.NoError(t, err)

	buf.Reset()
	require.NoError(t, tmpl.Execute(&buf, registryData))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_registry.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	registry := buf.String()
	require.Contains(t, registry, "CaseInsensitive: true")
	require.Contains(t, registry, `s.OwnerField != "" && lookupValue(candidate, s.OwnerField) != ownerID`)
	require.Contains(t, registry, "FieldSourceDefinitionID")
	require.Contains(t, registry, "FieldSourceInstanceID")
	require.NotContains(t, registry, "managedByFieldName")
	require.NotContains(t, registry, "managerFieldName")
	require.Contains(t, registry, `!lo.Contains(volatile, edge.Field)`)
	require.Contains(t, registry, `func (s *Schema) SystemControlledOnly(set ChangeSet) bool`)
	require.Contains(t, registry, `saveCtx = WithEmissionVetoed(ctx)`)
	require.Contains(t, registry, `return (!changes.Empty() && !bookkeeping) || len(throughIDs) > 0, true, nil`)
	require.Contains(t, registry, "client.AssetControl.Query()")
	require.Contains(t, registry, "linked[row.ControlID]")
	require.Contains(t, registry, `{Name: "updated_by_impersonator", Label: "UpdatedByImpersonator", Type: "string", SystemControlled: true}`)
	require.Contains(t, registry, "RunID string `json:\"runId,omitempty\"`")
	require.NotContains(t, registry, "'{changed}'")
}

// TestDetectIntegrationEdges verifies integration edge detection
func TestDetectIntegrationEdges(t *testing.T) {
	t.Run("mutable FK", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration", TargetSchema: "Integration", Unique: true, Field: "integration_id"}},
		}
		detectIntegrationEdges(&schema)
		require.Equal(t, "integration_id", schema.IntegrationFKField)
		require.Empty(t, schema.IntegrationM2MEdge)
	})

	t.Run("immutable FK excluded", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration", TargetSchema: "Integration", Unique: true, Field: "integration_id", Immutable: true}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationFKField)
		require.Empty(t, schema.IntegrationM2MEdge)
	})

	t.Run("many-to-many edge", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integrations", TargetSchema: "Integration"}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationFKField)
		require.Equal(t, "integrations", schema.IntegrationM2MEdge)
	})

	t.Run("through edge excluded", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integrations", TargetSchema: "Integration", ThroughType: "SomeJoin"}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationM2MEdge)
	})

	t.Run("immutable many-to-many edge excluded", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integrations", TargetSchema: "Integration", Immutable: true}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationM2MEdge)
	})

	t.Run("no integration edge", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "owner", TargetSchema: "Organization", Unique: true, Field: "owner_id"}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationFKField)
	})

	t.Run("integration run many-to-many edge", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration_runs", TargetSchema: "IntegrationRun"}},
		}
		detectIntegrationEdges(&schema)
		require.Equal(t, "integration_runs", schema.IntegrationRunM2MEdge)
		require.Empty(t, schema.IntegrationFKField)
	})

	t.Run("immutable integration run edge excluded", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration_runs", TargetSchema: "IntegrationRun", Immutable: true}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationRunM2MEdge)
	})

	t.Run("through integration run edge excluded", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration_runs", TargetSchema: "IntegrationRun", ThroughType: "SomeJoin"}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationRunM2MEdge)
	})

	t.Run("unique integration run edge is not the many-to-many case", func(t *testing.T) {
		schema := EntitySchema{
			Edges: []EntityEdge{{Name: "integration_run", TargetSchema: "IntegrationRun", Unique: true, Field: "integration_run_id"}},
		}
		detectIntegrationEdges(&schema)
		require.Empty(t, schema.IntegrationRunM2MEdge)
	})
}

// TestEntityMetadataTemplate verifies the metadata template renders console paths and
// mention specs into compilable map literals
func TestEntityMetadataTemplate(t *testing.T) {
	tmpl, err := parseTemplate("entity_metadata")
	require.NoError(t, err)

	data := EntityData{
		PackageName: "entityops",
	}

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	rendered := buf.String()
	require.Contains(t, rendered, "func ConsoleLanding(schemaType string) string")
	require.Contains(t, rendered, "func ConsoleObjectPath(schemaType, objectID string) string")
	require.Contains(t, rendered, "func MentionSpecFor(schemaType string) (MentionSpec, bool)")
}

// TestEntityMetadataTemplateEmpty verifies the metadata template renders with no annotated
// schemas, since most consuming repos start with empty metadata
func TestEntityMetadataTemplateEmpty(t *testing.T) {
	tmpl, err := parseTemplate("entity_metadata")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, EntityData{PackageName: "entityops"}))

	require.Contains(t, buf.String(), "schema, ok := LookupSchema(schemaType)")
}

// TestSynthesizeLookupAlternatives verifies declared alternatives win over the synthesized LookupKey default
func TestSynthesizeLookupAlternatives(t *testing.T) {
	require.Nil(t, synthesizeLookupAlternatives(nil, nil))
	require.Equal(t, [][]string{{"external_id"}}, synthesizeLookupAlternatives(nil, []string{"external_id"}))
	require.Equal(t, [][]string{{"external_id", "tenant_id"}}, synthesizeLookupAlternatives(nil, []string{"external_id", "tenant_id"}))

	declared := [][]string{{"email"}, {"external_id", "tenant_id"}}
	require.Equal(t, declared, synthesizeLookupAlternatives(declared, []string{"external_id"}))
}

// TestValidateLookupAlternatives verifies lookup alternatives over unknown or unmapped fields are rejected
func TestValidateLookupAlternatives(t *testing.T) {
	mapped := []EntityField{
		{Snake: "external_id", IntegrationMapped: true},
		{Snake: "tenant_id", IntegrationMapped: true},
		{Snake: "internal_note", IntegrationMapped: false},
	}

	t.Run("no alternatives is valid", func(t *testing.T) {
		require.NoError(t, validateLookupAlternatives(EntitySchema{Name: "Widget"}))
	})

	t.Run("valid composite alternative", func(t *testing.T) {
		schema := EntitySchema{Name: "Widget", IntegrationMapped: true, ObjectFields: mapped, LookupAlternatives: [][]string{{"external_id", "tenant_id"}}}
		require.NoError(t, validateLookupAlternatives(schema))
	})

	t.Run("unknown field", func(t *testing.T) {
		schema := EntitySchema{Name: "Widget", IntegrationMapped: true, ObjectFields: mapped, LookupAlternatives: [][]string{{"does_not_exist"}}}
		require.ErrorIs(t, validateLookupAlternatives(schema), ErrLookupAlternativeFieldUnknown)
	})

	t.Run("unmapped field", func(t *testing.T) {
		schema := EntitySchema{Name: "Widget", IntegrationMapped: true, ObjectFields: mapped, LookupAlternatives: [][]string{{"internal_note"}}}
		require.ErrorIs(t, validateLookupAlternatives(schema), ErrLookupAlternativeFieldUnknown)
	})

	t.Run("without integration mapping", func(t *testing.T) {
		schema := EntitySchema{Name: "Widget", IntegrationMapped: false, ObjectFields: mapped, LookupAlternatives: [][]string{{"external_id"}}}
		require.ErrorIs(t, validateLookupAlternatives(schema), ErrLookupAlternativeWithoutMapping)
	})
}

// TestApplyFieldMarkersSnapshotRemoval verifies snapshot-removal field markers fold onto the schema
func TestApplyFieldMarkersSnapshotRemoval(t *testing.T) {
	schema := &EntitySchema{Name: "Widget", OrgOwned: true}

	require.NoError(t, applyFieldMarkers(schema, "Widget", schemaFieldMarkers{
		removedAt:         "removed_at",
		removedAtEpisodic: true,
	}))

	require.Equal(t, "removed_at", schema.RemovedAtField)
	require.True(t, schema.RemovedAtEpisodic)
}

// TestIngestPhaseATemplateEmission verifies the ingest closures are emitted per schema capability
func TestIngestPhaseATemplateEmission(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Widget", Snake: "widget", Lower: "widget",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateWidgetInput", UpdateInputType: "UpdateWidgetInput",
				PredicatePackage: "widget", PredicateImport: "example.com/app/ent/generated/widget",
				IntegrationMapped:     true,
				InstanceScoped:        true,
				LookupAlternatives:    [][]string{{"external_id"}},
				LookupWhereNull:       []string{"standard_id"},
				RemovedAtField:        "removed_at",
				RemovedAtEpisodic:     true,
				HasIntegrationRunID:   true,
				IntegrationRunM2MEdge: "integration_runs",
				HasSourceProvenance:   true,
				HasUpdatedAt:          true,
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
				},
			},
			{
				Name: "Gadget", Snake: "gadget", Lower: "gadget",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateGadgetInput", UpdateInputType: "UpdateGadgetInput",
				PredicatePackage: "gadget", PredicateImport: "example.com/app/ent/generated/gadget",
				IntegrationMapped: true,
			},
		},
	}

	render := func(name string) string {
		tmpl, err := parseTemplate(name)
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, tmpl.Execute(&buf, data))

		_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
		require.NoError(t, err, buf.String())

		return buf.String()
	}

	registry := render("entity_registry")

	require.Contains(t, registry, "func WithActiveIntegrations(ctx context.Context, ids []string) context.Context")
	require.Contains(t, registry, "func integrationActive(ctx context.Context, client *generated.Client, integrationID string) (bool, error)")
	require.Contains(t, registry, "func selectIngestCandidate(s *Schema, rows []json.RawMessage, payload json.RawMessage) (row json.RawMessage, claimable bool, foreign bool, err error)")
	require.Contains(t, registry, "func (s *Schema) applyIngestClaim(")
	require.Contains(t, registry, "func (s *Schema) applyIngestResurrect(")
	require.Contains(t, registry, "func defaultIngestPersist(s *Schema) IngestPersist")
	require.NotContains(t, registry, "bound bool")
	require.NotContains(t, registry, "BindIngest")
	require.NotContains(t, registry, "TypedIngestPersist")
	require.NotContains(t, registry, "IngestManagesRecord")
	require.Contains(t, registry, "func WithLookupMatches(ctx context.Context, schema *Schema, matches map[int]map[string][]json.RawMessage) context.Context")
	require.Contains(t, registry, "func EncodeLookupKey(alternative LookupAlternative, keys LookupValues) string")

	require.Contains(t, registry, "me := lookupValue(payload, FieldManagedBy)")
	require.Contains(t, registry, "managedByMe := lo.Filter(rows, func(candidate json.RawMessage, _ int) bool")

	require.Contains(t, registry, "if name == s.IntegrationM2MEdge || name == s.IntegrationRunM2MEdge {\n\t\t\treturn false\n\t\t}")
	require.Contains(t, registry, "for _, name := range []string{s.IntegrationM2MEdge, s.IntegrationRunM2MEdge}")
	require.Contains(t, registry, "name == s.IntegrationM2MEdge || name == s.IntegrationRunM2MEdge")

	require.Contains(t, registry, "SchemaWidget.QueryByLookup = func")
	require.Contains(t, registry, "SchemaWidget.SnapshotScope = func")
	require.Contains(t, registry, "widget.ManagedBy(managedBy)")
	require.Contains(t, registry, "SchemaWidget.MarkRemoved = func")
	require.Contains(t, registry, "if errors.Is(err, ErrUpsertStaleRun) {\n\t\treturn nil\n\t}")
	require.Contains(t, registry, "SchemaWidget.Ingest.persist = defaultIngestPersist(SchemaWidget)")
	require.Contains(t, registry, "widget.RemovedAtIsNil()")
	require.Contains(t, registry, "query = query.Where(widget.StandardIDIsNil())")
	require.NotContains(t, registry, "gadget.StandardIDIsNil()")
	require.Contains(t, registry, "SetIntegrationRunID(runID)")
	require.Contains(t, registry, "update = update.AddIntegrationRunIDs(runID)")
	require.Contains(t, registry, `IntegrationRunM2MEdge: "integration_runs",`)
	require.Contains(t, registry, "widget.Or(widget.IntegrationRunIDIsNil(), widget.IntegrationRunIDLT(runID))")
	require.Contains(t, registry, "guarded = true")
	require.Contains(t, registry, "return ErrUpsertStaleRun")
	require.Contains(t, registry, `InstanceScoped: true,`)
	require.Contains(t, registry, `Fields: []string{"external_id"`)

	require.Contains(t, registry, "type SourceFilter struct")
	require.Contains(t, registry, "SchemaWidget.DeleteBySource = func(ctx context.Context, client *generated.Client, filter SourceFilter, cleanup RowCleanup) ([]string, error)")
	require.Contains(t, registry, "widget.SourceInstanceID(filter.InstanceID)")
	require.Contains(t, registry, "query = query.Where(widget.UpdatedAtLTE(*filter.NotUpdatedAfter))")
	require.Contains(t, registry, "client.Widget.Delete().Where(widget.IDIn(chunk...)).Exec(ctx)")

	require.NotContains(t, registry, "SchemaGadget.QueryByLookup")
	require.NotContains(t, registry, "SchemaGadget.SnapshotScope")
	require.NotContains(t, registry, "SchemaGadget.MarkRemoved")
	require.NotContains(t, registry, "SchemaGadget.DeleteBySource")

	gadget := registry[strings.Index(registry, "SchemaGadget = &Schema{"):strings.Index(registry, "// init wires")]
	require.NotContains(t, gadget, "Lookup: []LookupAlternative")
	require.NotContains(t, gadget, "InstanceScoped: true")

	require.Contains(t, registry, "SchemaGadget.Ingest.persist = defaultIngestPersist(SchemaGadget)")

	links := render("entity_links")
	require.Contains(t, links, "func PrefetchLinkTargets(ctx context.Context, client *generated.Client, ownerID string, schema *Schema, payloads []json.RawMessage, links []LinkSpec) (context.Context, error)")

	errorsGo := render("entity_errors")
	require.Contains(t, errorsGo, "ErrUpsertStaleRun")
	require.Contains(t, errorsGo, "ErrLookupAlternativeInvalid")
	require.NotContains(t, errorsGo, "ErrLinkAmbiguous")

	schemaGo := render("entity_schema")
	require.Contains(t, schemaGo, "type LookupAlternative struct")
	require.Contains(t, schemaGo, "type LookupValues map[string]string")
}

// TestOwnerScopeEmission verifies the null-owner scoping predicate is emitted and used only for
// schemas whose owner_id is nullable, while non-nullable owners keep the plain OwnerID predicate
func TestOwnerScopeEmission(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Asset", Snake: "asset", Lower: "asset",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id", SystemScoped: true,
				CreateInputType: "CreateAssetInput", UpdateInputType: "UpdateAssetInput",
				PredicatePackage: "asset", PredicateImport: "example.com/app/ent/generated/asset",
				IntegrationMapped:  true,
				LookupAlternatives: [][]string{{"external_id"}},
				RemovedAtField:     "removed_at",
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
				},
			},
			{
				Name: "Control", Snake: "control", Lower: "control",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateControlInput", UpdateInputType: "UpdateControlInput",
				PredicatePackage: "control", PredicateImport: "example.com/app/ent/generated/control",
				IntegrationMapped:  true,
				LookupAlternatives: [][]string{{"ref_code"}},
				RemovedAtField:     "removed_at",
				ObjectFields: []EntityField{
					{Name: "RefCode", Snake: "ref_code", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "ref_code", LookupKey: true},
				},
			},
			{
				Name: "Risk", Snake: "risk", Lower: "risk",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateRiskInput", UpdateInputType: "UpdateRiskInput",
				PredicatePackage: "risk", PredicateImport: "example.com/app/ent/generated/risk",
				IntegrationMapped:  true,
				LookupAlternatives: [][]string{{"external_id"}},
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
				},
			},
		},
	}

	tmpl, err := parseTemplate("entity_registry")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_registry.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	registry := buf.String()

	require.Contains(t, registry, "OwnerField: asset.FieldOwnerID,")
	require.Contains(t, registry, "OwnerField: control.FieldOwnerID,")
	require.NotContains(t, registry, `"owner_id"`)
	require.Contains(t, registry, "func ownerScopeAsset(ownerID string) predicate.Asset")
	require.Contains(t, registry, "asset.SystemOwned(true)")
	require.NotContains(t, registry, "OwnerIDIsNil()")
	require.Contains(t, registry, "Where(ownerScopeAsset(orgID)).")
	require.Contains(t, registry, "query = query.Where(ownerScopeAsset(ownerID))")
	require.Contains(t, registry, "ownerScopeAsset(ownerID),")
	require.NotContains(t, registry, "Where(asset.OwnerID(orgID)).")
	require.NotContains(t, registry, "query = query.Where(asset.OwnerID(ownerID))")
	require.NotContains(t, registry, "\t\t\t\tasset.OwnerID(ownerID),")

	require.NotContains(t, registry, "ownerScopeRisk")
	require.Contains(t, registry, "query = query.Where(risk.OwnerID(ownerID))")

	require.NotContains(t, registry, "ownerScopeControl")
	require.NotContains(t, registry, "control.OwnerIDIsNil()")
	require.Contains(t, registry, "Where(control.OwnerID(orgID)).")
	require.Contains(t, registry, "query = query.Where(control.OwnerID(ownerID))")
	require.Contains(t, registry, "control.OwnerID(ownerID),")
}

// TestFieldOptional verifies the owner field nullability probe reads the loaded field's Optional flag
func TestFieldOptional(t *testing.T) {
	schema := &load.Schema{Fields: []*load.Field{
		{Name: "owner_id", Optional: true},
		{Name: "name"},
	}}

	require.True(t, fieldOptional(schema, "owner_id"))
	require.False(t, fieldOptional(schema, "name"))
	require.False(t, fieldOptional(schema, "missing"))
	require.False(t, fieldOptional(schema, ""))
}

// TestSchemaOwnerField verifies org-owned schemas derive the owner field from the owner edge and fail without one
func TestSchemaOwnerField(t *testing.T) {
	organization := &gen.Type{Name: "Organization"}
	annotated := gen.Annotations{entx.OrgOwnedSchemaName: map[string]any{}}

	t.Run("not org owned", func(t *testing.T) {
		field, err := schemaOwnerField(&gen.Type{Name: "Widget", Edges: []*gen.Edge{{Name: "owner", Unique: true, Type: organization, Rel: gen.Relation{Type: gen.M2O, Columns: []string{"owner_id"}}}}})
		require.NoError(t, err)
		require.Empty(t, field)
	})

	t.Run("owner edge missing", func(t *testing.T) {
		_, err := schemaOwnerField(&gen.Type{Name: "Widget", Annotations: annotated})
		require.ErrorIs(t, err, ErrOwnerEdgeMissing)
	})

	t.Run("owner edge without a foreign-key field", func(t *testing.T) {
		_, err := schemaOwnerField(&gen.Type{Name: "Widget", Annotations: annotated, Edges: []*gen.Edge{{Name: "owner", Type: organization, Rel: gen.Relation{Type: gen.O2M}}}})
		require.ErrorIs(t, err, ErrOwnerEdgeMissing)
	})

	t.Run("owner edge owning its foreign key", func(t *testing.T) {
		field, err := schemaOwnerField(&gen.Type{Name: "Widget", Annotations: annotated, Edges: []*gen.Edge{{Name: "owner", Unique: true, Type: organization, Rel: gen.Relation{Type: gen.M2O, Columns: []string{"owner_id"}}}}})
		require.NoError(t, err)
		require.Equal(t, "owner_id", field)
	})
}

// TestFieldMatchKey verifies only plain string columns qualify as match keys
func TestFieldMatchKey(t *testing.T) {
	require.True(t, fieldMatchKey(&gen.Field{Name: "external_id", Type: &entfield.TypeInfo{Type: entfield.TypeString}}))
	require.False(t, fieldMatchKey(&gen.Field{Name: "domains", Type: &entfield.TypeInfo{Type: entfield.TypeJSON, Ident: "[]string"}}))
	require.False(t, fieldMatchKey(&gen.Field{Name: "metadata", Type: &entfield.TypeInfo{Type: entfield.TypeJSON, Ident: "map[string]any"}}))
	require.False(t, fieldMatchKey(&gen.Field{Name: "count", Type: &entfield.TypeInfo{Type: entfield.TypeInt}}))
	require.False(t, fieldMatchKey(&gen.Field{Name: "untyped"}))
}

// catalogEmissionData builds the registry test data for a catalog schema (Widget) beside a plain schema (Gadget)
func catalogEmissionData() EntityData {
	return EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Gadget", Snake: "gadget", Lower: "gadget",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id",
				CreateInputType: "CreateGadgetInput", UpdateInputType: "UpdateGadgetInput",
				PredicatePackage: "gadget", PredicateImport: "example.com/app/ent/generated/gadget",
			},
			{
				Name: "Widget", Snake: "widget", Lower: "widget",
				HasCreate: true, HasUpdate: true, OwnerField: "owner_id", SystemScoped: true,
				CreateInputType: "CreateWidgetInput", UpdateInputType: "UpdateWidgetInput",
				PredicatePackage: "widget", PredicateImport: "example.com/app/ent/generated/widget",
				IntegrationMapped:  true,
				CatalogPointer:     "catalog_id",
				CatalogFields:      []string{"name", "description"},
				HasCatalog:         true,
				LookupAlternatives: [][]string{{"external_id"}, {"name"}, {"domains"}},
				ObjectFields: []EntityField{
					{Name: "CatalogID", Snake: "catalog_id", Type: "string", MatchKey: true},
					{Name: "Description", Snake: "description", Type: "string", SourceManaged: true},
					{Name: "Domains", Snake: "domains", Type: "[]string", MatchKey: true, IntegrationMapped: true, InputKey: "domains", SourceManaged: true, CaseInsensitive: true},
					{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
					{Name: "ExternallyVisible", Snake: "externally_visible", Type: "bool"},
					{Name: "Name", Snake: "name", Type: "string", IntegrationMapped: true, InputKey: "name", SourceManaged: true, CaseInsensitive: true},
				},
				Edges: []EntityEdge{
					{Name: "catalog", TargetSchema: "Widget", TargetInRegistry: true, Unique: true, Field: "catalog_id"},
				},
			},
		},
	}
}

// renderTemplate renders one template for data and asserts the output parses as Go
func renderTemplate(t *testing.T, name string, data EntityData) string {
	t.Helper()

	tmpl, err := parseTemplate(name)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	return buf.String()
}

// TestCatalogEmission verifies the catalog constants, closures, Reconcile, refresh, unlink and listeners are emitted only for schemas with HasCatalog
func TestCatalogEmission(t *testing.T) {
	data := catalogEmissionData()
	registry := renderTemplate(t, "entity_registry", data)

	require.Contains(t, registry, `const CatalogPointerField = "catalog_id"`)
	require.Contains(t, registry, `const CatalogVisibilityField = "externally_visible"`)
	require.Contains(t, registry, "func (s *Schema) Reconcile(ctx context.Context, client *generated.Client, ownerID string, payload json.RawMessage, links ...LinkSpec) (string, error)")
	require.Contains(t, registry, "func (s *Schema) RefreshAdopted(ctx context.Context, client *generated.Client, catalogID string) (int, error)")
	require.Contains(t, registry, "func (s *Schema) Unlink(ctx context.Context, client *generated.Client, id string) error")
	require.Contains(t, registry, "func (s *Schema) CatalogFields() []string")
	require.Contains(t, registry, "return lo.FilterMap(s.Fields, func(f FieldDescriptor, _ int) (string, bool) { return f.Name, f.SourceManaged })")
	require.Contains(t, registry, `"github.com/theopenlane/iam/auth"`)

	for _, removed := range []string{"type MatchScope int", "type Ambiguity int", "type MatchTier struct", "type Resolution struct", "type Outcome int", "func (s *Schema) Adopt(", "func (s *Schema) Resolve(", "enrichRow", "fillEmpty(", "matchTierIDs", "matchTiers", "CatalogHandle", "MatchCandidate", "func (s *Schema) Match(", "MatchCatalog", "RelinkAdopted", "relink", "relinkCreated", "FillProvenance", "PointerField string", "VisibilityField string", "KeyField string", "SyncedField string", "TypeEdge string", ".Catalog.Fields", ".Catalog.PointerField", "CatalogPointerField()", "CatalogVisibilityField()", "CatalogSyncedField()", "CatalogKeyField", "lookupScope", "type lookupInput", "Catalog.row", "Catalog.lookup", "Catalog.pointed", "Catalog.visible", "catalogScope", "catalogRowWidget", "reconcileLookup", "alternativeInput", "catalogTypeName", "mergeNamedCatalog", "stampCatalogLink", "dropCatalogStamps", "syncedSnapshot", "refreshedValue", "snapshotEqual", "jsonEquivalent", "likeEscaper", "lookupKeyFor(", "ErrResolveAmbiguous", "ErrAdoptConflict", "ErrCatalogLinkedElsewhere", "Inferred", "Suffix: true", "CatalogSynced", "CatalogTypeEdge"} {
		require.NotContains(t, registry, removed)
	}

	capability := registry[strings.Index(registry, "type CatalogCapability struct"):strings.Index(registry, "func (s *Schema) CatalogFields()")]
	require.Contains(t, capability, "copies func(")
	require.Contains(t, capability, "unlink func(")
	require.Contains(t, capability, "unlinkAll func(")
	require.NotContains(t, capability, "Fields []string")

	widget := registry[strings.Index(registry, "SchemaWidget = &Schema{"):strings.Index(registry, "// init wires")]
	require.Contains(t, widget, `Catalog: &CatalogCapability{},`)
	require.Contains(t, registry, `{Name: "catalog_id", Label: "CatalogID", Type: "string", MatchKey: true},`)
	require.Contains(t, registry, `{Name: "externally_visible", Label: "ExternallyVisible", Type: "bool"},`)
	require.Contains(t, registry, `{Name: "name", Label: "Name", Type: "string", InputKey: "name", SourceManaged: true, CaseInsensitive: true},`)
	require.Contains(t, registry, `{Name: "domains", Label: "Domains", Type: "[]string", MatchKey: true, InputKey: "domains", SourceManaged: true, CaseInsensitive: true},`)
	require.Contains(t, registry, `{Name: "description", Label: "Description", Type: "string", SourceManaged: true},`)
	require.NotContains(t, registry, "LookupKey: true")
	require.Contains(t, widget, `{Fields: []string{"external_id"}},`)
	require.Contains(t, widget, `{Fields: []string{"name"}},`)
	require.Contains(t, widget, `{Fields: []string{"domains"}},`)
	require.Contains(t, widget, "Create: func")
	require.Contains(t, widget, "Update: func")
	require.Equal(t, 1, strings.Count(widget, `applyStampedFields(builder.Mutation(), input, CatalogPointerField)`))

	gadget := registry[strings.Index(registry, "SchemaGadget = &Schema{"):strings.Index(registry, "SchemaWidget = &Schema{")]
	require.NotContains(t, gadget, "Catalog:")
	require.NotContains(t, gadget, "Create: func")
	require.NotContains(t, gadget, "Update: func")
	require.NotContains(t, gadget, "applyStampedFields")

	require.Contains(t, registry, "SchemaWidget.QueryByLookup = func")
	require.Contains(t, registry, "return filterByAlternativeFields(SchemaWidget, results, fields, keys), nil")
	require.Contains(t, registry, "SchemaWidget.Catalog.copies = func(ctx context.Context, client *generated.Client, catalogID string) ([]json.RawMessage, error)")
	require.Contains(t, registry, "client.Widget.Query().Where(widget.CatalogID(catalogID)).All(ctx)")
	require.Contains(t, registry, "SchemaWidget.Catalog.unlink = func(ctx context.Context, client *generated.Client, id string) error")
	require.Contains(t, registry, "client.Widget.UpdateOneID(id).ClearCatalogID().Exec(ctx)")
	require.Contains(t, registry, "SchemaWidget.Catalog.unlinkAll = func")
	require.Contains(t, registry, "client.Widget.Update().Where(widget.CatalogID(catalogID)).ClearCatalogID().Save(ctx)")
	require.NotContains(t, registry, "SchemaGadget.Catalog")

	require.Contains(t, registry, "func CatalogListeners() []gala.Registration")
	require.Contains(t, registry, "Operations: []string{OpUpdate, OpUpdateOne},")
	require.Contains(t, registry, `Fields:     []string{"name", "description"},`)
	require.Contains(t, registry, "Handle:     catalogRefreshHandler(SchemaWidget),")
	require.Contains(t, registry, "Operations: []string{OpUpdate, OpUpdateOne, OpSoftDelete},")
	require.Contains(t, registry, "Fields:     []string{CatalogVisibilityField},")
	require.Contains(t, registry, "Handle:     catalogUnlinkHandler(SchemaWidget),")
	require.Contains(t, registry, "restored.WithCapabilities(auth.CapInternalOperation)")
	require.Contains(t, registry, "updated, err := s.RefreshAdopted(inv.Context, inv.Client, inv.EntityID)")
	require.Contains(t, registry, "if payload.Operation != OpSoftDelete {")
	require.Contains(t, registry, "visible, ok := payload.StringValue(CatalogVisibilityField)")
	require.Contains(t, registry, "unlinked, err := s.Catalog.unlinkAll(inv.Context, inv.Client, inv.EntityID)")
	require.NotContains(t, registry, "Schema:     SchemaGadget,")

	edges := registry[strings.Index(registry, "SchemaWidget.Edges = []EdgeDescriptor{"):strings.Index(registry, "SchemaWidget.Catalog.copies = ")]
	require.Contains(t, edges, "Name: \"catalog\",")

	index := renderTemplate(t, "entity_integration", data)
	require.Contains(t, index, `Name: FieldDescriptor{Name: "name", Label: "Name", Type: "string", InputKey: "name", SourceManaged: true, CaseInsensitive: true},`)
	require.NotContains(t, index, "Description")

	errorsGo := renderTemplate(t, "entity_errors", data)
	require.Contains(t, errorsGo, "ErrCatalogUnsupported")
	require.Contains(t, errorsGo, "ErrCatalogRowNotFound")
	require.Contains(t, errorsGo, `refOpReconcile = "reconcile"`)

	for _, removed := range []string{"ErrCatalogRowNotSystemOwned", "ErrCatalogRowNotVisible", "ErrResolveAmbiguous", "ErrAdoptConflict", "ErrCatalogLinkedElsewhere", "ErrResolveFieldUnsupported", "ErrCatalogMatchFieldUnsupported"} {
		require.NotContains(t, errorsGo, removed)
	}

	schemaGo := renderTemplate(t, "entity_schema", data)

	for _, removed := range []string{"Inferred bool", "Suffix bool", "CatalogPointer bool", "CatalogVisibility bool", "CatalogSynced bool", "CatalogTypeEdge bool"} {
		require.NotContains(t, schemaGo, removed)
	}

	plainData := catalogEmissionData()
	plainData.Schemas = plainData.Schemas[:1]
	plain := renderTemplate(t, "entity_registry", plainData)
	require.NotContains(t, plain, "func (s *Schema) Reconcile(")
	require.NotContains(t, plain, "iam/auth")
}

// TestReconcileEmission verifies the shared lookup resolution, the catalog merge, fill-only hits, adoption on create and the unconditional refresh
func TestReconcileEmission(t *testing.T) {
	registry := renderTemplate(t, "entity_registry", catalogEmissionData())

	require.Contains(t, registry, "catalogID := lookupValue(payload, CatalogPointerField)\n\tadopting := catalogID != \"\"")
	require.Contains(t, registry, "catalogRow, err := s.catalogRow(ctx, client, ref, catalogID)")
	require.Contains(t, registry, "if payload, err = s.mergeCatalogRow(ctx, ref, payload, catalogRow); err != nil {")
	require.Contains(t, registry, "row, err := s.resolveLookup(ctx, client, ownerID, payload)")
	require.Contains(t, registry, "if row == nil && !adopting {")
	require.Contains(t, registry, "catalogRow, err := s.resolveLookup(ctx, client, \"\", payload)")
	require.Contains(t, registry, "if row, err = s.resolveLookup(ctx, client, ownerID, payload); err != nil {")
	require.Contains(t, registry, "return entityID(row), s.reconcileHit(ctx, client, ownerID, row, payload, links)")
	require.Contains(t, registry, "return s.reconcileCreate(ctx, client, ref, ownerID, payload, adopting, links)")

	require.Contains(t, registry, "func (s *Schema) resolveLookup(ctx context.Context, client *generated.Client, ownerID string, payload json.RawMessage) (json.RawMessage, error)")
	require.Contains(t, registry, "for _, key := range lookupKeysFor(s, payload) {")
	require.Contains(t, registry, "rows, err := s.QueryByLookup(ctx, client, ownerID, key.alternative, key.values)")
	require.Contains(t, registry, "rows = lo.Filter(rows, func(row json.RawMessage, _ int) bool { return catalogVisible(row) })")
	require.Contains(t, registry, "case len(rows) > 1:")
	require.Contains(t, registry, "func (s *Schema) catalogRow(ctx context.Context, client *generated.Client, ref SchemaRef, catalogID string) (json.RawMessage, error)")
	require.Contains(t, registry, "case generated.IsNotFound(err), err == nil && !catalogVisible(row):")
	require.Contains(t, registry, "logError(ctx, ref, ErrCatalogRowNotFound,")
	require.Contains(t, registry, "func catalogVisible(row json.RawMessage) bool")
	require.Contains(t, registry, "if merged[CatalogPointerField], err = json.Marshal(entityID(catalogRow)); err != nil {")

	require.Contains(t, registry, "mutation, save, err := s.Ingest.buildUpdate(ctx, client, row, s.rekeyEdgesForUpdate(dropCatalogPointer(linked)))")
	require.Contains(t, registry, "if err := s.fillMutation(ctx, mutation); err != nil {")
	require.Contains(t, registry, "if ChangeSetFromMutation(mutation).Empty() {")
	require.Contains(t, registry, "if !adopting {\n\t\tlinked = dropCatalogPointer(linked)\n\t}")
	require.Contains(t, registry, "return s.Create(ctx, client, created)")

	require.Contains(t, registry, "func (s *Schema) fillMutation(ctx context.Context, mutation ent.Mutation) error")
	require.Contains(t, registry, `if !ok || field.Type == "bool" {`)
	require.Contains(t, registry, "old, err := mutation.OldField(ctx, name)")
	require.Contains(t, registry, "merged, grew := unionStringValues(old, proposed)")
	require.Contains(t, registry, "for _, name := range mutation.ClearedFields() {")

	require.Contains(t, registry, "func lookupPredicate(schema *Schema, field string, values []string) func(*sql.Selector)")
	require.Contains(t, registry, "case descriptor.CaseInsensitive && array:")
	require.Contains(t, registry, "return sql.EqualFold(s.C(field), value)")
	require.Contains(t, registry, "jsonArrayMatch(field, lo.Map(values, func(value string, _ int) string { return strings.ToLower(value) }))(s)")
	require.Contains(t, registry, "return lookupPredicate(schema, field, values)")
	require.Contains(t, registry, `WriteString(") AS v WHERE LOWER(v) IN (").Args(lo.ToAnySlice(values)...).WriteString("))")`)
	require.NotContains(t, registry, "LIKE")

	require.Contains(t, registry, "func lookupKeysFor(s *Schema, payload json.RawMessage) []lookupKey")
	require.Contains(t, registry, "func lookupFieldValues(payload json.RawMessage, field FieldDescriptor) []string")
	require.Contains(t, registry, "func lookupFieldMatches(s *Schema, row json.RawMessage, field, value string) bool")
	require.Contains(t, registry, "candidates := lookupKeysFor(s, payload)")
	require.Contains(t, registry, "if rows, err = s.lookupRows(ctx, client, ownerID, key, cache); err != nil {")
	require.Contains(t, registry, "cacheKey := candidates[0].cacheKey(s)")

	require.Contains(t, registry, "if !catalogVisible(row) {\n\t\treturn 0, nil\n\t}")
	require.Contains(t, registry, "payload, err := catalogPayload(row, s.CatalogFields())")
	require.Contains(t, registry, "if err := s.Update(ctx, client, entityID(adopted), payload); err != nil {")

	require.Contains(t, registry, "func (s *Schema) applyIngestFill(ctx context.Context, client *generated.Client, row json.RawMessage, payload json.RawMessage) (changed bool, managed bool, err error)")
	require.Contains(t, registry, `if ownerID != "" && lookupValue(row, FieldManagedBy) == "" && lookupValue(payload, FieldManagedBy) != "" {`)
	require.Contains(t, registry, "changed, managed, err = s.applyIngestUpdate(ctx, client, ownerID, candidate, payload)")
	require.Contains(t, registry, "return !changes.Empty() || len(throughIDs) > 0, false, nil")
}

// fieldAnnotations keys schema annotations by name the way ent loads them onto a gen.Field
func fieldAnnotations(annotations ...schema.Annotation) gen.Annotations {
	result := gen.Annotations{}

	for _, annotation := range annotations {
		result[annotation.Name()] = annotation
	}

	return result
}

// TestFieldCatalogManagedDirective verifies the directive probe reads the merged entgql annotation
func TestFieldCatalogManagedDirective(t *testing.T) {
	require.True(t, fieldCatalogManagedDirective(&gen.Field{Name: "name", Annotations: fieldAnnotations(entx.CatalogManagedField()...)}))
	require.False(t, fieldCatalogManagedDirective(&gen.Field{Name: "name", Annotations: fieldAnnotations(entx.FieldSourceManaged())}))
	require.False(t, fieldCatalogManagedDirective(&gen.Field{Name: "name", Annotations: fieldAnnotations(entgql.Directives(entgql.NewDirective("other")))}))
	require.False(t, fieldCatalogManagedDirective(&gen.Field{Name: "name"}))
}

// TestValidateCatalogManaged verifies source-managed fields on a catalog schema must carry the catalogManaged directive
func TestValidateCatalogManaged(t *testing.T) {
	require.NoError(t, validateCatalogManaged(EntitySchema{Name: "Widget"}, []string{"name"}))
	require.NoError(t, validateCatalogManaged(EntitySchema{Name: "Widget", CatalogPointer: "catalog_widget_id"}, nil))

	err := validateCatalogManaged(EntitySchema{Name: "Widget", CatalogPointer: "catalog_widget_id"}, []string{"name", "domains"})
	require.ErrorIs(t, err, ErrCatalogManagedDirectiveMissing)
	require.ErrorContains(t, err, "Widget.name, Widget.domains")
}

// TestEdgeCatalogPointer verifies the catalog edge must be a unique self edge owning its foreign key
func TestEdgeCatalogPointer(t *testing.T) {
	node := &gen.Type{Name: "Widget"}
	annotated := gen.Annotations{entx.CatalogEdgeAnnotationName: map[string]any{}}

	t.Run("no annotation", func(t *testing.T) {
		pointer, err := edgeCatalogPointer(node, &gen.Edge{Name: "parent", Unique: true, Type: node})
		require.NoError(t, err)
		require.Empty(t, pointer)
	})

	t.Run("non-unique edge", func(t *testing.T) {
		_, err := edgeCatalogPointer(node, &gen.Edge{Name: "catalog_widgets", Type: node, Annotations: annotated})
		require.ErrorIs(t, err, ErrCatalogEdgeInvalid)
	})

	t.Run("edge to another schema", func(t *testing.T) {
		_, err := edgeCatalogPointer(node, &gen.Edge{Name: "catalog_gadget", Unique: true, Type: &gen.Type{Name: "Gadget"}, Annotations: annotated})
		require.ErrorIs(t, err, ErrCatalogEdgeInvalid)
	})

	t.Run("edge without a foreign-key field", func(t *testing.T) {
		_, err := edgeCatalogPointer(node, &gen.Edge{Name: "catalog_widget", Unique: true, Type: node, Annotations: annotated})
		require.ErrorIs(t, err, ErrCatalogEdgeInvalid)
	})

	t.Run("unique self edge owning its foreign key", func(t *testing.T) {
		edge := &gen.Edge{
			Name: "catalog_widget", Unique: true, Type: node, Annotations: annotated,
			Rel: gen.Relation{Type: gen.M2O, Columns: []string{"catalog_widget_id"}},
		}

		pointer, err := edgeCatalogPointer(node, edge)
		require.NoError(t, err)
		require.Equal(t, "catalog_widget_id", pointer)
	})
}

// TestValidateCatalog verifies a catalog schema must carry both inputs, the visibility and system_owned columns, a nullable owner, copied fields and a lookup alternative
func TestValidateCatalog(t *testing.T) {
	complete := EntitySchema{
		Name: "Widget", CatalogPointer: "catalog_id", HasCatalog: true, HasCreate: true, HasUpdate: true,
		OwnerField: "owner_id", SystemScoped: true, CatalogFields: []string{"name"}, LookupAlternatives: [][]string{{"external_id"}},
	}

	require.NoError(t, validateCatalog(complete, true, true))
	require.NoError(t, validateCatalog(EntitySchema{Name: "Widget"}, false, false))

	cases := []struct {
		name        string
		part        string
		mutate      func(*EntitySchema)
		systemOwned bool
		visible     bool
	}{
		{"missing create", "needs create and update inputs", func(s *EntitySchema) { s.HasCreate = false }, true, true},
		{"missing update", "needs create and update inputs", func(s *EntitySchema) { s.HasUpdate = false }, true, true},
		{"missing visibility", "needs the externally_visible field", func(*EntitySchema) {}, true, false},
		{"missing system_owned", "needs the system_owned field", func(*EntitySchema) {}, false, true},
		{"missing owner", "needs an org owner edge", func(s *EntitySchema) { s.OwnerField = "" }, true, true},
		{"non-nullable owner", "needs a nullable owner field", func(s *EntitySchema) { s.SystemScoped = false }, true, true},
		{"missing fields", "needs a source-managed field", func(s *EntitySchema) { s.CatalogFields = nil }, true, true},
		{"missing lookup", "needs a lookup alternative", func(s *EntitySchema) { s.LookupAlternatives = nil }, true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			schema := complete
			tc.mutate(&schema)

			err := validateCatalog(schema, tc.systemOwned, tc.visible)
			require.ErrorIs(t, err, ErrCatalogIncomplete)
			require.ErrorContains(t, err, tc.part)
		})
	}
}

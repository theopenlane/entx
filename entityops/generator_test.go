package entityops

import (
	"bytes"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/require"
)

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
			raw, err := _templates.ReadFile("templates/" + name + ".tpl")
			require.NoError(t, err)

			tmpl, err := template.New(name).Funcs(gen.Funcs).Parse(string(raw))
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
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateAssetInput", UpdateInputType: "UpdateAssetInput",
				PredicatePackage: "asset", PredicateImport: "example.com/app/ent/generated/asset",
				IntegrationMapped: true,
				ObjectFields:      []EntityField{{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true}},
				Edges:             []EntityEdge{{Name: "controls", TargetSchema: "Control", TargetInRegistry: true}},
			},
			{
				Name: "Control", Snake: "control", Lower: "control",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateControlInput", UpdateInputType: "UpdateControlInput",
				PredicatePackage: "control", PredicateImport: "example.com/app/ent/generated/control",
				LinkTarget:       true,
				WorkflowEligible: true,
				ObjectFields:     []EntityField{{Name: "RefCode", Snake: "ref_code", Type: "string", MatchKey: true, Projectable: true, WorkflowEligible: true}},
			},
			{
				Name: "Note", Snake: "note", Lower: "note",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateNoteInput", UpdateInputType: "UpdateNoteInput",
				PredicatePackage: "note", PredicateImport: "example.com/app/ent/generated/note",
				ObjectFields: []EntityField{{Name: "Text", Snake: "text", Type: "string", Projectable: true}},
			},
		},
	}

	render := func(name string) string {
		raw, err := _templates.ReadFile("templates/" + name + ".tpl")
		require.NoError(t, err)

		tmpl, err := template.New(name).Funcs(gen.Funcs).Parse(string(raw))
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

	raw, err := _templates.ReadFile("templates/entity_schema.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_schema").Funcs(gen.Funcs).Parse(string(raw))
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

	raw, err = _templates.ReadFile("templates/entity_integration.tpl")
	require.NoError(t, err)

	tmpl, err = template.New("entity_integration").Funcs(gen.Funcs).Parse(string(raw))
	require.NoError(t, err)

	buf.Reset()
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_integration.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	index := buf.String()
	require.Contains(t, index, `ExternalID: FieldDescriptor{Name: "external_id", InputKey: "external_id"}`)
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
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
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

	raw, err = _templates.ReadFile("templates/entity_registry.tpl")
	require.NoError(t, err)

	tmpl, err = template.New("entity_registry").Funcs(gen.Funcs).Parse(string(raw))
	require.NoError(t, err)

	buf.Reset()
	require.NoError(t, tmpl.Execute(&buf, registryData))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_registry.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	registry := buf.String()
	require.Contains(t, registry, "CaseInsensitive: true")
	require.Contains(t, registry, `s.FieldByName(FieldOwnerID)`)
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
	raw, err := _templates.ReadFile("templates/entity_metadata.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_metadata").Funcs(gen.Funcs).Parse(string(raw))
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
	raw, err := _templates.ReadFile("templates/entity_metadata.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_metadata").Funcs(gen.Funcs).Parse(string(raw))
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
		schema := EntitySchema{
			Name: "Widget", IntegrationMapped: true, ObjectFields: mapped,
			LookupAlternatives: [][]string{{"external_id", "tenant_id"}},
		}
		require.NoError(t, validateLookupAlternatives(schema))
	})

	t.Run("unknown field", func(t *testing.T) {
		schema := EntitySchema{
			Name: "Widget", IntegrationMapped: true, ObjectFields: mapped,
			LookupAlternatives: [][]string{{"does_not_exist"}},
		}
		require.ErrorIs(t, validateLookupAlternatives(schema), ErrLookupAlternativeFieldUnknown)
	})

	t.Run("unmapped field", func(t *testing.T) {
		schema := EntitySchema{
			Name: "Widget", IntegrationMapped: true, ObjectFields: mapped,
			LookupAlternatives: [][]string{{"internal_note"}},
		}
		require.ErrorIs(t, validateLookupAlternatives(schema), ErrLookupAlternativeFieldUnknown)
	})

	t.Run("without integration mapping", func(t *testing.T) {
		schema := EntitySchema{
			Name: "Widget", IntegrationMapped: false, ObjectFields: mapped,
			LookupAlternatives: [][]string{{"external_id"}},
		}
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
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateWidgetInput", UpdateInputType: "UpdateWidgetInput",
				PredicatePackage: "widget", PredicateImport: "example.com/app/ent/generated/widget",
				IntegrationMapped:     true,
				InstanceScoped:        true,
				LookupAlternatives:    [][]string{{"external_id"}},
				RemovedAtField:        "removed_at",
				RemovedAtEpisodic:     true,
				HasIntegrationRunID:   true,
				IntegrationRunM2MEdge: "integration_runs",
				ObjectFields: []EntityField{
					{Name: "ExternalID", Snake: "external_id", Type: "string", IntegrationMapped: true, InputKey: "external_id", LookupKey: true},
				},
			},
			{
				Name: "Gadget", Snake: "gadget", Lower: "gadget",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateGadgetInput", UpdateInputType: "UpdateGadgetInput",
				PredicatePackage: "gadget", PredicateImport: "example.com/app/ent/generated/gadget",
				IntegrationMapped: true,
			},
		},
	}

	render := func(name string) string {
		raw, err := _templates.ReadFile("templates/" + name + ".tpl")
		require.NoError(t, err)

		tmpl, err := template.New(name).Funcs(gen.Funcs).Parse(string(raw))
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
	require.Contains(t, registry, "SetIntegrationRunID(runID)")
	require.Contains(t, registry, "update = update.AddIntegrationRunIDs(runID)")
	require.Contains(t, registry, `IntegrationRunM2MEdge: "integration_runs",`)
	require.Contains(t, registry, "widget.Or(widget.IntegrationRunIDIsNil(), widget.IntegrationRunIDLT(runID))")
	require.Contains(t, registry, "guarded = true")
	require.Contains(t, registry, "return ErrUpsertStaleRun")
	require.Contains(t, registry, `InstanceScoped: true,`)
	require.Contains(t, registry, `Fields: []string{"external_id"`)

	require.NotContains(t, registry, "SchemaGadget.QueryByLookup")
	require.NotContains(t, registry, "SchemaGadget.SnapshotScope")
	require.NotContains(t, registry, "SchemaGadget.MarkRemoved")

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

// TestFillProvenanceTemplateEmission verifies FillProvenance emission per link kind
func TestFillProvenanceTemplateEmission(t *testing.T) {
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
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateWidgetInput", UpdateInputType: "UpdateWidgetInput",
				PredicatePackage: "widget", PredicateImport: "example.com/app/ent/generated/widget",
				IntegrationMapped:  true,
				HasIntegrationID:   true,
				IntegrationFKField: "integration_id",
			},
			{
				Name: "Gizmo", Snake: "gizmo", Lower: "gizmo",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateGizmoInput", UpdateInputType: "UpdateGizmoInput",
				PredicatePackage: "gizmo", PredicateImport: "example.com/app/ent/generated/gizmo",
				IntegrationMapped:  true,
				IntegrationM2MEdge: "integrations",
			},
			{
				Name: "Gadget", Snake: "gadget", Lower: "gadget",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateGadgetInput", UpdateInputType: "UpdateGadgetInput",
				PredicatePackage: "gadget", PredicateImport: "example.com/app/ent/generated/gadget",
				IntegrationMapped: true,
			},
		},
	}

	raw, err := _templates.ReadFile("templates/entity_registry.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_registry").Funcs(gen.Funcs).Parse(string(raw))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	_, err = parser.ParseFile(token.NewFileSet(), "entity_registry.go", buf.Bytes(), parser.AllErrors)
	require.NoError(t, err, buf.String())

	registry := buf.String()

	require.Contains(t, registry, "FillProvenance func(ctx context.Context, client *generated.Client, installation *generated.Integration) (int, error)")

	require.Contains(t, registry, "SchemaWidget.FillProvenance = func")
	require.Contains(t, registry, "widget.IntegrationID(installation.ID),")
	require.Contains(t, registry, "widget.Or(widget.SourceDefinitionIDIsNil(), widget.SourceDefinitionID(\"\"), widget.SourceInstanceIDIsNil(), widget.SourceInstanceID(\"\"))")
	require.Contains(t, registry, "widget.Or(widget.ManagedByIsNil(), widget.ManagedBy(\"\"), widget.ManagedBy(installation.ID))")
	require.Contains(t, registry, "widget.Or(widget.SourceInstanceIDIsNil(), widget.SourceInstanceIDNEQ(instanceID))")
	require.Contains(t, registry, "update = update.SetSourceDefinitionVersion(installation.DefinitionVersion)")
	require.Contains(t, registry, "update = update.SetSourceInstanceID(instanceID)")

	require.NotContains(t, registry, "widget.HasIntegrationsWith")

	require.Contains(t, registry, "SchemaGizmo.FillProvenance = func")
	require.Contains(t, registry, "gizmo.HasIntegrationsWith(integration.ID(installation.ID)),")
	require.Contains(t, registry, "gizmo.Not(gizmo.HasIntegrationsWith(integration.IDNEQ(installation.ID))),")
	require.NotContains(t, registry, "gizmo.IntegrationID(")

	require.NotContains(t, registry, "SchemaGadget.FillProvenance")
}

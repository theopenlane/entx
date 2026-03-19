package integrationmapping

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/entx"
)

func TestBuildMappingTemplate_GeneratesTypedIngestContracts(t *testing.T) {
	tmpl := buildMappingTemplate()

	data := MappingData{
		PackageName:             "integrationgenerated",
		EntPackage:              "github.com/example/project/internal/ent/generated",
		GalaPackage:             "github.com/example/project/pkg/gala",
		GenerateIngestContracts: true,
		Schemas: []MappingSchema{
			{
				Name:                  "DirectoryAccount",
				ConstName:             "IntegrationMappingSchemaDirectoryAccount",
				IngestTopic:           "integration.ingest.directory_account.requested",
				IngestRequestTypeName: "IntegrationIngestDirectoryAccountRequested",
				IngestTopicVarName:    "IntegrationIngestDirectoryAccountRequestedTopic",
				InputTypeName:         "CreateDirectoryAccountInput",
				Fields: []MappingField{
					{
						InputKey:  "externalID",
						ConstName: "IntegrationMappingDirectoryAccountExternalID",
						EntField:  "external_id",
						Type:      "string",
						Required:  true,
						UpsertKey: true,
					},
				},
				RuntimeDefaults: []IngestRuntimeDefault{
					{
						Field:             "owner_id",
						GoField:           "OwnerID",
						IntegrationField: "OwnerID",
					},
				},
				StockPersist: true,
			},
		},
	}

	var out bytes.Buffer
	err := tmpl.Execute(&out, data)
	assert.NoError(t, err)

	generated := out.String()
	assert.Contains(t, generated, `generated "github.com/example/project/internal/ent/generated"`)
	assert.Contains(t, generated, `"github.com/example/project/pkg/gala"`)
	assert.Contains(t, generated, "type IntegrationIngestMetadata struct")
	assert.Contains(t, generated, "IntegrationID string `json:\"integrationId\"`")
	assert.Contains(t, generated, "type IntegrationIngestDirectoryAccountRequested struct")
	assert.Contains(t, generated, "Input generated.CreateDirectoryAccountInput")
	assert.Contains(t, generated, "var IntegrationIngestDirectoryAccountRequestedTopic = gala.Topic[IntegrationIngestDirectoryAccountRequested]")
	assert.Contains(t, generated, `Name: "integration.ingest.directory_account.requested"`)
	assert.NotContains(t, generated, "IntegrationIngestTopicDirectoryAccountRequested")
	assert.Contains(t, generated, "PrepareDirectoryAccountInput")
	assert.NotContains(t, generated, "DefaultOperation")
}

func TestBuildMappingTemplate_OmitsTypedIngestContractsWhenDisabled(t *testing.T) {
	tmpl := buildMappingTemplate()

	data := MappingData{
		PackageName: "integrationgenerated",
		Schemas: []MappingSchema{
			{
				Name:        "Vulnerability",
				ConstName:   "IntegrationMappingSchemaVulnerability",
				IngestTopic: "integration.ingest.vulnerability.requested",
			},
		},
	}

	var out bytes.Buffer
	err := tmpl.Execute(&out, data)
	assert.NoError(t, err)

	generated := out.String()
	assert.NotContains(t, generated, "type IntegrationIngestMetadata struct")
	assert.NotContains(t, generated, "gala.Topic[")
	assert.False(t, strings.Contains(generated, "github.com/example/project/pkg/gala"))
	assert.NotContains(t, generated, "DefaultOperation")
}

func TestBuildMappingTemplate_FromIntegrationGeneratesPrepareFunc(t *testing.T) {
	tmpl := buildMappingTemplate()

	data := MappingData{
		PackageName:             "integrationgenerated",
		EntPackage:              "github.com/example/project/internal/ent/generated",
		GalaPackage:             "github.com/example/project/pkg/gala",
		GenerateIngestContracts: true,
		Schemas: []MappingSchema{
			{
				Name:                  "DirectoryGroup",
				ConstName:             "IntegrationMappingSchemaDirectoryGroup",
				IngestTopic:           "integration.ingest.directory_group.requested",
				IngestRequestTypeName: "IntegrationIngestDirectoryGroupRequested",
				IngestTopicVarName:    "IntegrationIngestDirectoryGroupRequestedTopic",
				InputTypeName:         "CreateDirectoryGroupInput",
				StockPersist:          true,
				RuntimeDefaults: []IngestRuntimeDefault{
					{Field: "owner_id", GoField: "OwnerID", IntegrationField: "OwnerID"},
					{Field: "integration_id", GoField: "IntegrationID", IntegrationField: "ID"},
				},
			},
		},
	}

	var out bytes.Buffer
	err := tmpl.Execute(&out, data)
	assert.NoError(t, err)

	generated := out.String()
	assert.Contains(t, generated, "PrepareDirectoryGroupInput")
	assert.Contains(t, generated, "integration.OwnerID")
	assert.Contains(t, generated, "integration.ID")
	assert.Contains(t, generated, "StockPersist: true")
	assert.NotContains(t, generated, "DefaultOperation")
}

func TestSchemaConstName(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"Vulnerability", "IntegrationMappingSchemaVulnerability"},
		{"DirectoryAccount", "IntegrationMappingSchemaDirectoryAccount"},
		{"", ""},
	}

	for _, c := range cases {
		assert.Equal(t, c.out, schemaConstName(c.in))
	}
}

func TestSchemaIngestTopicName(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"Asset", "integration.ingest.asset.requested"},
		{"DirectoryAccount", "integration.ingest.directory_account.requested"},
		{"DirectoryMembership", "integration.ingest.directory_membership.requested"},
		{"", ""},
	}

	for _, c := range cases {
		assert.Equal(t, c.out, schemaIngestTopicName(c.in))
	}
}

func TestFieldConstName(t *testing.T) {
	cases := []struct {
		schema string
		key    string
		out    string
	}{
		{"Vulnerability", "externalID", "IntegrationMappingVulnerabilityExternalID"},
		{"Asset", "sourceIdentifier", "IntegrationMappingAssetSourceIdentifier"},
		{"", "externalID", ""},
		{"Vulnerability", "", ""},
	}

	for _, c := range cases {
		assert.Equal(t, c.out, fieldConstName(c.schema, c.key))
	}
}

func TestIntegrationFieldForEntField(t *testing.T) {
	cases := []struct {
		in      string
		out     string
		wantErr bool
	}{
		{"integration_id", "ID", false},
		{"owner_id", "OwnerID", false},
		{"platform_id", "PlatformID", false},
		{"unknown_field", "", true},
	}

	for _, c := range cases {
		got, err := integrationFieldForEntField(c.in)
		if c.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, c.out, got)
		}
	}
}

func TestCollectLookupFields(t *testing.T) {
	fields := []MappingField{
		{EntField: "external_id", GoField: "ExternalID", Required: true, LookupKey: true},
		{EntField: "name", GoField: "Name", Required: true, LookupKey: false},
		{EntField: "integration_id", GoField: "IntegrationID", Required: false, LookupKey: true},
	}

	got := collectLookupFields(fields)
	assert.Len(t, got, 2)
	assert.Equal(t, "external_id", got[0].Field)
	assert.True(t, got[0].Required)
	assert.Equal(t, "integration_id", got[1].Field)
	assert.False(t, got[1].Required)
}

func TestCollectRuntimeDefaults(t *testing.T) {
	fields := []MappingField{
		{EntField: "integration_id", GoField: "IntegrationID", Required: false, FromIntegration: true},
		{EntField: "owner_id", GoField: "OwnerID", Required: false, FromIntegration: true},
		{EntField: "name", GoField: "Name", Required: true, FromIntegration: false},
	}

	got, err := collectRuntimeDefaults(fields)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "integration_id", got[0].Field)
	assert.Equal(t, "ID", got[0].IntegrationField)
	assert.Equal(t, "owner_id", got[1].Field)
	assert.Equal(t, "OwnerID", got[1].IntegrationField)
}

func TestCollectRuntimeDefaults_UnknownFieldErrors(t *testing.T) {
	fields := []MappingField{
		{EntField: "custom_field", GoField: "CustomField", FromIntegration: true},
	}

	_, err := collectRuntimeDefaults(fields)
	assert.Error(t, err)
}

func TestFieldIsIncluded(t *testing.T) {
	includeSet := map[string]struct{}{"name": {}, "email": {}}
	excludeSet := map[string]struct{}{"deleted_at": {}}

	// include list is exhaustive — only listed fields pass regardless of system status
	assert.True(t, fieldIsIncluded("name", includeSet, excludeSet, true, false, nil))
	assert.False(t, fieldIsIncluded("other", includeSet, excludeSet, true, false, nil))
	assert.False(t, fieldIsIncluded("deleted_at", includeSet, excludeSet, true, false, nil))

	// no include list — exclude set applies
	assert.False(t, fieldIsIncluded("deleted_at", nil, excludeSet, false, false, nil))

	// no include list — system fields blocked without stock persist + annotation
	assert.False(t, fieldIsIncluded("owner_id", nil, nil, false, false, nil))
	assert.False(t, fieldIsIncluded("owner_id", nil, nil, false, true, nil))

	ant := &entx.IntegrationMappingFieldAnnotation{}
	assert.True(t, fieldIsIncluded("owner_id", nil, nil, false, true, ant))

	// regular field with no lists passes
	assert.True(t, fieldIsIncluded("name", nil, nil, false, false, nil))
}

func TestIntegrationIDIsUpsertKey(t *testing.T) {
	withUpsert := []MappingField{
		{EntField: "integration_id", UpsertKey: true, FromIntegration: true},
		{EntField: "name"},
	}
	without := []MappingField{
		{EntField: "integration_id", UpsertKey: false, FromIntegration: true},
	}
	notFromIntegration := []MappingField{
		{EntField: "integration_id", UpsertKey: true, FromIntegration: false},
	}

	assert.True(t, integrationIDIsUpsertKey(withUpsert))
	assert.False(t, integrationIDIsUpsertKey(without))
	assert.False(t, integrationIDIsUpsertKey(notFromIntegration))
}

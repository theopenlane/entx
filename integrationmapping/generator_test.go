package integrationmapping

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
				IngestTopicConstName:  "IntegrationIngestTopicDirectoryAccountRequested",
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
	assert.Contains(t, generated, "PrepareDirectoryAccountInput")
	assert.NotContains(t, generated, "DefaultOperation")
}

func TestBuildMappingTemplate_OmitsTypedIngestContractsWhenDisabled(t *testing.T) {
	tmpl := buildMappingTemplate()

	data := MappingData{
		PackageName: "integrationgenerated",
		Schemas: []MappingSchema{
			{
				Name:                 "Vulnerability",
				ConstName:            "IntegrationMappingSchemaVulnerability",
				IngestTopicConstName: "IntegrationIngestTopicVulnerabilityRequested",
				IngestTopic:          "integration.ingest.vulnerability.requested",
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
				IngestTopicConstName:  "IntegrationIngestTopicDirectoryGroupRequested",
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

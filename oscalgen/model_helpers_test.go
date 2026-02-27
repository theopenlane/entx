package oscalgen

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewOscalModelsBuilders validates root OSCAL model wrapper helper behavior
func TestNewOscalModelsBuilders(t *testing.T) {
	componentDoc := ComponentDefinition{UUID: "component-definition-uuid"}
	sspDoc := SystemSecurityPlan{UUID: "ssp-uuid"}
	poamDoc := PlanOfActionAndMilestones{UUID: "poam-uuid"}

	componentRoot := NewOscalModelsForComponentDefinition(componentDoc)
	sspRoot := NewOscalModelsForSystemSecurityPlan(sspDoc)
	poamRoot := NewOscalModelsForPOAM(poamDoc)

	assert.NotNil(t, componentRoot.ComponentDefinition)
	assert.Equal(t, "component-definition-uuid", componentRoot.ComponentDefinition.UUID)
	assert.NotNil(t, sspRoot.SystemSecurityPlan)
	assert.Equal(t, "ssp-uuid", sspRoot.SystemSecurityPlan.UUID)
	assert.NotNil(t, poamRoot.PlanOfActionAndMilestones)
	assert.Equal(t, "poam-uuid", poamRoot.PlanOfActionAndMilestones.UUID)
}

// TestMetadataHelpers validates metadata creation and last-modified touch behavior
func TestMetadataHelpers(t *testing.T) {
	initial := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.FixedZone("Offset", -7*60*60))

	metadata := NewMetadata("Doc", "v1", "1.1.3", initial)
	assert.Equal(t, "Doc", metadata.Title)
	assert.Equal(t, "v1", metadata.Version)
	assert.Equal(t, "1.1.3", metadata.OscalVersion)
	assert.Equal(t, initial.UTC(), metadata.LastModified)

	next := time.Date(2026, time.February, 3, 4, 5, 6, 0, time.FixedZone("Offset", 5*60*60))
	TouchMetadata(&metadata, next)
	assert.Equal(t, next.UTC(), metadata.LastModified)

	TouchMetadata(nil, next)
}

// TestExternalTypesAliases validates local aliases resolve to the external oscalot types
func TestExternalTypesAliases(t *testing.T) {
	root := OscalModels{
		ComponentDefinition: &ComponentDefinition{
			UUID: "component-definition-uuid",
		},
		SystemSecurityPlan: &SystemSecurityPlan{
			UUID: "ssp-uuid",
		},
		PlanOfActionAndMilestones: &PlanOfActionAndMilestones{
			UUID: "poam-uuid",
		},
	}

	assert.NotNil(t, root.ComponentDefinition)
	assert.NotNil(t, root.SystemSecurityPlan)
	assert.NotNil(t, root.PlanOfActionAndMilestones)
	assert.Equal(t, "component-definition-uuid", root.ComponentDefinition.UUID)
	assert.Equal(t, "ssp-uuid", root.SystemSecurityPlan.UUID)
	assert.Equal(t, "poam-uuid", root.PlanOfActionAndMilestones.UUID)
	assert.NotEmpty(t, OSCALTypesVersion)
}

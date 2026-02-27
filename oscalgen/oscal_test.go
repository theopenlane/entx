package oscalgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOSCALModelAnnotation validates schema-level OSCAL annotation behavior
func TestOSCALModelAnnotation(t *testing.T) {
	ann := NewOSCALModel(
		WithOSCALModels(OSCALModelComponentDefinition, OSCALModelSSP, OSCALModelSSP),
		WithOSCALAssembly("system-characteristics"),
	)

	assert.Equal(t, OSCALModelAnnotationName, ann.Name())
	assert.Equal(t, []OSCALModelType{OSCALModelComponentDefinition, OSCALModelSSP}, ann.Models)
	assert.Equal(t, "system-characteristics", ann.Assembly)

	decoded := &OSCALModel{}
	err := decoded.Decode(map[string]any{
		"Models":   []string{"component-definition", "component-definition", "poam"},
		"Assembly": "component",
	})
	require.NoError(t, err)
	assert.Equal(t, []OSCALModelType{OSCALModelComponentDefinition, OSCALModelPOAM}, decoded.Models)
	assert.Equal(t, "component", decoded.Assembly)
}

// TestOSCALFieldAnnotation validates field-level OSCAL annotation behavior
func TestOSCALFieldAnnotation(t *testing.T) {
	ann := NewOSCALField(
		OSCALFieldRoleImplementationDetails,
		WithOSCALFieldModels(OSCALModelComponentDefinition, OSCALModelComponentDefinition),
		WithOSCALIdentityAnchor(),
	)

	assert.Equal(t, OSCALFieldAnnotationName, ann.Name())
	assert.Equal(t, OSCALFieldRoleImplementationDetails, ann.Role)
	assert.Equal(t, []OSCALModelType{OSCALModelComponentDefinition}, ann.Models)
	assert.True(t, ann.IdentityAnchor)

	decoded := &OSCALField{}
	err := decoded.Decode(map[string]any{
		"Role":           "statement-id",
		"Models":         []string{"ssp", "ssp"},
		"IdentityAnchor": true,
	})
	require.NoError(t, err)
	assert.Equal(t, OSCALFieldRoleStatementID, decoded.Role)
	assert.Equal(t, []OSCALModelType{OSCALModelSSP}, decoded.Models)
	assert.True(t, decoded.IdentityAnchor)
}

// TestOSCALRelationshipAnnotation validates edge-level OSCAL annotation behavior
func TestOSCALRelationshipAnnotation(t *testing.T) {
	ann := NewOSCALRelationship(
		OSCALRelationshipRoleSatisfiesControl,
		WithOSCALRelationshipModels(OSCALModelComponentDefinition, OSCALModelSSP, OSCALModelSSP),
	)

	assert.Equal(t, OSCALRelationshipAnnotationName, ann.Name())
	assert.Equal(t, OSCALRelationshipRoleSatisfiesControl, ann.Role)
	assert.Equal(t, []OSCALModelType{OSCALModelComponentDefinition, OSCALModelSSP}, ann.Models)

	decoded := &OSCALRelationship{}
	err := decoded.Decode(map[string]any{
		"Role":   "links-to-control-id",
		"Models": []string{"poam"},
	})
	require.NoError(t, err)
	assert.Equal(t, OSCALRelationshipRoleLinksToControlID, decoded.Role)
	assert.Equal(t, []OSCALModelType{OSCALModelPOAM}, decoded.Models)
}

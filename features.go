package entx

// FeatureModule represents a unique feature module identifier.
// Using a defined type because sarah is lazy and also i guess prevents typos
type FeatureModule string

const (
	// ModuleBase is the base feature module available to all users.
	ModuleBase FeatureModule = "base"

	// ModuleCompliance is the compliance feature module.
	ModuleCompliance FeatureModule = "compliance"

	// ModuleContinuousComplianceAutomation is the continuous compliance automation feature module.
	ModuleContinuousComplianceAutomation FeatureModule = "continuous-compliance-automation"

	// ModuleTrustCenter is the trust center feature module.
	ModuleTrustCenter FeatureModule = "trust-center"

	// ModulePolicyManagement is the policy management feature module.
	ModulePolicyManagement FeatureModule = "policy-management"

	// ModuleRiskManagement is the risk management feature module.
	ModuleRiskManagement FeatureModule = "risk-management"

	// ModuleAssetManagement is the asset management feature module.
	ModuleAssetManagement FeatureModule = "asset-management"

	// ModuleEntityManagement is the entity management feature module.
	ModuleEntityManagement FeatureModule = "entity-management"

	// ModuleAuditLog is the feature module containing ent-history schemas.
	ModuleAuditLog FeatureModule = "audit-log"
)

// Visibility is used to flag modules and schemas as public, beta or private.
type Visibility string

const (
	// VisibilityPublic indicates that a module or schema is publicly available.
	VisibilityPublic Visibility = "public"

	// VisibilityBeta marks a module or schema as beta.
	VisibilityBeta Visibility = "beta"

	// VisibilityPrivate marks a module or schema as private/internal.
	VisibilityPrivate Visibility = "private"
)

// ModuleVisibility holds the visibility flag for each feature module.
// Use SetModuleVisibility to configure values at init time.
var ModuleVisibility = map[FeatureModule]Visibility{}

// SetModuleVisibility sets the visibility for a feature module.
func SetModuleVisibility(m FeatureModule, v Visibility) {
	ModuleVisibility[m] = v
}

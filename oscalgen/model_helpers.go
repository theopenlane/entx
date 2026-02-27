package oscalgen

import "time"

// NewOscalModelsForComponentDefinition creates a root OSCAL models container for component-definition payloads
func NewOscalModelsForComponentDefinition(doc ComponentDefinition) OscalModels {
	return OscalModels{
		ComponentDefinition: &doc,
	}
}

// NewOscalModelsForSystemSecurityPlan creates a root OSCAL models container for SSP payloads
func NewOscalModelsForSystemSecurityPlan(doc SystemSecurityPlan) OscalModels {
	return OscalModels{
		SystemSecurityPlan: &doc,
	}
}

// NewOscalModelsForPOAM creates a root OSCAL models container for POA&M payloads
func NewOscalModelsForPOAM(doc PlanOfActionAndMilestones) OscalModels {
	return OscalModels{
		PlanOfActionAndMilestones: &doc,
	}
}

// NewMetadata creates minimal OSCAL metadata with an explicit last-modified timestamp
func NewMetadata(title, version, oscalVersion string, lastModified time.Time) Metadata {
	return Metadata{
		Title:        title,
		Version:      version,
		OscalVersion: oscalVersion,
		LastModified: lastModified.UTC(),
	}
}

// TouchMetadata updates metadata last-modified using UTC for deterministic serialization
func TouchMetadata(metadata *Metadata, now time.Time) {
	if metadata == nil {
		return
	}

	metadata.LastModified = now.UTC()
}

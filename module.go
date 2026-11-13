package entx

const SkipQueryCheckForModuleAnnotationName = "OPENLANE_SKIP_QUERY_CHECK_FOR_MODULE"

// SkipQueryCheckForModuleAnnotation allows certain queries to be executed regardless of module checks.
type SkipQueryCheckForModuleAnnotation struct{}

// Name returns the name of the SkipQueryCheckForModuleAnnotation
func (SkipQueryCheckForModuleAnnotation) Name() string { return SkipQueryCheckForModuleAnnotationName }

// Decode unmarshalls the SkipQueryCheckForModuleAnnotation
func (a *SkipQueryCheckForModuleAnnotation) Decode(annotation any) error {
	return DecodeAnnotation(annotation, a)
}

// SkipQueryCheckForModule returns a new SearchFieldAnnotation with the flag set
func SkipQueryCheckForModule() SkipQueryCheckForModuleAnnotation {
	return SkipQueryCheckForModuleAnnotation{}
}

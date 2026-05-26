package entx

const (
	// SkipType skips generating fga crud types or fields in the model
	FGASkipType FGASkipMode = 1 << iota
	// SkipCreate skips crud create fga settings
	SkipCreate
	// SkipUpdate skips crud update fga settings
	SkipUpdate
	// SkipDelete skips crud delete fga settings
	SkipDelete

	// SkipAll is default mode to skip all.
	SkipAll = FGASkipType |
		SkipCreate |
		SkipUpdate |
		SkipDelete
)

// FGASkipMode is the skip for crud annotations
type FGASkipMode int

// Has determines if the skip mode contains the flag
func (m FGASkipMode) Has(flag FGASkipMode) bool {
	return m&flag == flag
}

// FGACrudSkip returns a skip annotation.
// The Skip() annotation is used to skip
// generating crud tuples for fga settings
//
// It gives you the flexibility to skip generating
// the type or the field based on the SkipMode flags.
func FGACrudSkip(flags ...FGASkipMode) FGACrudAnnotation {
	if len(flags) == 0 {
		return FGACrudAnnotation{Skip: SkipAll}
	}

	skip := FGASkipMode(0)
	for _, f := range flags {
		skip |= f
	}

	return FGACrudAnnotation{Skip: skip}
}

// FGACrudParent returns the FGAParentCrudAnnotation with the parent schemas set
func FGACrudParent(schemas ...string) FGAParentCrudAnnotation {
	if len(schemas) == 0 {
		return FGAParentCrudAnnotation{}
	}

	return FGAParentCrudAnnotation{ParentSchemas: schemas}
}

package entx

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

var (
	softDeleteSkipKey = contextx.NewKey[bool]()
	softDeleteTypeKey = contextx.NewKey[string]()
)

// SkipSoftDelete returns a new context that skips the soft-delete interceptor/hooks.
func SkipSoftDelete(parent context.Context) context.Context {
	return softDeleteSkipKey.Set(parent, true)
}

// CheckSkipSoftDelete checks whether soft-delete skipping was requested.
func CheckSkipSoftDelete(ctx context.Context) bool {
	skip, _ := softDeleteSkipKey.Get(ctx)
	return skip
}

// IsSoftDelete returns a new context that informs the delete is a soft-delete for interceptor/hooks.
func IsSoftDelete(parent context.Context, objectType string) context.Context {
	return softDeleteTypeKey.Set(parent, objectType)
}

// CheckIsSoftDeleteType checks if the softDeleteKey is set in the context for the given object type
func CheckIsSoftDeleteType(ctx context.Context, objectType string) bool {
	val, ok := softDeleteTypeKey.Get(ctx)

	return ok && val == objectType
}

// CheckIsSoftDelete checks if the softDeleteKey is set in the context
// @deprecated use CheckIsSoftDeleteType with object type instead
func CheckIsSoftDelete(ctx context.Context) bool {
	_, ok := softDeleteTypeKey.Get(ctx)
	return ok
}

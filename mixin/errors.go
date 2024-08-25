package mixin

import (
	"fmt"

	"entgo.io/ent"
)

// UnexpectedAuditError is returned when an unexpected audit log call is received
type UnexpectedAuditError struct {
	MutationType ent.Mutation
}

// Error returns the UnexpectedAuditError in string format
func (e *UnexpectedAuditError) Error() string {
	return fmt.Sprintf("unexpected audit log call from mutation type: %T", e.MutationType)
}

func newUnexpectedAuditError(arg ent.Mutation) *UnexpectedAuditError {
	return &UnexpectedAuditError{
		MutationType: arg,
	}
}

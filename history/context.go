package history

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/theopenlane/utils/contextx"
)

// RequestMarker marks an internal history write/read context.
// It is used to bypass auth checks during creation/query of history records.
type RequestMarker struct{}

var historyContextKey = contextx.NewKey[RequestMarker]()

// WithContext sets the history context in the context
func WithContext(ctx context.Context) context.Context {
	return historyContextKey.Set(ctx, RequestMarker{})
}

// FromContext retrieves the history context from the context
func FromContext(ctx context.Context) (RequestMarker, bool) {
	return historyContextKey.Get(ctx)
}

// IsHistoryRequest checks if the context has a history request key
func IsHistoryRequest(ctx context.Context) bool {
	_, ok := FromContext(ctx)

	return ok
}

// AllowIfHistoryRequest allows the query to proceed if the context history in the context
func AllowIfHistoryRequest() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if ok := IsHistoryRequest(ctx); ok {
			return privacy.Allow
		}

		return privacy.Skipf("history request not found in context")
	})
}

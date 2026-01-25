package genhooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// normalizeWhitespace collapses all whitespace sequences into single spaces
// and removes spaces around dots, parens, and brackets for consistent comparison
func normalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	joined := strings.Join(fields, " ")
	joined = strings.ReplaceAll(joined, ". ", ".")
	joined = strings.ReplaceAll(joined, " .", ".")
	joined = strings.ReplaceAll(joined, "( ", "(")
	joined = strings.ReplaceAll(joined, " )", ")")
	joined = strings.ReplaceAll(joined, "[ ", "[")
	joined = strings.ReplaceAll(joined, " ]", "]")

	return joined
}

func TestUpdateWorkflowResolvers(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	graphDir := filepath.Join(root, "internal", "graphapi")
	if err := os.MkdirAll(graphDir, 0o755); err != nil { // nolint:mnd
		t.Fatalf("mkdir graphapi: %v", err)
	}

	source := `package graphapi

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"example.com/test/internal/ent/generated"
)

func (r *controlResolver) HasPendingWorkflow(ctx context.Context, obj *generated.Control) (bool, error) {
	panic(fmt.Errorf("not implemented: HasPendingWorkflow - hasPendingWorkflow"))
}

func (r *controlResolver) HasWorkflowHistory(ctx context.Context, obj *generated.Control) (bool, error) {
	panic(fmt.Errorf("not implemented: HasWorkflowHistory - hasWorkflowHistory"))
}

func (r *controlResolver) ActiveWorkflowInstances(ctx context.Context, obj *generated.Control) ([]*generated.WorkflowInstance, error) {
	panic(fmt.Errorf("not implemented: ActiveWorkflowInstances - activeWorkflowInstances"))
}

func (r *controlResolver) WorkflowTimeline(ctx context.Context, obj *generated.Control, after *entgql.Cursor[string], first *int, before *entgql.Cursor[string], last *int, orderBy []*generated.WorkflowEventOrder, where *generated.WorkflowEventWhereInput, includeEmitFailures *bool) (*generated.WorkflowEventConnection, error) {
	panic(fmt.Errorf("not implemented: WorkflowTimeline - workflowTimeline"))
}
`

	resolverPath := filepath.Join(graphDir, "control.resolvers.go")
	if err := os.WriteFile(resolverPath, []byte(source), 0o600); err != nil { // nolint:mnd
		t.Fatalf("write resolver file: %v", err)
	}

	if err := UpdateWorkflowResolvers(graphDir); err != nil {
		t.Fatalf("UpdateWorkflowResolvers failed: %v", err)
	}

	updated, err := os.ReadFile(resolverPath)
	if err != nil {
		t.Fatalf("read updated resolver file: %v", err)
	}

	updatedStr := string(updated)

	normalizedStr := normalizeWhitespace(updatedStr)
	if strings.Contains(updatedStr, "\"fmt\"") {
		t.Fatalf("expected fmt import to be removed")
	}

	if !strings.Contains(normalizedStr, "return workflowResolverHasPending(ctx, generated.TypeControl, obj.ID)") {
		t.Fatalf("expected HasPendingWorkflow to call helper")
	}

	if !strings.Contains(normalizedStr, "return workflowResolverHasHistory(ctx, generated.TypeControl, obj.ID)") {
		t.Fatalf("expected HasWorkflowHistory to call helper")
	}

	if !strings.Contains(normalizedStr, "return workflowResolverActiveInstances(ctx, generated.TypeControl, obj.ID)") {
		t.Fatalf("expected ActiveWorkflowInstances to call helper")
	}

	// Note: Go formatter may add trailing comma for multi-line calls, so check with or without it
	timelineCall := "return workflowResolverTimeline(ctx, generated.TypeControl, obj.ID, after, first, before, last, orderBy, where, includeEmitFailures"
	if !strings.Contains(normalizedStr, timelineCall) {
		t.Logf("normalized output:\n%s", normalizedStr)
		t.Fatalf("expected WorkflowTimeline to call helper with pagination params")
	}

	helperPath := filepath.Join(graphDir, workflowResolverHelperFile)

	helperBytes, err := os.ReadFile(helperPath)
	if err != nil {
		t.Fatalf("read helper file: %v", err)
	}

	helperStr := string(helperBytes)
	if !strings.Contains(helperStr, "package graphapi") {
		t.Fatalf("expected helper file to use graphapi package")
	}

	if !strings.Contains(helperStr, "example.com/test/internal/ent/generated") {
		t.Fatalf("expected helper file to include generated import")
	}

	if !strings.Contains(helperStr, "example.com/test/common/enums") {
		t.Fatalf("expected helper file to include enums import")
	}
}

package entityops

import (
	"testing"

	"github.com/theopenlane/entx/vanilla/_example/tasktemplates"
)

// TestOrganizationTaskRules verifies the entx.FieldTaskRule/SchemaTaskRule annotations declared on
// the example Organization schema (see ../schema/organization.go) flow through codegen correctly,
// and that every generated rule resolves against a real task template -- the full annotation ->
// codegen -> consumer-resolution round trip, not just the trigger half
func TestOrganizationTaskRules(t *testing.T) {
	if !SchemaOrganization.TaskRuleEligible() {
		t.Fatal("SchemaOrganization should be task-rule eligible")
	}

	found := map[string]FieldTaskRule{}

	for _, rule := range SchemaOrganization.AllTaskRules() {
		found[rule.Rule.RuleID] = rule
	}

	schemaLevel, ok := found["setup-payment-method"]
	if !ok {
		t.Fatal("expected schema-level setup-payment-method rule")
	}

	if schemaLevel.Field != "" {
		t.Fatalf("schema-level rule should have no field, got %q", schemaLevel.Field)
	}

	policies, ok := found["import-existing-policies"]
	if !ok {
		t.Fatal("expected import-existing-policies rule")
	}

	if policies.Field != "preferences" {
		t.Fatalf("expected import-existing-policies on preferences field, got %q", policies.Field)
	}

	if policies.Rule.Expression != "value.policies.has_existing == true" {
		t.Fatalf("unexpected expression: %q", policies.Rule.Expression)
	}

	framework, ok := found["framework"]
	if !ok {
		t.Fatal("expected framework rule")
	}

	if framework.Rule.EachElement != "value.frameworks" {
		t.Fatalf("unexpected eachElement: %q", framework.Rule.EachElement)
	}

	names := map[string]bool{}
	for _, s := range TaskRuleEligibleSchemas() {
		names[s.Name] = true
	}

	if !names["Organization"] {
		t.Fatal("TaskRuleEligibleSchemas should include Organization")
	}

	// every rule the schema declares must resolve to real task content -- proving there are no
	// dangling RuleIDs left half-wired between the annotation and the consumer's template registry
	for _, rule := range SchemaOrganization.AllTaskRules() {
		tmpl, ok := tasktemplates.Lookup(rule.Rule.RuleID)
		if !ok {
			t.Fatalf("rule %q has no matching task template", rule.Rule.RuleID)
		}

		if tmpl.Title == "" {
			t.Fatalf("rule %q resolved to an empty title", rule.Rule.RuleID)
		}
	}

	policyTemplate, ok := tasktemplates.Lookup(policies.Rule.RuleID)
	if !ok || policyTemplate.Title != "Import existing policies" {
		t.Fatalf("import-existing-policies should resolve to its rendered title, got %+v", policyTemplate)
	}
}

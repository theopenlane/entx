// Package tasktemplates is example consumer-side plumbing for task rules, used as examples only
package tasktemplates

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed templates.yaml
var templatesYAML []byte

// Template is the rendered content for one task rule
type Template struct {
	Title    string `yaml:"title"`
	Details  string `yaml:"details"`
	Priority int    `yaml:"priority"`
}

type templateFile struct {
	Rules map[string]Template `yaml:"rules"`
}

var registry = mustLoadRegistry(templatesYAML)

func mustLoadRegistry(raw []byte) map[string]Template {
	var file templateFile
	if err := yaml.Unmarshal(raw, &file); err != nil {
		panic("tasktemplates: invalid templates.yaml: " + err.Error())
	}

	return file.Rules
}

// Lookup returns the rendered template registered for ruleID
func Lookup(ruleID string) (Template, bool) {
	t, ok := registry[ruleID]

	return t, ok
}

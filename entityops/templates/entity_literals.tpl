{{- define "taskRuleLiteral" -}}
{RuleID: {{ printf "%q" .RuleID }}{{ if .Expression }}, Expression: {{ printf "%q" .Expression }}{{ end }}{{ if .EachElement }}, EachElement: {{ printf "%q" .EachElement }}{{ end }}{{ if .Trigger }}, Trigger: {{ printf "%q" .Trigger }}{{ end }}}
{{- end }}

{{- define "fieldDescriptorLiteral" -}}
{Name: "{{ .Snake }}", Label: "{{ .Name }}", Type: "{{ .Type }}"{{ if .WorkflowEligible }}, WorkflowEligible: true{{ end }}{{ if .MatchKey }}, MatchKey: true{{ end }}{{ if and .IntegrationMapped (not .SystemControlled) }}, InputKey: "{{ .InputKey }}"{{ end }}{{ if .LookupKey }}, LookupKey: true{{ end }}{{ if .DisplayKey }}, DisplayKey: true{{ end }}{{ if .Clearable }}, Clearable: true{{ end }}{{ if .WebhookPayload }}, WebhookPayload: true{{ end }}{{ if .SystemControlled }}, SystemControlled: true{{ end }}{{ if .SourceManaged }}, SourceManaged: true{{ end }}{{ if .Volatile }}, Volatile: true{{ end }}{{ if .CaseInsensitive }}, CaseInsensitive: true{{ end }}{{ if .TaskRules }}, TaskRules: []TaskRuleDescriptor{ {{ range .TaskRules }}{{ template "taskRuleLiteral" . }}, {{ end }} }{{ end }}}
{{- end }}

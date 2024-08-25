{{- if .IncludeMutations }}
mutation CreateBulkCSV{{ .Name }}($input: Upload!) {
  createBulkCSV{{ .Name }}(input: $input) {
    {{ .Name | ToLowerCamel | ToPlural }} {
      {{- range .Fields }}
      {{.}}
      {{- end}}
    }
  }
}

mutation CreateBulk{{ .Name }}($input: [Create{{ .Name }}Input!]) {
  createBulk{{ .Name }}(input: $input) {
    {{ .Name | ToLowerCamel | ToPlural }} {
      {{- range .Fields }}
      {{.}}
      {{- end}}
    }
  }
}

mutation Create{{ .Name }}($input: Create{{ .Name }}Input!) {
  create{{ .Name }}(input: $input) {
    {{ .Name | ToLowerCamel }} {
      {{- range .Fields }}
      {{.}}
      {{- end}}
    }
  }
}

mutation Delete{{ .Name }}($delete{{ .Name }}Id: ID!) {
  delete{{ .Name }}(id: $delete{{ .Name }}Id) {
    deletedID
  }
}
{{- end}}

query GetAll{{ .Name | ToPlural }} {
  {{ .Name | ToLowerCamel | ToPlural }} {
    edges {
      node {
        {{- range .Fields }}
        {{.}}
        {{- end}}
      }
    }
  }
}

{{- if not .IsHistory }}
query Get{{ .Name }}ByID(${{ .Name | ToLowerCamel }}Id: ID!) {
  {{ .Name | ToLowerCamel }}(id: ${{ .Name | ToLowerCamel }}Id) {
    {{- range .Fields }}
    {{.}}
    {{- end}}
  }
}
{{- end}}

query Get{{ .Name | ToPlural }}($where: {{ .Name }}WhereInput) {
  {{ .Name | ToLowerCamel | ToPlural }}(where: $where) {
    edges {
      node {
        {{- range .Fields }}
        {{.}}
        {{- end}}
      }
    }
  }
}

{{- if .IncludeMutations }}
mutation Update{{ .Name }}($update{{ .Name }}Id: ID!, $input: Update{{ .Name }}Input!) {
  update{{ .Name }}(id: $update{{ .Name }}Id, input: $input) {
    {{ .Name | ToLowerCamel }} {
      {{- range .Fields }}
      {{.}}
      {{- end}}
    }
  }
}
{{- end}}

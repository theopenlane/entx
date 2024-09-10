query {{ $.Name }}Search($query: String!) {
  {{- if eq $.Name "Admin" }}
  adminSearch(query: $query) {
  {{- else }}
  search(query: $query) {
  {{- end }}
    nodes {
    {{- range $object := $.Objects }}
      ... on {{ $object.Name | toUpperCamel }}SearchResult {
        {{ $object.Name| toLower | toPlural }} {
          {{- if eq $.Name "Admin" }}
          {{- range $field := $object.AdminFields }}
          {{ $field.Name | toLower }}
          {{- end }}
          {{- else }}
          {{- range $field := $object.Fields }}
          {{ $field.Name  | toLower }}
          {{- end }}
          {{- end }}
        }
      }
    {{- end }}
    }
  }
}

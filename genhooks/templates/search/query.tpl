query {{ $.Name }}Search($query: String!) {
  search(query: $query) {
    nodes {
    {{- range $object := $.Objects }}
      ... on {{ $object.Name | toUpperCamel }}SearchResult {
        {{ $object.Name| toLowerCamel | toPlural }} {
          {{- if eq $.Name "Admin" }}
          {{- range $field := $object.AdminFields }}
          {{ $field | toLowerCamel }}
          {{- end }}
          {{- else }}
          {{- range $field := $object.Fields }}
          {{ $field | toLowerCamel }}
          {{- end }}
          {{- end }}
        }
      }
    {{- end }}
    }
  }
}

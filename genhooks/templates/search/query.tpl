query Search($query: String!) {
  search(query: $query) {
    nodes {
    {{- range $object := $.Objects }}
      ... on {{ $object.Name | toUpperCamel }}SearchResult {
        {{ $object.Name| toLowerCamel | toPlural }} {
          {{- range $field := $object.Fields }}
          {{ $field | toLowerCamel }}
          {{- end }}
        }
      }
    {{- end }}
    }
  }
}

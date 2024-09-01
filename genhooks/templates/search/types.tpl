union SearchResult =
  {{- range $object := $.Objects }}
  | {{ $object.Name | toUpperCamel }}SearchResult
  {{- end }}

extend type Query{
    """
    Search across all objects
    """
    search(
        """
        Search query
        """
        query: String!
    ): SearchResult
}

{{ range $object := $.Objects }}
type  {{ $object.Name }}SearchResult {
   {{ $object.Name | toLowerCamel | toPlural }}: [ {{ $object.Name | toUpperCamel}}!]
}
{{ end }}
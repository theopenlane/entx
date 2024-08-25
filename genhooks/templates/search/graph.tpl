union GlobalSearchResult =
  {{- range $object := $.Objects }}
  | {{ $object.Name | toUpperCamel }}SearchResult
  {{- end }}


{{- range $object := $.Objects }}
type  {{ $object.Name }}SearchResult {
   {{ $object.Name | toLowerCamel | toPlural }}: [ {{ $object.Name | toUpperCamel}}!]
}

{{- end }}

type GlobalSearchResultConnection {
  page: PageInfo!

  nodes: [GlobalSearchResult!]!
}


extend type Query{
    """
    Search across objects
    """
    search(
        """
        Search query
        """
        query: String!
    ): GlobalSearchResultConnection
}

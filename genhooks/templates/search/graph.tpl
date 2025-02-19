extend type Query{
    {{- range $object := $.Objects }}
    """
    Search across {{ $object.Name }} objects
    """
    {{- if eq $.Name "Global" }}
    {{ $object.Name | toLower }}Search(
    {{- else }}
    {{ $.Name | toLower }}{{ $object.Name | toUpperCamel }}Search(
    {{- end }}
        """
        Search query
        """
        query: String!
    ): {{ $object.Name }}SearchResult
    {{- end }}
}

{{- if eq $.Name "Global" }}
union SearchResult =
  {{- range $object := $.Objects }}
  | {{ $object.Name | toUpperCamel }}SearchResult
  {{- end }}

type SearchResultConnection {
  """
  Information to aid in pagination.
  """
  page: PageInfo!
  """
  Identifies the total count of items in the connection.
  """
  totalCount: Int!
  """
  A list of nodes with results.
  """
  nodes: [SearchResult!]!
}

extend type Query{
    """
    Search across all objects
    """
    search(
        """
        Search query
        """
        query: String!
    ): SearchResultConnection
    """
    Admin search across all objects
    """
    adminSearch(
        """
        Search query
        """
        query: String!
    ): SearchResultConnection
}
{{ range $object := $.Objects }}
type  {{ $object.Name }}SearchResult {
   {{ $object.Name | toLower | toPlural }}: [ {{ $object.Name | toUpperCamel}}!]
}
{{ end }}
{{- end }}
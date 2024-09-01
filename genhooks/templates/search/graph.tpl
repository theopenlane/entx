extend type Query{
    {{- range $object := $.Objects }}
    """
    Search across {{ $object.Name }} objects
    """
    {{- if eq $.Name "Global" }}
    {{ $object.Name | toLowerCamel }}Search(
    {{- else }}
    {{ $.Name | toLowerCamel }}{{ $object.Name | toUpperCamel }}Search(
    {{- end }}
        """
        Search query
        """
        query: String!
    ): {{ $object.Name }}SearchResult
    {{- end }}
}
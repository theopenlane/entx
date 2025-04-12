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
        Query string to search across objects
        """
        query: String!
        """
        Returns the elements in the list that come after the specified cursor.
        """
        after: Cursor
        """
        Returns the first _n_ elements from the list.
        """
        first: Int
        """
        Returns the elements in the list that come before the specified cursor.
        """
        before: Cursor
        """
        Returns the last _n_ elements from the list.
        """
        last: Int
    ): {{ $object.Name | toUpperCamel }}Connection
    {{- end }}
}

{{- if eq $.Name "Global" }}
type SearchResults{
  """
  Information to aid in pagination.
  """
  page: PageInfo!
  """
  Identifies the total count of items in the connection.
  """
  totalCount: Int!
  {{- range $object := $.Objects }}
  {{ $object.Name | toLower |toPlural }}: {{ $object.Name | toUpperCamel }}Connection
  {{- end }}
}

extend type Query{
    """
    Search across all objects
    """
    search(
        """
        Query string to search across objects
        """
        query: String!
        """
        Returns the elements in the list that come after the specified cursor.
        """
        after: Cursor
        """
        Returns the first _n_ elements from the list.
        """
        first: Int
        """
        Returns the elements in the list that come before the specified cursor.
        """
        before: Cursor
        """
        Returns the last _n_ elements from the list.
        """
        last: Int
    ): SearchResults
    """
    Admin search across all objects
    """
    adminSearch(
        """
        Query string to search across objects
        """
        query: String!
        """
        Returns the elements in the list that come after the specified cursor.
        """
        after: Cursor
        """
        Returns the first _n_ elements from the list.
        """
        first: Int
        """
        Returns the elements in the list that come before the specified cursor.
        """
        before: Cursor
        """
        Returns the last _n_ elements from the list.
        """
        last: Int
    ): SearchResults
}
{{- end }}
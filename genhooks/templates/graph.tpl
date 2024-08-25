extend type Query {
    """
    Look up {{ .Name | ToLowerCamel }} by ID
    """
     {{ .Name | ToLowerCamel }}(
        """
        ID of the {{ .Name | ToLowerCamel }}
        """
        id: ID!
    ):  {{ .Name }}!
}

extend type Mutation{
    """
    Create a new {{ .Name | ToLowerCamel }}
    """
    create{{ .Name }}(
        """
        values of the {{ .Name | ToLowerCamel}}
        """
        input: Create{{ .Name }}Input!
    ): {{ .Name }}CreatePayload!
    """
    Create multiple new {{ .Name | ToLowerCamel }}s
    """
    createBulk{{ .Name }}(
        """
        values of the {{ .Name | ToLowerCamel}}
        """
        input: [Create{{ .Name }}Input!]
    ): {{ .Name }}BulkCreatePayload!
    """
    Create multiple new {{ .Name | ToLowerCamel }}s via file upload
    """
    createBulkCSV{{ .Name }}(
        """
        csv file containing values of the {{ .Name | ToLowerCamel}}
        """
        input: Upload!
    ): {{ .Name }}BulkCreatePayload!
    """
    Update an existing {{ .Name | ToLowerCamel }}
    """
    update{{ .Name }}(
        """
        ID of the {{ .Name | ToLowerCamel }}
        """
        id: ID!
        """
        New values for the {{ .Name | ToLowerCamel }}
        """
        input: Update{{ .Name }}Input!
    ): {{ .Name }}UpdatePayload!
    """
    Delete an existing {{ .Name | ToLowerCamel }}
    """
    delete{{ .Name }}(
        """
        ID of the {{ .Name | ToLowerCamel }}
        """
        id: ID!
    ): {{ .Name }}DeletePayload!
}

"""
Return response for create{{ .Name }} mutation
"""
type {{ .Name }}CreatePayload {
    """
    Created {{ .Name | ToLowerCamel }}
    """
    {{ .Name | ToLowerCamel }}: {{ .Name }}!
}

"""
Return response for update{{ .Name }} mutation
"""
type {{ .Name }}UpdatePayload {
    """
    Updated {{ .Name | ToLowerCamel }}
    """
    {{ .Name | ToLowerCamel }}: {{ .Name }}!
}

"""
Return response for delete{{ .Name }} mutation
"""
type {{ .Name }}DeletePayload {
    """
    Deleted {{ .Name | ToLowerCamel }} ID
    """
    deletedID: ID!
}

"""
Return response for createBulk{{ .Name }} mutation
"""
type {{ .Name }}BulkCreatePayload {
    """
    Created {{ .Name | ToLowerCamel }}s
    """
    {{ .Name | ToLowerCamel | ToPlural }}: [{{ .Name }}!]
}
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
    Update multiple existing {{ .Name | ToLowerCamel }}s
    """
    updateBulk{{ .Name }}(
        """
        IDs of the {{ .Name | ToLowerCamel }}s to update
        """
        ids: [ID!]!
        """
        values to update the {{ .Name | ToLowerCamel }}s with
        """
        input: Update{{ .Name }}Input!
    ): {{ .Name }}BulkUpdatePayload!
    """
    Update multiple existing {{ .Name | ToLowerCamel }}s via file upload
    """
    updateBulkCSV{{ .Name }}(
        """
        csv file containing values of the {{ .Name | ToLowerCamel}}, must include ID column
        """
        input: Upload!
    ): {{ .Name }}BulkUpdatePayload!
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
    """
    Delete multiple {{ .Name | ToLowerCamel }}s
    """
    deleteBulk{{ .Name }}(
        """
        IDs of the {{ .Name | ToLowerCamel }}s to delete
        """
        ids: [ID!]!
    ): {{ .Name }}BulkDeletePayload!
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

"""
Return response for updateBulk{{ .Name }} mutation
"""
type {{ .Name }}BulkUpdatePayload {
    """
    Updated {{ .Name | ToLowerCamel }}s
    """
    {{ .Name | ToLowerCamel | ToPlural }}: [{{ .Name }}!]
    """
    IDs of the updated {{ .Name | ToLowerCamel }}s
    """
    updatedIDs: [ID!]
}

"""
Return response for deleteBulk{{ .Name }} mutation
"""
type {{ .Name }}BulkDeletePayload {
    """
    Deleted {{ .Name | ToLowerCamel }} IDs
    """
    deletedIDs: [ID!]!
}

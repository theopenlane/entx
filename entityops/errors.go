package entityops

import "errors"

// ErrProvenanceFieldsMissing indicates an ingest-capable schema lacks the provenance mixin fields
var ErrProvenanceFieldsMissing = errors.New("entityops: ingest-capable schema is missing the provenance fields; add ProvenanceMixin to the schema mixin list")

// ErrInvalidConsoleRoute indicates a console route declares both a query id parameter and a path suffix
var ErrInvalidConsoleRoute = errors.New("entityops: console route id parameter and suffix are mutually exclusive")

// ErrDisplayNameConflict indicates a schema marks more than one display-name field
var ErrDisplayNameConflict = errors.New("entityops: schema declares multiple display name fields")

// ErrMentionSourceConflict indicates a schema marks more than one mention source of the same kind
var ErrMentionSourceConflict = errors.New("entityops: schema declares multiple mention sources of the same kind")

// ErrMentionSourceType indicates a mention source field is neither a string nor a JSON field
var ErrMentionSourceType = errors.New("entityops: mention source must be a string or JSON field")

// ErrMentionOrgOwnedRequired indicates mention sources are declared on a schema that is not org owned
var ErrMentionOrgOwnedRequired = errors.New("entityops: mention scanning requires an org-owned schema")

// ErrApprovalFieldConflict indicates a schema marks more than one approval field of the same kind
var ErrApprovalFieldConflict = errors.New("entityops: schema declares multiple approval fields of the same kind")

// ErrApprovalSpecIncomplete indicates a schema marks only one of the approval status and approver fields
var ErrApprovalSpecIncomplete = errors.New("entityops: approval requires both a status and an approver field")

// ErrApprovalFieldType indicates an approval marker sits on a field of the wrong type
var ErrApprovalFieldType = errors.New("entityops: approval status must be an enum field and approver a string field")

// ErrApprovalOrgOwnedRequired indicates approval fields are declared on a schema that is not org owned
var ErrApprovalOrgOwnedRequired = errors.New("entityops: approval flow requires an org-owned schema")

// ErrLookupAlternativeFieldUnknown indicates a lookup alternative references an unknown or unmapped field
var ErrLookupAlternativeFieldUnknown = errors.New("entityops: lookup alternative references an unknown or unmapped field")

// ErrLookupAlternativeWithoutMapping indicates lookup alternatives declared on an unmapped schema
var ErrLookupAlternativeWithoutMapping = errors.New("entityops: lookup alternative declared on a schema without integration mapping")

// ErrLookupWhereNullFieldInvalid indicates a lookup null filter references an unknown or non-optional field
var ErrLookupWhereNullFieldInvalid = errors.New("entityops: lookup null filter references an unknown or non-optional field")

// ErrSnapshotRemovalConflict indicates a schema marks more than one snapshot-removal field
var ErrSnapshotRemovalConflict = errors.New("entityops: schema declares multiple snapshot removal fields")

// ErrCatalogEdgeInvalid indicates a catalog edge is not a unique self edge with a declared foreign-key field
var ErrCatalogEdgeInvalid = errors.New("entityops: catalog edge must be a unique self edge with a declared foreign-key edge field")

// ErrOwnerEdgeMissing indicates an org-owned schema lacks an owner edge with a foreign-key field
var ErrOwnerEdgeMissing = errors.New("entityops: org-owned schema requires an owner edge with a foreign-key field")

// ErrCatalogManagedDirectiveMissing indicates a source-managed field lacks the catalogManaged directive
var ErrCatalogManagedDirectiveMissing = errors.New("entityops: catalog source-managed field lacks the catalogManaged directive")

// ErrCatalogIncomplete indicates a catalog schema is missing a required part
var ErrCatalogIncomplete = errors.New("entityops: catalog schema is missing a required part")

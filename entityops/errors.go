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

// ErrSnapshotRemovalConflict indicates a schema marks more than one snapshot-removal field
var ErrSnapshotRemovalConflict = errors.New("entityops: schema declares multiple snapshot removal fields")

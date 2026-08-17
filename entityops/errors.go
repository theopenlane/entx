package entityops

import "errors"

// ErrNoIntegrationFieldMapping indicates a FromIntegration field has no mapping to a *generated.Integration Go field
var ErrNoIntegrationFieldMapping = errors.New("entityops: no integration field mapping for ent field")

// ErrInvalidConsoleRoute indicates a console route declares both a query id parameter and a path suffix
var ErrInvalidConsoleRoute = errors.New("entityops: console route id parameter and suffix are mutually exclusive")

// ErrMentionFieldMissing indicates a mentionable annotation references a field the schema does not declare
var ErrMentionFieldMissing = errors.New("entityops: mentionable field does not exist on schema")

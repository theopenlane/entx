package entityops

import "errors"

// ErrNoIntegrationFieldMapping indicates a FromIntegration field has no mapping to a *generated.Integration Go field
var ErrNoIntegrationFieldMapping = errors.New("entityops: no integration field mapping for ent field")

package integrationmapping

import "errors"

// ErrNoIntegrationFieldMapping is returned when a field annotated with FromIntegration
// has no corresponding Integration field mapping
var ErrNoIntegrationFieldMapping = errors.New("field has no known Integration field mapping")

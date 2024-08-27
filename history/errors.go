package enthistory

import (
	"errors"
)

var (
	// ErrUnsupportedIDType is returned when id type other than string or int is used
	ErrUnsupportedIDType = errors.New("unsupported id type, only int and strings are allowed")

	// ErrUnsupportedType is returned when the object type is not supported
	ErrUnsupportedType = errors.New("unsupported type")

	// ErrNoIDType is returned when the id type cannot be determined from the schema
	ErrNoIDType = errors.New("could not get id type for schema")

	// ErrInvalidSchemaPath is returned when the schema path cannot be determined
	ErrInvalidSchemaPath = errors.New("invalid schema path, unable to find package name in path")

	// ErrFailedToGenerateTemplate is returned when the template cannot be generated
	ErrFailedToGenerateTemplate = errors.New("failed to generate template")

	// ErrFailedToWriteTemplate is returned when the template cannot be written
	ErrFailedToWriteTemplate = errors.New("failed to write template")
)

package entx

import (
	"errors"
	"fmt"
)

var (
	// ErrUnsupportedDialect is returned when an unsupported dialect is used
	ErrUnsupportedDialect = errors.New("unsupported dialect")
)

func newDialectError(dialect string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedDialect, dialect)
}

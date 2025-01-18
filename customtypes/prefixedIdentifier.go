package customtypes

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"entgo.io/ent/schema/field"
)

// PrefixedIdentifier is a custom type that implements the TypeValueScanner interface
type PrefixedIdentifier struct {
	prefix string
}

// NewPrefixedIdentifier returns a new PrefixedIdentifier with the given prefix
func NewPrefixedIdentifier(prefix string) PrefixedIdentifier {
	return PrefixedIdentifier{prefix: prefix}
}

// Value implements the TypeValueScanner.Value method.
func (p PrefixedIdentifier) Value(s string) (driver.Value, error) {
	return strings.TrimPrefix(s, p.prefix+"-"), nil
}

// ScanValue implements the TypeValueScanner.ScanValue method.
func (PrefixedIdentifier) ScanValue() field.ValueScanner {
	return &sql.NullString{}
}

// FromValue implements the TypeValueScanner.FromValue method.
func (p PrefixedIdentifier) FromValue(v driver.Value) (string, error) {
	s, ok := v.(*sql.NullString)
	if !ok {
		return "", fmt.Errorf("unexpected input for FromValue: %T", v) // nolint:err113
	}

	if !s.Valid {
		return "", nil
	}

	return fmt.Sprintf("%s-%06s", p.prefix, s.String), nil
}

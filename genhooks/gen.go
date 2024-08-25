package genhooks

import (
	"embed"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc/gen"
	sliceutil "github.com/theopenlane/utils/slice"
)

var (
	//go:embed templates/*
	_templates embed.FS
)

var (
	softDeleteFields = []string{"deleted_at", "deleted_by"}
)

// isSoftDeleteField checks if the field is a soft delete field
func isSoftDeleteField(f *gen.Field) bool {
	return sliceutil.Contains(softDeleteFields, f.Name)
}

// getFileName returns the file name for the query file
func getFileName(dir, name string) string {
	return filepath.Clean(dir + strings.ToLower(name) + ".graphql")
}

// toFirstLower converts the first character of a string to lowercase
// except if the entire string is uppercase, e.g TTL, should remain as TTL
func toFirstLower(s string) string {
	if strings.ToUpper(s) == s {
		return s
	}

	return strings.ToLower(s[:1]) + s[1:]
}

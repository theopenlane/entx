package genhooks

import (
	"embed"
	"path/filepath"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/samber/lo"
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
	return lo.Contains(softDeleteFields, f.Name)
}

// getFileName returns the file name for the query file
func getFileName(dir, name string) string {
	return filepath.Clean(dir + strings.ToLower(name) + ".graphql")
}

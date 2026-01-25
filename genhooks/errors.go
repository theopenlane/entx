package genhooks

import "errors"

// ErrGraphResolverDirRequired is returned when UpdateWorkflowResolvers is called without a directory
var ErrGraphResolverDirRequired = errors.New("graphResolverDir is required")

// ErrModuleRootNotFound is returned when the module root cannot be determined from import paths or go.mod
var ErrModuleRootNotFound = errors.New("unable to determine module root")

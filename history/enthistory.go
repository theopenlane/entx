package history

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type ExtensionOption = func(*Extension)

const (
	defaultSchemaPath  = "./schema"
	defaultPackageName = "schema"
)

// UpdatedBy is a struct that holds the key and type for the updated_by field
type UpdatedBy struct {
	key       string
	valueType ValueType
	// Nillable indicates if the updated_by field should be nillable
	Nillable bool
}

// FieldProperties is a struct that holds the properties for the fields in the history schema
type FieldProperties struct {
	// Nillable indicates if the fields should be nillable
	Nillable bool
	// Immutable indicates if the fields should be immutable (not allowed to be updated)
	Immutable bool
}

// Config is the configuration for the history extension
type Config struct {
	// IncludeUpdatedBy optionally adds an updated_by field to the history schema
	IncludeUpdatedBy bool
	// UpdatedBy holds the key and type for the updated_by field
	UpdatedBy *UpdatedBy
	// Auditing enables the generation of the Audit() method on the client
	Auditing bool
	// QueryHelpers enables the generation the query helpers for the history schema
	QueryHelpers bool
	// InputSchemaPath is the path to the input schema directory, defaults to "./schema"
	InputSchemaPath string
	// OutputSchemaPath is the path to the output schema directory, defaults to "./schema"
	OutputSchemaPath string
	// SchemaName is an optional schema name to use instead of the default generated name
	SchemaName string
	// PackageName is an optional package name to use for the history schema
	PackageName string
	// Query is a boolean that tells the extension to add the entgql query annotations
	Query bool
	// Skipper is an optional function name to use as a skipper for history tracking
	Skipper string
	// FieldProperties holds the properties for the fields in the history schema
	FieldProperties *FieldProperties
	// HistoryTimeIndex tells the extension to add an index to the history_time field
	HistoryTimeIndex bool
	// Auth includes the authz policy settings
	Auth AuthzSettings
	// UsePondPool to create history updates in parallel
	UsePondPool bool
}

type AuthzSettings struct {
	// Enabled is a boolean that tells the extension to generate the authz policy
	Enabled bool
	// FirstRun is a boolean that tells the extension to only generate the policies after the first run
	FirstRun bool
	// AllowedRelation is the name of the relation that should be used to restrict
	// all audit log queries to users with that role, if not set the interceptor will not be added
	AllowedRelation string
}

// Name of the Config
func (c Config) Name() string {
	return "HistoryConfig"
}

// Extension implements entc.Extension
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new history extension
func New(opts ...ExtensionOption) *Extension {
	extension := &Extension{
		// Set configuration defaults that can get overridden with ExtensionOption
		config: &Config{
			InputSchemaPath:  defaultSchemaPath,
			OutputSchemaPath: defaultSchemaPath,
			Auditing:         false,
			PackageName:      defaultPackageName,
			FieldProperties:  &FieldProperties{},
		},
	}

	for _, opt := range opts {
		opt(extension)
	}

	return extension
}

// Templates returns the generated templates which include the client, history query, history from mutation
// and an optional auditing template
func (h *Extension) Templates() []*gen.Template {
	templates := []*gen.Template{
		parseTemplate("historyFromMutation", "templates/historyFromMutation.tmpl"),
		parseTemplate("historyClient", "templates/historyClient.tmpl"),
	}

	if h.config.QueryHelpers {
		templates = append(templates, parseTemplate("historyQuery", "templates/historyQuery.tmpl"))
	}

	if h.config.Auditing {
		templates = append(templates, parseTemplate("auditing", "templates/auditing.tmpl"))
	}

	return templates
}

// Annotations of the HistoryExtension
func (h *Extension) Annotations() []entc.Annotation {
	return []entc.Annotation{
		h.config,
	}
}

// SetFirstRun sets the first run value for the history extension outside of the options
func (h *Extension) SetFirstRun(firstRun bool) {
	h.config.Auth.FirstRun = firstRun
}

// WithAuditing allows you to turn on the code generation for the `.Audit()` method
func WithAuditing() ExtensionOption {
	return func(h *Extension) {
		h.config.Auditing = true
	}
}

// WithQueryHelpers generates the history query helpers for the history schema for pagination with Next(), Latest(), etc.
func WithQueryHelpers() ExtensionOption {
	return func(h *Extension) {
		h.config.QueryHelpers = true
	}
}

func WithAuthzPolicy() ExtensionOption {
	return func(h *Extension) {
		h.config.Auth.Enabled = true
	}
}

// WithGQLQuery adds the entgql Query annotation to the history schema in order to allow for querying
func WithGQLQuery() ExtensionOption {
	return func(h *Extension) {
		h.config.Query = true
	}
}

// WithHistoryTimeIndex allows you to add an index to the "history_time" fields
func WithHistoryTimeIndex() ExtensionOption {
	return func(h *Extension) {
		h.config.HistoryTimeIndex = true
	}
}

// WithImmutableFields allows you to set all tracked fields in history to Immutable
func WithImmutableFields() ExtensionOption {
	return func(h *Extension) {
		h.config.FieldProperties.Immutable = true
	}
}

// WithPackageName allows you to set an alternative package name for the history schema
func WithPackageName(packageName string) ExtensionOption {
	return func(h *Extension) {
		h.config.PackageName = packageName
	}
}

// WithNillableFields allows you to set all tracked fields in history to Nillable
// except enthistory managed fields (history_time, ref, operation, updated_by, & deleted_by)
func WithNillableFields() ExtensionOption {
	return func(h *Extension) {
		h.config.FieldProperties.Nillable = true
	}
}

// WithSchemaName allows you to set an alternative schema name
// This can be used to set a schema name for multi-schema migrations and SchemaConfig feature
// https://entgo.io/docs/multischema-migrations/
func WithSchemaName(schemaName string) ExtensionOption {
	return func(h *Extension) {
		h.config.SchemaName = schemaName
	}
}

// WithInputSchemaPath allows you to set an alternative schemaPath
// Defaults to "./schema"
func WithInputSchemaPath(schemaPath string) ExtensionOption {
	return func(h *Extension) {
		h.config.InputSchemaPath = schemaPath
	}
}

// WithOutputSchemaPath allows you to set an alternative schemaPath
// Defaults to "./schema"
func WithOutputSchemaPath(schemaPath string) ExtensionOption {
	return func(h *Extension) {
		h.config.OutputSchemaPath = schemaPath
	}
}

// WithFirstRun tells the extension to generate the history schema on the first run
// which leaves out the entfga policy
func WithFirstRun(firstRun bool) ExtensionOption {
	return func(h *Extension) {
		h.config.Auth.FirstRun = firstRun
	}
}

// WithUsePondPool allows you to use the pond pool to create history records in parallel
func WithUsePondPool() ExtensionOption {
	return func(h *Extension) {
		h.config.UsePondPool = true
	}
}

// WithAllowedRelation sets the relation that should be used to restrict all audit log queries to users with that role
func WithAllowedRelation(relation string) ExtensionOption {
	return func(h *Extension) {
		h.config.Auth.AllowedRelation = relation
	}
}

// WithSkipper allows you to set a skipper function to skip history tracking
func WithSkipper(skipper string) ExtensionOption {
	return func(h *Extension) {
		h.config.Skipper = skipper
	}
}

// WithUpdatedBy sets the key and type for pulling updated_by from the context,
// usually done via a middleware to track which users are making which changes
func WithUpdatedBy(key string, valueType ValueType) ExtensionOption {
	return func(h *Extension) {
		h.config.IncludeUpdatedBy = true
		h.config.UpdatedBy = &UpdatedBy{
			key:       key,
			valueType: valueType,
			Nillable:  true,
		}
	}
}

// WithUpdatedByFromSchema uses the original update_by value in the schema and includes in the audit results
func WithUpdatedByFromSchema(valueType ValueType, nillable bool) ExtensionOption {
	return func(h *Extension) {
		h.config.IncludeUpdatedBy = true
		h.config.UpdatedBy = &UpdatedBy{
			valueType: ValueTypeString,
			Nillable:  nillable,
		}
	}
}

package enthistory

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

type ExtensionOption = func(*HistoryExtension)

// UpdatedBy is a struct that holds the key and type for the updated_by field
type UpdatedBy struct {
	key       string
	valueType ValueType
	Nillable  bool
}

// FieldProperties is a struct that holds the properties for the fields in the history schema
type FieldProperties struct {
	Nillable  bool
	Immutable bool
}

// Config is the configuration for the history extension
type Config struct {
	IncludeUpdatedBy bool
	UpdatedBy        *UpdatedBy
	Auditing         bool
	SchemaPath       string
	SchemaName       string
	Query            bool
	Skipper          string
	FieldProperties  *FieldProperties
	HistoryTimeIndex bool
	Auth             AuthzSettings
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

// HistoryExtension implements entc.Extension.
type HistoryExtension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new history extension
func New(opts ...ExtensionOption) *HistoryExtension {
	extension := &HistoryExtension{
		// Set configuration defaults that can get overridden with ExtensionOption
		config: &Config{
			SchemaPath:      "./schema",
			Auditing:        false,
			FieldProperties: &FieldProperties{},
		},
	}

	for _, opt := range opts {
		opt(extension)
	}

	return extension
}

// Templates returns the generated templates which include the client, history query, history from mutation
// and an optional auditing template
func (h *HistoryExtension) Templates() []*gen.Template {
	templates := []*gen.Template{
		parseTemplate("historyFromMutation", "templates/historyFromMutation.tmpl"),
		parseTemplate("historyQuery", "templates/historyQuery.tmpl"),
		parseTemplate("historyClient", "templates/historyClient.tmpl"),
	}

	if h.config.Auditing {
		templates = append(templates, parseTemplate("auditing", "templates/auditing.tmpl"))
	}

	return templates
}

// Annotations of the HistoryExtension
func (h *HistoryExtension) Annotations() []entc.Annotation {
	return []entc.Annotation{
		h.config,
	}
}

// SetFirstRun sets the first run value for the history extension outside of the options
func (h *HistoryExtension) SetFirstRun(firstRun bool) {
	h.config.Auth.FirstRun = firstRun
}

// WithAuditing allows you to turn on the code generation for the `.Audit()` method
func WithAuditing() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Auditing = true
	}
}

func WithAuthzPolicy() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Auth.Enabled = true
	}
}

// WithGQLQuery adds the entgql Query annotation to the history schema in order to allow for querying
func WithGQLQuery() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Query = true
	}
}

// WithHistoryTimeIndex allows you to add an index to the "history_time" fields
func WithHistoryTimeIndex() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.HistoryTimeIndex = true
	}
}

// WithImmutableFields allows you to set all tracked fields in history to Immutable
func WithImmutableFields() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.FieldProperties.Immutable = true
	}
}

// WithNillableFields allows you to set all tracked fields in history to Nillable
// except enthistory managed fields (history_time, ref, operation, updated_by, & deleted_by)
func WithNillableFields() ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.FieldProperties.Nillable = true
	}
}

// WithSchemaName allows you to set an alternative schema name
// This can be used to set a schema name for multi-schema migrations and SchemaConfig feature
// https://entgo.io/docs/multischema-migrations/
func WithSchemaName(schemaName string) ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.SchemaName = schemaName
	}
}

// WithSchemaPath allows you to set an alternative schemaPath
// Defaults to "./schema"
func WithSchemaPath(schemaPath string) ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.SchemaPath = schemaPath
	}
}

// WithFirstRun tells the extension to generate the history schema on the first run
// which leaves out the entfga policy
func WithFirstRun(firstRun bool) ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Auth.FirstRun = firstRun
	}
}

// WithAllowedRelation sets the relation that should be used to restrict all audit log queries to users with that role
func WithAllowedRelation(relation string) ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Auth.AllowedRelation = relation
	}
}

// WithSkipper allows you to set a skipper function to skip history tracking
func WithSkipper(skipper string) ExtensionOption {
	return func(h *HistoryExtension) {
		h.config.Skipper = skipper
	}
}

// WithUpdatedBy sets the key and type for pulling updated_by from the context,
// usually done via a middleware to track which users are making which changes
func WithUpdatedBy(key string, valueType ValueType) ExtensionOption {
	return func(h *HistoryExtension) {
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
	return func(h *HistoryExtension) {
		h.config.IncludeUpdatedBy = true
		h.config.UpdatedBy = &UpdatedBy{
			valueType: ValueTypeString,
			Nillable:  nillable,
		}
	}
}

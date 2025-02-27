// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx/vanilla/_example/ent/organization"
)

// OrganizationCreate is the builder for creating a Organization entity.
type OrganizationCreate struct {
	config
	mutation *OrganizationMutation
	hooks    []Hook
}

// SetDisplayID sets the "display_id" field.
func (oc *OrganizationCreate) SetDisplayID(s string) *OrganizationCreate {
	oc.mutation.SetDisplayID(s)
	return oc
}

// SetCreatedAt sets the "created_at" field.
func (oc *OrganizationCreate) SetCreatedAt(t time.Time) *OrganizationCreate {
	oc.mutation.SetCreatedAt(t)
	return oc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableCreatedAt(t *time.Time) *OrganizationCreate {
	if t != nil {
		oc.SetCreatedAt(*t)
	}
	return oc
}

// SetUpdatedAt sets the "updated_at" field.
func (oc *OrganizationCreate) SetUpdatedAt(t time.Time) *OrganizationCreate {
	oc.mutation.SetUpdatedAt(t)
	return oc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableUpdatedAt(t *time.Time) *OrganizationCreate {
	if t != nil {
		oc.SetUpdatedAt(*t)
	}
	return oc
}

// SetCreatedBy sets the "created_by" field.
func (oc *OrganizationCreate) SetCreatedBy(s string) *OrganizationCreate {
	oc.mutation.SetCreatedBy(s)
	return oc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableCreatedBy(s *string) *OrganizationCreate {
	if s != nil {
		oc.SetCreatedBy(*s)
	}
	return oc
}

// SetUpdatedBy sets the "updated_by" field.
func (oc *OrganizationCreate) SetUpdatedBy(s string) *OrganizationCreate {
	oc.mutation.SetUpdatedBy(s)
	return oc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableUpdatedBy(s *string) *OrganizationCreate {
	if s != nil {
		oc.SetUpdatedBy(*s)
	}
	return oc
}

// SetName sets the "name" field.
func (oc *OrganizationCreate) SetName(s string) *OrganizationCreate {
	oc.mutation.SetName(s)
	return oc
}

// SetDescription sets the "description" field.
func (oc *OrganizationCreate) SetDescription(s string) *OrganizationCreate {
	oc.mutation.SetDescription(s)
	return oc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableDescription(s *string) *OrganizationCreate {
	if s != nil {
		oc.SetDescription(*s)
	}
	return oc
}

// SetID sets the "id" field.
func (oc *OrganizationCreate) SetID(s string) *OrganizationCreate {
	oc.mutation.SetID(s)
	return oc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (oc *OrganizationCreate) SetNillableID(s *string) *OrganizationCreate {
	if s != nil {
		oc.SetID(*s)
	}
	return oc
}

// Mutation returns the OrganizationMutation object of the builder.
func (oc *OrganizationCreate) Mutation() *OrganizationMutation {
	return oc.mutation
}

// Save creates the Organization in the database.
func (oc *OrganizationCreate) Save(ctx context.Context) (*Organization, error) {
	if err := oc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, oc.sqlSave, oc.mutation, oc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (oc *OrganizationCreate) SaveX(ctx context.Context) *Organization {
	v, err := oc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (oc *OrganizationCreate) Exec(ctx context.Context) error {
	_, err := oc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (oc *OrganizationCreate) ExecX(ctx context.Context) {
	if err := oc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (oc *OrganizationCreate) defaults() error {
	if _, ok := oc.mutation.CreatedAt(); !ok {
		if organization.DefaultCreatedAt == nil {
			return fmt.Errorf("ent: uninitialized organization.DefaultCreatedAt (forgotten import ent/runtime?)")
		}
		v := organization.DefaultCreatedAt()
		oc.mutation.SetCreatedAt(v)
	}
	if _, ok := oc.mutation.UpdatedAt(); !ok {
		if organization.DefaultUpdatedAt == nil {
			return fmt.Errorf("ent: uninitialized organization.DefaultUpdatedAt (forgotten import ent/runtime?)")
		}
		v := organization.DefaultUpdatedAt()
		oc.mutation.SetUpdatedAt(v)
	}
	if _, ok := oc.mutation.ID(); !ok {
		if organization.DefaultID == nil {
			return fmt.Errorf("ent: uninitialized organization.DefaultID (forgotten import ent/runtime?)")
		}
		v := organization.DefaultID()
		oc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (oc *OrganizationCreate) check() error {
	if _, ok := oc.mutation.DisplayID(); !ok {
		return &ValidationError{Name: "display_id", err: errors.New(`ent: missing required field "Organization.display_id"`)}
	}
	if v, ok := oc.mutation.DisplayID(); ok {
		if err := organization.DisplayIDValidator(v); err != nil {
			return &ValidationError{Name: "display_id", err: fmt.Errorf(`ent: validator failed for field "Organization.display_id": %w`, err)}
		}
	}
	if _, ok := oc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Organization.name"`)}
	}
	if v, ok := oc.mutation.Name(); ok {
		if err := organization.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Organization.name": %w`, err)}
		}
	}
	return nil
}

func (oc *OrganizationCreate) sqlSave(ctx context.Context) (*Organization, error) {
	if err := oc.check(); err != nil {
		return nil, err
	}
	_node, _spec := oc.createSpec()
	if err := sqlgraph.CreateNode(ctx, oc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Organization.ID type: %T", _spec.ID.Value)
		}
	}
	oc.mutation.id = &_node.ID
	oc.mutation.done = true
	return _node, nil
}

func (oc *OrganizationCreate) createSpec() (*Organization, *sqlgraph.CreateSpec) {
	var (
		_node = &Organization{config: oc.config}
		_spec = sqlgraph.NewCreateSpec(organization.Table, sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString))
	)
	if id, ok := oc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := oc.mutation.DisplayID(); ok {
		_spec.SetField(organization.FieldDisplayID, field.TypeString, value)
		_node.DisplayID = value
	}
	if value, ok := oc.mutation.CreatedAt(); ok {
		_spec.SetField(organization.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := oc.mutation.UpdatedAt(); ok {
		_spec.SetField(organization.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := oc.mutation.CreatedBy(); ok {
		_spec.SetField(organization.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := oc.mutation.UpdatedBy(); ok {
		_spec.SetField(organization.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := oc.mutation.Name(); ok {
		_spec.SetField(organization.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := oc.mutation.Description(); ok {
		_spec.SetField(organization.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	return _node, _spec
}

// OrganizationCreateBulk is the builder for creating many Organization entities in bulk.
type OrganizationCreateBulk struct {
	config
	err      error
	builders []*OrganizationCreate
}

// Save creates the Organization entities in the database.
func (ocb *OrganizationCreateBulk) Save(ctx context.Context) ([]*Organization, error) {
	if ocb.err != nil {
		return nil, ocb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ocb.builders))
	nodes := make([]*Organization, len(ocb.builders))
	mutators := make([]Mutator, len(ocb.builders))
	for i := range ocb.builders {
		func(i int, root context.Context) {
			builder := ocb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*OrganizationMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ocb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ocb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ocb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ocb *OrganizationCreateBulk) SaveX(ctx context.Context) []*Organization {
	v, err := ocb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ocb *OrganizationCreateBulk) Exec(ctx context.Context) error {
	_, err := ocb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ocb *OrganizationCreateBulk) ExecX(ctx context.Context) {
	if err := ocb.Exec(ctx); err != nil {
		panic(err)
	}
}
